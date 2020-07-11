package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/JojiiOfficial/gaw"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

//UploadfileHandler handler for uploading files
func UploadfileHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request models.UploadRequest

	// Get data from header
	requestData := r.Header.Get(models.HeaderRequest)
	if len(requestData) == 0 {
		return RErrBadRequest
	}

	// Decode header base64
	rBaseBytes, err := base64.StdEncoding.DecodeString(requestData)
	if err != nil {
		return RErrBadRequest
	}

	// Parse json from request header
	err = json.Unmarshal(rBaseBytes, &request)
	if err != nil {
		return RErrBadRequest
	}

	// Check requested encryption type
	if len(request.Encryption) > 0 && !libdm.IsValidCipher(request.Encryption) {
		return RErrNotSupported.Prepend("Encryption")
	}

	// Validating request, for desired upload Type
	switch request.UploadType {
	case models.FileUploadType:
		{
			// Check if user is allowed to upload files
			if !handlerData.User.CanUploadFiles() {
				return RErrNotAllowed.Append("to upload files")
			}
		}
	case models.URLUploadType:
		{
			// Check if user is allowed to upload URLs
			if !handlerData.User.AllowedToUploadURLs() {
				return RErrNotAllowed.Append("to upload URLs")
			}

			// Check if url is set and valid
			if len(request.URL) == 0 || !isValidHTTPURL(request.URL) {
				return RErrMissing.Append("or malformed URL")
			}
		}
	default:
		{
			// Send error if UploadType was not found
			return RErrInvalid.Append("upload type")
		}
	}

	var namespace *models.Namespace
	var file *models.File
	var replaceMode bool

	if request.ReplaceFile > 0 {
		replaceMode = true

		// Find file
		file, err = models.FindFile(handlerData.Db, request.ReplaceFile, handlerData.User.ID)
		if err != nil || file == nil || file.Namespace == nil {
			return RErrNotFound.Prepend("File")
		}

		// Use new name if set
		if len(request.Name) > 0 {
			file.Name = request.Name
		}

		// Select namespace
		if len(request.Attributes.Namespace) > 0 {
			// Find namespace specified by client
			namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)
			file.Namespace = namespace
		} else {
			namespace = file.Namespace
		}
	} else {
		if len(request.Name) == 0 {
			request.Name = gaw.RandString(25)
		}

		// Select namespace
		namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)

		// Generate file
		file = &models.File{
			Namespace: namespace,
			Name:      request.Name,
		}

		// Pick a random, not already used name for file
		if !file.SetUniqueFilename(handlerData.Db) {
			sendServerError(w)
			return nil
		}
	}

	// Check if namespace is valid and user has access to it
	if !handleNamespaceErorrs(namespace, handlerData.User, w) {
		return nil
	}

	// Set Tags, Groups and encryption
	if len(request.Attributes.Tags) > 0 {
		file.Tags = models.TagsFromStringArr(request.Attributes.Tags, *namespace, handlerData.User)
	}
	if len(request.Attributes.Groups) > 0 {
		file.Groups = models.GroupsFromStringArr(request.Attributes.Groups, *namespace, handlerData.User)
	}
	file.SetEncryption(request.Encryption)

	if request.Public {
		// Determine public name
		publicName := request.PublicName
		if len(publicName) == 0 {
			publicName = gaw.RandString(25)
		}

		// Set file public name
		file.PublicFilename = sql.NullString{
			String: publicName,
			Valid:  true,
		}
		file.IsPublic = true

		// Check if public name already exists
		_, found, _ := models.GetPublicFile(handlerData.Db, publicName)
		if found {
			return RErrAlreadyExists.Prepend("public name")
		}
	}

	// Create local file
	localFile := handlerData.Config.GetStorageFile(file.LocalName)
	f, err := os.Create(localFile)
	if err != nil {
		return err
	}

	// Read from the desired source (file/url)
	switch request.UploadType {
	case models.FileUploadType:
		{
			// Read requestd file
			size, checksum, err := readMultipartToFile(f, r.Body, w)

			// Close file and log error only
			LogError(f.Close())

			// success is false if the calculated
			// and provided hash are not equal
			if err != nil {
				// Only shredder file if not in replace mode
				if request.ReplaceFile == 0 {
					go models.ShredderFile(localFile, -1)
				}

				// If error is a timeout error, send timeout error and close connectio
				if err == http.ErrHandlerTimeout {
					err = RErrTimeout
				}

				return err
			}

			file.FileSize = size
			file.Checksum = checksum
		}
	case models.URLUploadType:
		{
			// Read from HTTP request
			status, err := downloadHTTP(handlerData.User, request.URL, f, file)
			if err != nil {
				return RErrBadRequest.Prepend(err.Error())
			}

			// Check statuscode
			if status > 299 || status < 200 {
				return NewRequestError("Non HTTP OK response: "+strconv.Itoa(status), http.StatusBadRequest)
			}
		}
	}

	// Detect mime type
	mime, err := mimetype.DetectFile(handlerData.Config.GetStorageFile(file.LocalName))
	if err != nil {
		log.Info("Can't detect mime: ", err.Error())
	} else {
		file.FileType = strings.Split(mime.String(), ";")[0]
	}

	if replaceMode {
		// Update file
		err = file.Save(handlerData.Db)
	} else {
		// Insert file to DB
		err = file.Insert(handlerData.Db, handlerData.User)
	}

	if err != nil {
		return err
	}

	sendResponse(w, models.ResponseSuccess, "", models.UploadResponse{
		FileID:         file.ID,
		Filename:       file.Name,
		PublicFilename: file.PublicFilename.String,
		Checksum:       file.Checksum,
		FileSize:       file.FileSize,
		Namespace:      namespace.Name,
	})

	return nil
}

