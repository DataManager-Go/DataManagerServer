package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/JojiiOfficial/DataManagerServer/models"
	gaw "github.com/JojiiOfficial/GoAw"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

//UploadfileHandler handler for uploading files
func UploadfileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.UploadRequest
	if !parseUserInput(handlerData.config, w, r, &request) {
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

			//Data validation
			if GetMD5Hash(request.Data) != request.Sum {
				sendResponse(w, models.ResponseError, "Content wasn't delivered completely", nil, http.StatusUnprocessableEntity)
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
	namespace := models.FindNamespace(handlerData.db, request.Attributes.Namespace)
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
	tags := models.TagsFromStringArr(request.Attributes.Tags, *namespace)

	//Get Groups
	groups := models.GroupsFromStringArr(request.Attributes.Groups, *namespace)

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

	//set local name
	file.LocalName = localName

	//Create local file
	f, err := os.Create(handlerData.config.GetStorageFile(localName))
	if LogError(err) {
		sendServerError(w)
		return
	}

	switch request.UploadType {
	case models.FileUploadType:
		size, err := f.Write(request.Data)
		if LogError(err) {
			sendServerError(w)
			return
		}

		file.FileSize = int64(size)
	case models.URLUploadType:
		//Do request
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
	if LogError(err) {
		sendServerError(w)
	} else {
		sendResponse(w, models.ResponseSuccess, "", models.UploadResponse{
			FileID: file.ID,
		})
	}
}

//ListFilesHandler handler for listing files
func ListFilesHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.FileRequest
	if !parseUserInput(handlerData.config, w, r, &request) {
		return
	}

	//Select namespace
	namespace := models.FindNamespace(handlerData.db, request.Attributes.Namespace)
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
			}

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

//UpdateFileHandler handler for updating files
func UpdateFileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.FileRequest
	if !parseUserInput(handlerData.config, w, r, &request) {
		return
	}

	//Select namespace
	namespace := models.FindNamespace(handlerData.db, request.Attributes.Namespace)
	if namespace == nil || namespace.ID == 0 {
		sendResponse(w, models.ResponseError, "Namespace not found", http.NotFound)
		return
	}

	//Check if user can access this namespace
	if !namespace.IsOwnedBy(handlerData.user) && !handlerData.user.CanWriteForeignNamespace() {
		sendResponse(w, models.ResponseError, "Write permission denied for foreign namespaces", nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	action, has := vars["action"]
	if !has {
		sendResponse(w, models.ResponseError, "missing action", nil)
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

	//Send error if multiple files are available and no ID was specified
	if c > 1 && request.FileID == 0 {
		sendResponse(w, models.ResponseError, "multiple files with same name", nil)
		return
	}

	//Exit if file not found
	if c == 0 {
		sendResponse(w, models.ResponseError, "File not found", nil)
		return
	}

	err = nil

	//Execute action
	switch action {
	case "delete":
		{
			err = models.DeleteFile(handlerData.db, request.FileID, namespace, request.Name, handlerData.user, handlerData.config)
		}
	}

	if LogError(err) {
		sendServerError(w)
		return
	}

	sendResponse(w, models.ResponseSuccess, "success", nil)
}
