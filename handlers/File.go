package handlers

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"gorm.io/gorm"
)

// FileHandler handler for updating files
func FileHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request libdm.FileRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	namespace, action, err := validateFileActionRequest(r, w, &handlerData, request)
	if err != nil {
		return err
	}

	// Find files
	files, err := models.FindFiles(handlerData.Db, handlerData.Config, models.File{
		Model: gorm.Model{
			ID: request.FileID,
		},
		Name:      request.Name,
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	// Apply group filter
	if len(request.Attributes.Groups) > 0 {
		var newFiles []models.File
		for i := range files {
			if files[i].IsInGroupList(request.Attributes.Groups) {
				newFiles = append(newFiles, files[i])
			}
		}

		files = newFiles
	}

	// Apply tag filter
	if len(request.Attributes.Tags) > 0 {
		var newFiles []models.File
		for i := range files {
			if files[i].IsInTagList(request.Attributes.Tags) {
				newFiles = append(newFiles, files[i])
			}
		}

		files = newFiles
	}

	// Exit if no file was found
	if len(files) == 0 {
		return RErrNotFound
	}

	// Check if files are more than requested
	if len(files) > 1 && !request.All {
		return NewRequestError("found multiple files with same name", http.StatusConflict)
	}

	// If namespace was not set, use the namespace of the returned file
	if namespace == nil {
		namespace = files[0].Namespace
		if !handlerData.User.HasAccess(namespace) {
			return RErrPermissionDenied.Append("for this namespace")
		}
	}

	// Execute action
	switch action {
	case "delete":
		{
			ids := make([]uint, len(files))

			for i, file := range files {
				// Delete each file
				err = file.Delete(handlerData.Db, handlerData.Config)
				if err != nil {
					return err
				}

				ids[i] = file.ID
			}

			// Send response
			sendResponse(w, libdm.ResponseSuccess, "", libdm.IDsResponse{
				IDs: ids,
			})
		}
	case "update":
		{
			var count uint32
			var didUpdate bool

			for _, file := range files {
				didUpdate, err = updateFile(&file, handlerData, request.Updates)
				if err != nil {
					return err
				}

				// Only count if update
				// was applied
				if didUpdate {
					count++
				}
			}

			// Send response
			sendResponse(w, libdm.ResponseSuccess, "", libdm.CountResponse{
				Count: count,
			})
		}
	// Download file
	case "get":
		{
			// Use first file
			err := serveFile(files[0], w, handlerData)
			if err != nil {
				return err
			}
		}
	case "publish":
		{
			resp, err := publishFiles(files, request.PublicName, request.All, handlerData.Db)
			if err != nil {
				return err
			}

			sendResponse(w, libdm.ResponseSuccess, "", resp)
		}
	}

	return nil
}

// Validate FileRequest
func validateFileActionRequest(r *http.Request, w http.ResponseWriter, handlerData *web.HandlerData, request libdm.FileRequest) (*models.Namespace, string, error) {
	// Validate input
	if len(request.Name) == 0 && request.FileID <= 0 {
		return nil, "", RErrBadRequest
	}

	// Get action
	vars := mux.Vars(r)
	action, has := vars["action"]
	if !has {
		return nil, "", RErrBadRequest
	}

	// Getting all files is not allowed
	if request.All && action == "get" {
		return nil, "", RErrBadRequest
	}

	var namespace *models.Namespace

	// Use given namespace if fileID is not set
	if request.FileID == 0 {
		// Select namespace
		namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)

		// Handle namespace errors (not found || no access)
		if !handleNamespaceErorrs(namespace, handlerData.User, w) {
			return nil, "", nil
		}
	}

	// Check if action is valid
	if !gaw.IsInStringArray(action, []string{"delete", "update", "get", "publish"}) {
		return nil, "", RErrInvalid.Append("action")
	}

	return namespace, action, nil
}