// ListFilesHandler handler for listing files
func ListFilesHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request models.FileListRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	var namespace *models.Namespace

	if !request.AllNamespaces {
		// Select namespace
		namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)

		// Handle namespace errors (not found || no access)
		if !handleNamespaceErorrs(namespace, handlerData.User, w) {
			return nil
		}
	}

	var foundFiles []models.File

	loaded := handlerData.Db.Model(&foundFiles)
	if len(request.Attributes.Tags) > 0 || request.OptionalParams.Verbose > 1 {
		loaded = loaded.Preload("Tags")
	}

	if len(request.Attributes.Groups) > 0 || request.OptionalParams.Verbose > 1 {
		loaded = loaded.Preload("Groups")
	}

	if request.OptionalParams.Verbose > 1 || request.AllNamespaces {
		loaded = loaded.Preload("Namespace")
	}

	if len(request.Name) > 0 {
		loaded = loaded.Where("files.name LIKE ?", "%"+request.Name+"%")
	}

	if request.AllNamespaces {
		// Join to filter by namespace creator
		loaded = loaded.
			Joins("INNER JOIN namespaces ON namespaces.id = files.namespace_id").
			Where("namespaces.creator = ?", handlerData.User.ID)
	} else {
		// Just select the specified namespace
		loaded = loaded.Where("namespace_id = ?", namespace.ID)
	}

	// Search
	err := loaded.Find(&foundFiles).Error
	if err != nil {
		return err
	}

	// Convert to ResponseFile
	var retFiles []models.FileResponseItem
	for _, file := range foundFiles {
		// Filter tags
		if (len(request.Attributes.Tags) == 0 || (len(request.Attributes.Tags) > 0 && file.IsInTagList(request.Attributes.Tags))) &&
			// Filter groups
			(len(request.Attributes.Groups) == 0 || (len(request.Attributes.Groups) > 0 && file.IsInGroupList(request.Attributes.Groups))) {
			respItem := models.FileResponseItem{
				ID:           file.ID,
				Name:         file.Name,
				CreationDate: file.CreatedAt,
				Size:         file.FileSize,
				IsPublic:     file.IsPublic,
				Checksum:     file.Checksum,
			}

			// Set encryption
			if file.Encryption.Valid && libdm.EncryptionIValid(file.Encryption.Int32) {
				respItem.Encryption = libdm.ChiperToString(file.Encryption.Int32)
			}

			// Append public name if available
			if file.PublicFilename.Valid && len(file.PublicFilename.String) > 0 {
				respItem.PublicName = file.PublicFilename.String
			}

			// Return attributes on verbose
			if request.OptionalParams.Verbose > 1 || request.AllNamespaces {
				respItem.Attributes = file.GetAttributes()
			}

			// Add if matching filter
			retFiles = append(retFiles, respItem)
		}
	}

	sendResponse(w, models.ResponseSuccess, "", models.ListFileResponse{
		Files: retFiles,
	})

	return nil
}

