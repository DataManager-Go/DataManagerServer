package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"
)

//UploadfileHandler handler for uploading files
func UploadfileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.UploadRequest

	//Get data from header
	requestData := r.Header.Get(models.HeaderRequest)
	if len(requestData) == 0 {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	//Decode header base64
	rBaseBytes, err := base64.StdEncoding.DecodeString(requestData)
	if err != nil {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	//Parse json from request header
	err = json.Unmarshal(rBaseBytes, &request)
	if LogError(err) {
		fmt.Println("Invalid Json:", err)
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	// Validating request, for desired upload Type
	switch request.UploadType {
	case models.FileUploadType:
		{
			//Check if user is allowed to upload files
			if !handlerData.user.CanUploadFiles() {
				sendResponse(w, models.ResponseError, "not allowed to upload files", nil, http.StatusForbidden)
				return
			}
		}
	case models.URLUploadType:
		{
			//Check if user is allowed to upload URLs
			if !handlerData.user.AllowedToUploadURLs() {
				sendResponse(w, models.ResponseError, "not allowed to upload urls", nil, http.StatusForbidden)
				return
			}

			//Check if url is set and valid
			if len(request.URL) == 0 || !isValidHTTPURL(request.URL) {
				sendResponse(w, models.ResponseError, "missing or malformed url", nil, http.StatusUnprocessableEntity)
				return
			}
		}
	default:
		{
			//Send error if UploadType was not found
			sendResponse(w, models.ResponseError, "invalid upload type", nil, http.StatusUnprocessableEntity)
			return
		}
	}

	//Set random name if not specified
	if len(request.Name) == 0 {
		request.Name = gaw.RandString(20)
	}

	//Select namespace
	namespace := models.FindNamespace(handlerData.db, request.Attributes.Namespace, handlerData.user)
	if namespace == nil {
		sendResponse(w, models.ResponseError, "namespace not found", nil, http.StatusNotFound)
		return
	}

	//Check if user can access this namespace
	if !namespace.IsOwnedBy(handlerData.user) && !handlerData.user.CanWriteForeignNamespace() {
		sendResponse(w, models.ResponseError, "Write permission denied for foreign namespaces", nil, http.StatusForbidden)
		return
	}

	//Get Tags
	tags := models.TagsFromStringArr(request.Attributes.Tags, *namespace, handlerData.user)

	//Get Groups
	groups := models.GroupsFromStringArr(request.Attributes.Groups, *namespace, handlerData.user)

	//Ensure localname is not already in use
	uniqueNameFound := false
	var localName string
	for i := 0; i < 5; i++ {
		localName = gaw.RandString(40)
		var c int
		handlerData.db.Model(&models.File{}).Where(&models.File{LocalName: localName}).Count(&c)
		if c == 0 {
			uniqueNameFound = true
			break
		}

		log.Warn("Name collision found. Trying again (%d/%d)", i, 5)
	}

	if !uniqueNameFound {
		sendServerError(w)
		return
	}

	//Generate file
	file := models.File{
		Groups:    groups,
		Tags:      tags,
		Namespace: namespace,
		Name:      request.Name,
	}

	if request.Public {
		//Determine public name
		publicName := request.PublicName
		if len(publicName) == 0 {
			publicName = gaw.RandString(25)
		}

		//Set file public name
		file.PublicFilename = sql.NullString{
			String: publicName,
			Valid:  true,
		}
		file.IsPublic = true

		//Check if public name already exists
		_, found, _ := models.GetPublicFile(handlerData.db, publicName)
		if found {
			sendResponse(w, models.ResponseError, "public name already exists", nil)
			return
		}
	}

	//set local name
	file.LocalName = localName

	//Create local file
	f, err := os.Create(handlerData.config.GetStorageFile(localName))
	if LogError(err) {
		sendServerError(w)
		return
	}

	//Read from the desired source (file/url)
	switch request.UploadType {
	case models.FileUploadType:
		//Read from uploaded file
		r.ParseMultipartForm(handlerData.user.Role.MaxUploadFileSize)

		uploadfile, _, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer uploadfile.Close()

		//Copy stream to file
		size, err := io.Copy(f, uploadfile)
		if LogError(err) {
			sendServerError(w)
			return
		}

		//Set filesize to written bytes
		file.FileSize = int64(size)

		if len(request.FileType) > 0 && filetype.IsMIMESupported(request.FileType) {
			file.FileType = request.FileType
		}
	case models.URLUploadType:
		//Read from HTTP request
		status, err := downloadHTTP(handlerData.user, request.URL, f, &file)
		if err != nil {
			sendResponse(w, models.ResponseError, err.Error(), nil, http.StatusBadRequest)
			return
		}

		//Check statuscode
		if status > 299 || status < 200 {
			sendResponse(w, models.ResponseError, "Non ok response: "+strconv.Itoa(status), nil, http.StatusBadRequest)
			return
		}
	}

	//Close file
	if LogError(f.Close()) {
		sendServerError(w)
		return
	}

	//Save file to DB
	err = file.Insert(handlerData.db, handlerData.user)

	if !LogError(err) {
		sendResponse(w, models.ResponseSuccess, "", models.UploadResponse{
			FileID:         file.ID,
			Filename:       file.Name,
			PublicFilename: file.PublicFilename.String,
		})
	} else {
		sendServerError(w)
	}
}

//ListFilesHandler handler for listing files
func ListFilesHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.FileListRequest
	if !readRequestLimited(w, r, &request, handlerData.config.Webserver.MaxRequestBodyLength) {
		return
	}

	//Select namespace
	namespace := models.FindNamespace(handlerData.db, request.Attributes.Namespace, handlerData.user)
	if namespace == nil || namespace.ID == 0 {
		sendResponse(w, models.ResponseError, "Namespace not found", 404)
		return
	}

	//Check if user can read from this namespace
	if !namespace.IsOwnedBy(handlerData.user) && !handlerData.user.CanReadForeignNamespace() {
		sendResponse(w, models.ResponseError, "Read permission denied for foreign namespaces", nil, http.StatusForbidden)
		return
	}

	//Gen Tags
	tags := models.FindTags(handlerData.db, request.Attributes.Tags, namespace)
	if len(tags) == 0 && len(request.Attributes.Tags) > 0 {
		sendResponse(w, models.ResponseError, "No matching tag found", 404)
		return
	}

	//Gen Groups
	groups := models.FindGroups(handlerData.db, request.Attributes.Groups, namespace)
	if len(groups) == 0 && len(request.Attributes.Groups) > 0 {
		sendResponse(w, models.ResponseError, "No matching group found", 404)
		return
	}

	loaded := handlerData.db
	if len(tags) > 0 || request.OptionalParams.Verbose > 1 {
		loaded = loaded.Preload("Tags")
	}

	if len(groups) > 0 || request.OptionalParams.Verbose > 1 {
		loaded = loaded.Preload("Groups")
	}

	if request.OptionalParams.Verbose > 2 {
		loaded = loaded.Preload("Namespace")
	}

	loaded = loaded.Where("namespace_id = ?", namespace.ID)

	if len(request.Name) > 0 {
		loaded = loaded.Where("name LIKE ?", "%"+request.Name+"%")
	}

	var foundFiles []models.File

	//search
	loaded.Find(&foundFiles)

	//Convert to ResponseFile
	var retFiles []models.FileResponseItem
	for _, file := range foundFiles {
		//Filter tags
		if (len(tags) == 0 || (len(tags) > 0 && file.IsInTagList(tags))) &&
			//Filter groups
			(len(groups) == 0 || (len(groups) > 0 && file.IsInGroupList(groups))) {
			respItem := models.FileResponseItem{
				ID:           file.ID,
				Name:         file.Name,
				CreationDate: file.CreatedAt,
				Size:         file.FileSize,
				IsPublic:     file.IsPublic,
			}

			//Append public name if available
			if file.PublicFilename.Valid && len(file.PublicFilename.String) > 0 {
				respItem.PublicName = file.PublicFilename.String
			}

			//Return attributes on verbose
			if request.OptionalParams.Verbose > 1 {
				respItem.Attributes = file.GetAttributes()
			}

			//Add if matching filter
			retFiles = append(retFiles, respItem)
		}
	}

	sendResponse(w, models.ResponseSuccess, "", models.ListFileResponse{
		Files: retFiles,
	})
}

//FileHandler handler for updating files
func FileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.FileRequest
	if !readRequestLimited(w, r, &request, handlerData.config.Webserver.MaxRequestBodyLength) {
		return
	}

	//Select namespace
	namespace := models.FindNamespace(handlerData.db, request.Attributes.Namespace, handlerData.user)
	if namespace == nil || namespace.ID == 0 {
		sendResponse(w, models.ResponseError, "Namespace not found", nil, http.StatusNotFound)
		return
	}

	//Check if user can access this namespace
	if !namespace.IsOwnedBy(handlerData.user) && !handlerData.user.CanWriteForeignNamespace() {
		sendResponse(w, models.ResponseError, "Write permission denied for foreign namespaces", nil, http.StatusForbidden)
		return
	}

	//Get action
	vars := mux.Vars(r)
	action, has := vars["action"]
	if !has {
		sendResponse(w, models.ResponseError, "missing action", nil)
		return
	}

	//Check if action is valid
	if !gaw.IsInStringArray(action, []string{"delete", "update", "get", "publish"}) {
		sendResponse(w, models.ResponseError, "invalid action", nil)
		return
	}

	//Get count of files with same name (ID only if provided)
	c, err := models.File{
		Name:      request.Name,
		Namespace: namespace,
	}.GetCount(handlerData.db, request.FileID, handlerData.user)

	//Handle errors
	if LogError(err) {
		sendServerError(w)
		return
	}

	// Get all files not allowed
	if request.All && action == "get" {
		sendResponse(w, models.ResponseError, "Illegal request", nil)
		return
	}

	var files []models.File

	// Only verify count if all is false
	if !request.All {
		//Send error if multiple files are available and no ID was specified
		if c > 1 && request.FileID == 0 {
			sendResponse(w, models.ResponseError, "multiple files with same name", nil)
			return
		}
	}

	//Exit if file not found
	if c == 0 {
		sendResponse(w, models.ResponseError, "Nothing found", nil)
		return
	}

	//Fill array with either one ore more instances
	if request.All {
		var err error
		files, err = models.FindFiles(handlerData.db, request.Name, *namespace, handlerData.user)
		if LogError(err) {
			sendServerError(w)
			return
		}
	} else {
		//Get target file
		file, err := models.FindFile(handlerData.db, request.Name, request.FileID, *namespace, handlerData.user)
		if LogError(err) {
			sendServerError(w)
			return
		}
		files = append(files, *file)
	}

	err = nil
	var didUpdate bool

	//Execute action
	switch action {
	case "delete":
		{
			for _, file := range files {
				err = file.Delete(handlerData.db, handlerData.config)
				if LogError(err) {
					break
				}
			}

			// Send response
			sendResponse(w, models.ResponseSuccess, "", models.CountResponse{
				Count: uint32(len(files)),
			})
		}
	case "update":
		{
			var count uint32

			//Do for every file
			for _, file := range files {
				update := request.Updates

				//Update namespace
				if len(update.NewNamespace) > 0 {
					//Get new namespace
					newNamespace := models.FindNamespace(handlerData.db, update.NewNamespace, handlerData.user)
					if newNamespace == nil || namespace.ID == 0 {
						sendResponse(w, models.ResponseError, "New namespace not found", nil, http.StatusNotFound)
						return
					}

					//Check if user can access this new namespace
					if !newNamespace.IsOwnedBy(handlerData.user) && !handlerData.user.CanWriteForeignNamespace() {
						sendResponse(w, models.ResponseError, "Write permission denied for foreign namespaces", nil, http.StatusForbidden)
						return
					}

					//Update files namespace
					err := file.UpdateNamespace(handlerData.db, newNamespace, handlerData.user)
					if LogError(err) {
						sendServerError(w)
						return
					}

					didUpdate = true
				}

				//Rename file
				if len(update.NewName) > 0 {
					if LogError(file.Rename(handlerData.db, update.NewName)) {
						sendServerError(w)
						return
					}
					didUpdate = true
				}

				//Set public/private
				if len(update.IsPublic) > 0 {
					if !file.PublicFilename.Valid {
						sendResponse(w, models.ResponseError, "You need to share this file first", nil)
						return
					}

					newVisibility, err := strconv.ParseBool(update.IsPublic)
					if err != nil {
						sendResponse(w, models.ResponseError, "isPublic must be a bool", nil, http.StatusUnprocessableEntity)
						return
					}

					if LogError(file.SetVilibility(handlerData.db, newVisibility)) {
						sendServerError(w)
						return
					}
					didUpdate = true
				}

				//Add tags
				if len(update.AddTags) > 0 {
					currLenTags := len(file.Tags)
					if LogError(file.AddTags(handlerData.db, update.AddTags, handlerData.user)) {
						sendServerError(w)
						return
					}
					didUpdate = len(file.Tags) > currLenTags
				}

				//Remove tags
				if len(update.RemoveTags) > 0 {
					currLenTags := len(file.Tags)
					if LogError(file.RemoveTags(handlerData.db, update.RemoveTags)) {
						sendServerError(w)
						return
					}
					didUpdate = len(file.Tags) < currLenTags
				}

				//Add Groups
				if len(update.AddGroups) > 0 {
					currLenGroups := len(file.Groups)
					if LogError(file.AddGroups(handlerData.db, update.AddGroups, handlerData.user)) {
						sendServerError(w)
						return
					}
					didUpdate = len(file.Groups) > currLenGroups
				}

				//Remove Groups
				if len(update.RemoveGroups) > 0 {
					currLenGroups := len(file.Groups)
					if LogError(file.RemoveGroups(handlerData.db, update.RemoveGroups)) {
						sendServerError(w)
						return
					}
					didUpdate = len(file.Groups) < currLenGroups
				}

				//Only count if updated
				if didUpdate {
					count++
				}
			}

			// Send response
			sendResponse(w, models.ResponseSuccess, "", models.CountResponse{
				Count: count,
			})
		}
	//Get file
	case "get":
		{
			//use first file
			file := files[0]

			//Open local file
			f, err := os.Open(handlerData.config.GetStorageFile(file.LocalName))
			if LogError(err) {
				if os.IsNotExist(err) {
					sendResponse(w, models.ResponseError, "File not found on server", nil, 404)
					return
				}

				sendServerError(w)
				return
			}

			//Set ContentType header
			if len(file.FileType) > 0 && filetype.IsMIMESupported(file.FileType) {
				w.Header().Set(models.HeaderContentType, file.FileType)
			}

			//Set filename header
			w.Header().Set(models.HeaderFileName, file.Name)

			//Write contents to responsewriter
			_, err = io.Copy(w, f)
			if LogError(err) {
				sendServerError(w)
				return
			}

			//Close file
			LogError(f.Close())
		}
	//Publish a file
	case "publish":
		{
			publishResponse := models.PublishResponse{}
			bulkPublishResponse := models.BulkPublishResponse{}

			for _, file := range files {
				//Ignore if already public
				if file.IsPublic {
					//Send error if publishing only one file
					if len(files) == 1 {
						sendResponse(w, models.ResponseError, "Already public", nil, http.StatusConflict)
						return
					}
					continue
				}

				nameTaken, err := file.Publish(handlerData.db, request.PublicName)
				if err != nil {
					sendServerError(w)
					return
				}
				if nameTaken {
					sendResponse(w, models.ResponseError, "public name already exists", nil)
					return
				}

				fmt.Println(file.PublicFilename.String)
				//use buik response if requested "all"
				if request.All {
					bulkPublishResponse.Files = append(bulkPublishResponse.Files, models.UploadResponse{
						FileID:         file.ID,
						Filename:       file.Name,
						PublicFilename: file.PublicFilename.String,
					})
				} else {
					//Otherwise respond with a single item
					publishResponse = models.PublishResponse{
						PublicFilename: file.PublicFilename.String,
					}
				}
			}

			//Send success
			if request.All {
				sendResponse(w, models.ResponseSuccess, "", bulkPublishResponse)
			} else {
				sendResponse(w, models.ResponseSuccess, "", publishResponse)
			}

		}
	}
}