// Serve file contents for client
func serveFile(file models.File, w http.ResponseWriter, handlerData web.HandlerData) error {
	// Open local file
	f, err := os.Open(handlerData.Config.GetStorageFile(file.LocalName))
	if LogError(err) {
		if os.IsNotExist(err) {
			return RErrNotFound.Prepend("File").Append("on server")
		}

		return err
	}

	// Set required headers
	if len(file.FileType) > 0 && filetype.IsMIMESupported(file.FileType) {
		w.Header().Set(libdm.HeaderContentType, file.FileType)
	}

	w.Header().Set(libdm.HeaderFileName, file.Name)
	w.Header().Set(libdm.HeaderChecksum, file.Checksum)
	w.Header().Set(libdm.HeaderFileID, strconv.FormatUint(uint64(file.ID), 10))
	w.Header().Set(libdm.HeaderContentLength, strconv.FormatInt(file.FileSize, 10))

	if file.Encryption.Valid {
		w.Header().Set(libdm.HeaderEncryption, libdm.ChiperToString(file.Encryption.Int32))
	}

	// Write contents to responsewriter
	buff := make([]byte, 1024*1024)
	_, err = io.CopyBuffer(w, f, buff)
	if err != nil {
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			return nil
		}

		return err
	}

	// Close file
	LogError(f.Close())
	return nil
}

// Publish multiple files
func publishFiles(files []models.File, publicName string, all bool, db *gorm.DB) (interface{}, error) {
	bulkPublishResponse := libdm.BulkPublishResponse{}

	for _, file := range files {
		// Ignore if already public
		if file.IsPublic {
			// Send error if publishing only one file
			if len(files) == 1 {
				return nil, NewRequestError("Already public", http.StatusConflict)
			}

			continue
		}

		nameTaken, err := file.Publish(db, publicName)
		if err != nil {
			return nil, err
		}

		if nameTaken {
			return nil, RErrAlreadyExists.Prepend("Public name")
		}

		if all && len(files) > 1 {
			bulkPublishResponse.Files = append(bulkPublishResponse.Files, libdm.UploadResponse{
				FileID:         file.ID,
				Filename:       file.Name,
				PublicFilename: file.PublicFilename.String,
			})
		} else {
			return libdm.PublishResponse{
				PublicFilename: file.PublicFilename.String,
			}, nil
		}
	}

	return bulkPublishResponse, nil
}

// Apply all given updates to a file
func updateFile(file *models.File, handlerData web.HandlerData, update libdm.FileUpdateItem) (didUpdate bool, err error) {
	// Update namespace
	if len(update.NewNamespace) > 0 {
		// Get new namespace
		newNamespace := models.FindNamespace(handlerData.Db, update.NewNamespace, handlerData.User)
		if newNamespace == nil || file.Namespace.ID == 0 {
			err = RErrNotFound.Prepend("New namespace")
			return
		}

		// Check if user can access this new namespace
		if !newNamespace.IsOwnedBy(handlerData.User) && !handlerData.User.CanWriteForeignNamespace() {
			err = RErrPermissionDenied.Append("for this namespace")
			return
		}

		// Update files namespace
		err = file.UpdateNamespace(handlerData.Db, newNamespace, handlerData.User)
		if err != nil {
			return
		}

		didUpdate = true
	}

	// Rename file
	if len(update.NewName) > 0 {
		if err = file.Rename(handlerData.Db, update.NewName); err != nil {
			return
		}

		didUpdate = true
	}

	// Set public/private
	if len(update.IsPublic) > 0 {
		if !file.PublicFilename.Valid {
			err = NewRequestError("You need to share this file first", http.StatusBadRequest)
			return
		}

		var newVisibility bool
		newVisibility, err = strconv.ParseBool(update.IsPublic)
		if err != nil {
			err = NewRequestError("isPublic must be a bool", http.StatusUnprocessableEntity)
			return
		}

		if err = file.SetVilibility(handlerData.Db, newVisibility); err != nil {
			return
		}

		didUpdate = true
	}

	// Add tags
	if len(update.AddTags) > 0 {
		currLenTags := len(file.Tags)
		if err = file.AddTags(handlerData.Db, update.AddTags, handlerData.User); err != nil {
			return
		}

		didUpdate = len(file.Tags) > currLenTags
	}

	// Remove tags
	if len(update.RemoveTags) > 0 {
		currLenTags := len(file.Tags)
		if err = file.RemoveTags(handlerData.Db, update.RemoveTags); err != nil {
			return
		}

		didUpdate = len(file.Tags) < currLenTags
	}

	// Add Groups
	if len(update.AddGroups) > 0 {
		currLenGroups := len(file.Groups)
		if err = file.AddGroups(handlerData.Db, update.AddGroups, handlerData.User); err != nil {
			return
		}

		didUpdate = len(file.Groups) > currLenGroups
	}

	// Remove Groups
	if len(update.RemoveGroups) > 0 {
		currLenGroups := len(file.Groups)
		if err = file.RemoveGroups(handlerData.Db, update.RemoveGroups); err != nil {
			return
		}

		didUpdate = len(file.Groups) < currLenGroups
	}

	return
}