// FileHandler handler for updating files
func FileHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request models.FileRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	// Validate input
	if len(request.Name) == 0 && request.FileID <= 0 {
		return RErrBadRequest
	}

	// Get action
	vars := mux.Vars(r)
	action, has := vars["action"]
	if !has {
		return RErrBadRequest
	}

	// Getting all files is not allowed
	if request.All && action == "get" {
		return RErrBadRequest
	}

	var namespace *models.Namespace

	// Use given namespace if fileID is not set
	if request.FileID == 0 {
		// Select namespace
		namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)

		// Handle namespace errors (not found || no access)
		if !handleNamespaceErorrs(namespace, handlerData.User, w) {
			return nil
		}
	}

	// Check if action is valid
	if !gaw.IsInStringArray(action, []string{"delete", "update", "get", "publish"}) {
		return RErrInvalid.Append("action")
	}

	// Find files
	files, err := models.FindFiles(handlerData.Db, handlerData.Config, models.File{
		Model: gorm.Model{
			ID: request.FileID,
		},
		Name:      request.Name,
		Namespace: namespace,
		// TODO add group/tag filter
	})

	if err != nil {
		return err
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

	// Determine if an update was applied
	var didUpdate bool

	// Execute action
	switch action {
	case "delete":
		{
			var ids []uint
			for _, file := range files {
				// Delete each file
				err = file.Delete(handlerData.Db, handlerData.Config)
				if err != nil {
					return err
				}
				ids = append(ids, file.ID)
			}

			// Send response
			sendResponse(w, models.ResponseSuccess, "", models.IDsResponse{
				IDs: ids,
			})
		}
	case "update":
		{
			var count uint32

			// Do for every file
			for _, file := range files {
				update := request.Updates

				// Update namespace
				if len(update.NewNamespace) > 0 {
					// Get new namespace
					newNamespace := models.FindNamespace(handlerData.Db, update.NewNamespace, handlerData.User)
					if newNamespace == nil || namespace.ID == 0 {
						return RErrNotFound.Prepend("New namespace")
					}

					// Check if user can access this new namespace
					if !newNamespace.IsOwnedBy(handlerData.User) && !handlerData.User.CanWriteForeignNamespace() {
						return RErrPermissionDenied.Append("for this namespace")
					}

					// Update files namespace
					err := file.UpdateNamespace(handlerData.Db, newNamespace, handlerData.User)
					if err != nil {
						return err
					}

					didUpdate = true
				}

				// Rename file
				if len(update.NewName) > 0 {
					if err = file.Rename(handlerData.Db, update.NewName); err != nil {
						return err
					}

					didUpdate = true
				}

				// Set public/private
				if len(update.IsPublic) > 0 {
					if !file.PublicFilename.Valid {
						return NewRequestError("You need to share this file first", http.StatusBadRequest)
					}

					newVisibility, err := strconv.ParseBool(update.IsPublic)
					if err != nil {
						return NewRequestError("isPublic must be a bool", http.StatusUnprocessableEntity)
					}

					if err = file.SetVilibility(handlerData.Db, newVisibility); err != nil {
						return err
					}
					didUpdate = true
				}

				// Add tags
				if len(update.AddTags) > 0 {
					currLenTags := len(file.Tags)
					if err = file.AddTags(handlerData.Db, update.AddTags, handlerData.User); err != nil {
						return err
					}

					didUpdate = len(file.Tags) > currLenTags
				}

				// Remove tags
				if len(update.RemoveTags) > 0 {
					currLenTags := len(file.Tags)
					if err = file.RemoveTags(handlerData.Db, update.RemoveTags); err != nil {
						return err
					}

					didUpdate = len(file.Tags) < currLenTags
				}

				// Add Groups
				if len(update.AddGroups) > 0 {
					currLenGroups := len(file.Groups)
					if err = file.AddGroups(handlerData.Db, update.AddGroups, handlerData.User); err != nil {
						return err
					}

					didUpdate = len(file.Groups) > currLenGroups
				}

				// Remove Groups
				if len(update.RemoveGroups) > 0 {
					currLenGroups := len(file.Groups)
					if err = file.RemoveGroups(handlerData.Db, update.RemoveGroups); err != nil {
						return err
					}

					didUpdate = len(file.Groups) < currLenGroups
				}

				// Only count if updated
				if didUpdate {
					count++
				}
			}

			// Send response
			sendResponse(w, models.ResponseSuccess, "", models.CountResponse{
				Count: count,
			})
		}
	// Get file
	case "get":
		{
			// Use first file
			file := files[0]

			// Open local file
			f, err := os.Open(handlerData.Config.GetStorageFile(file.LocalName))
			if LogError(err) {
				if os.IsNotExist(err) {
					return RErrNotFound.Prepend("File").Append("on server")
				}

				return err
			}

			// Set ContentType header
			if len(file.FileType) > 0 && filetype.IsMIMESupported(file.FileType) {
				w.Header().Set(models.HeaderContentType, file.FileType)
			}

			// Set filename header
			w.Header().Set(models.HeaderFileName, file.Name)

			// Set checksum header
			w.Header().Set(models.HeaderChecksum, file.Checksum)

			// Set fileID header
			w.Header().Set(models.HeaderFileID, strconv.FormatUint(uint64(file.ID), 10))

			// Set ContentLength header
			w.Header().Set(models.HeaderContentLength, strconv.FormatInt(file.FileSize, 10))

			// Set encryption cipher header
			if file.Encryption.Valid {
				w.Header().Set(models.HeaderEncryption, libdm.ChiperToString(file.Encryption.Int32))
			}

			// Write contents to responsewriter
			buff := make([]byte, 10*1024)
			_, err = io.CopyBuffer(w, f, buff)
			if err != nil {
				return err
			}

			// Close file
			LogError(f.Close())
		}
	// Publish a file
	case "publish":
		{
			publishResponse := models.PublishResponse{}
			bulkPublishResponse := models.BulkPublishResponse{}

			for _, file := range files {
				// Ignore if already public
				if file.IsPublic {
					// Send error if publishing only one file
					if len(files) == 1 {
						return NewRequestError("Already public", http.StatusConflict)
					}

					continue
				}

				nameTaken, err := file.Publish(handlerData.Db, request.PublicName)
				if err != nil {
					return err
				}

				if nameTaken {
					return RErrAlreadyExists.Prepend("Public name")
				}

				// Use bulk response if requested "all"
				if request.All {
					bulkPublishResponse.Files = append(bulkPublishResponse.Files, models.UploadResponse{
						FileID:         file.ID,
						Filename:       file.Name,
						PublicFilename: file.PublicFilename.String,
					})
				} else {
					// Otherwise respond with a single item
					publishResponse = models.PublishResponse{
						PublicFilename: file.PublicFilename.String,
					}
				}
			}

			// Send success
			if request.All {
				sendResponse(w, models.ResponseSuccess, "", bulkPublishResponse)
			} else {
				sendResponse(w, models.ResponseSuccess, "", publishResponse)
			}
		}
	}

	return nil
}

// Stats for user
func Stats(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	return nil
}
