package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/JojiiOfficial/gaw"
	"github.com/gabriel-vasile/mimetype"
	log "github.com/sirupsen/logrus"
)

//UploadfileHandler handler for uploading files
func UploadfileHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	request, err := parseUploadRequest(r)
	if err != nil {
		return err
	}

	err = validateUploadRequest(handlerData.User, request)
	if err != nil {
		return err
	}

	var namespace *models.Namespace
	var file *models.File
	var needNewFile = request.ReplaceFileByID == 0

	// Replace with same name
	if request.ReplaceEqualNames {
		namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)
		if !handleNamespaceErorrs(namespace, handlerData.User, w) {
			return nil
		}

		// We don't need errors since it should only
		// replace files if some were found
		files, _ := models.FilesByName(handlerData.Db, handlerData.User.ID, namespace.ID, request.Name)
		if files != nil && len(files) > 0 {
			if len(files) > 1 && !request.All {
				return NewRequestError("found multiple files with same name", http.StatusConflict)
			}

			// We want to have only one
			// file with this name anymore
			for i := range files {
				err := files[i].Delete(handlerData.Db, handlerData.Config)
				if err != nil {
					return err
				}
			}

		}
	}

	// Replace by ID
	if request.ReplaceFileByID > 0 {
		file, err = models.FindFileByID(handlerData.Db, request.ReplaceFileByID, handlerData.User.ID)
		if err != nil || file == nil || file.Namespace == nil {
			return RErrNotFound.Prepend("File")
		}

		// Use new name if set
		if len(request.Name) > 0 {
			file.Name = request.Name
		}

		if len(request.Attributes.Namespace) > 0 {
			// Switch to by client specified namespace
			namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)
			file.Namespace = namespace
		} else {
			// Keep namespace
			namespace = file.Namespace
		}
	}

	// Create new file
	if needNewFile {
		// Use random filename if none was set
		if len(request.Name) == 0 {
			request.Name = gaw.RandString(25)
		}

		if namespace == nil {
			namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)
		}

		file = &models.File{
			Name:      request.Name,
			User:      handlerData.User,
			Namespace: namespace,
		}

		// Pick a random, not already used name for local file
		if !file.SetUniqueFilename(handlerData.Db) {
			sendServerError(w)
			return nil
		}
	}

	// Check if namespace is valid and user has access to it
	if !handleNamespaceErorrs(namespace, handlerData.User, w) {
		return nil
	}

	file.ApplyAttributes(request.Attributes.Groups, request.Attributes.Tags)
	file.SetEncryption(request.Encryption)

	// Publish file
	if request.Public && file.MakePublic(handlerData.Db, request.PublicName) {
		return RErrAlreadyExists.Prepend("public name")
	}

	// Create local file
	localFile := handlerData.Config.GetStorageFile(file.LocalName)
	f, err := os.Create(localFile)
	if err != nil {
		return err
	}

	// Read from the desired source (file/url)
	switch request.UploadType {
	case libdm.FileUploadType:
		{
			// Read requestd file
			size, checksum, err := readMultipartToFile(f, r.Body, w)

			// Close file and log error only
			LogError(f.Close())

			// success is false if the calculated
			// and provided hash are not equal
			if err != nil {
				// Only shredder file if not in replace mode
				if request.ReplaceFileByID == 0 {
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
	case libdm.URLUploadType:
		{
			// TODO improve

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

	if needNewFile {
		// Insert file to DB
		err = file.Insert(handlerData.Db, handlerData.User)
	} else {
		// Update file
		err = file.Save(handlerData.Db)
	}

	if err != nil {
		return err
	}

	sendResponse(w, libdm.ResponseSuccess, "", libdm.UploadResponse{
		FileID:         file.ID,
		Filename:       file.Name,
		PublicFilename: file.PublicFilename.String,
		Checksum:       file.Checksum,
		FileSize:       file.FileSize,
		Namespace:      namespace.Name,
	})

	return nil
}

func parseUploadRequest(r *http.Request) (*libdm.UploadRequestStruct, error) {
	var request libdm.UploadRequestStruct

	// Get data from header
	requestData := r.Header.Get(libdm.HeaderRequest)
	if len(requestData) == 0 {
		return nil, RErrBadRequest
	}

	// Decode header base64
	rBaseBytes, err := base64.StdEncoding.DecodeString(requestData)
	if err != nil {
		return nil, RErrBadRequest
	}

	// Parse json from request header
	err = json.Unmarshal(rBaseBytes, &request)
	if err != nil {
		return nil, RErrBadRequest
	}

	return &request, nil
}

func validateUploadRequest(user *models.User, request *libdm.UploadRequestStruct) error {
	// Check requested encryption type
	if len(request.Encryption) > 0 && !libdm.IsValidCipher(request.Encryption) {
		return RErrNotSupported.Prepend("Encryption")
	}

	// Check if both replace modes were applied
	if request.ReplaceEqualNames && request.ReplaceFileByID > 0 {
		return RErrNotAllowed
	}

	// Validating request, for desired upload Type
	switch request.UploadType {
	case libdm.FileUploadType:
		{
			// Check if user is allowed to upload files
			if !user.CanUploadFiles() {
				return RErrNotAllowed.Append("to upload files")
			}
		}
	case libdm.URLUploadType:
		{
			// Check if user is allowed to upload URLs
			if !user.AllowedToUploadURLs() {
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

	return nil
}
