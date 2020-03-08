package handlers

import (
	"io/ioutil"
	"net/http"
	"os"

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

	//Data validation
	if GetMD5Hash(request.Data) != request.Sum {
		sendResponse(w, models.ResponseError, "Content wasn't delivered completely", nil, 422)
		return
	}

	//Select namespace
	namespace := models.GetNamespaceFromString(request.Attributes.Namespace)

	//Gen Tags
	tags := models.TagsFromStringArr(request.Attributes.Tags, *namespace)

	//Gen Groups
	groups := models.GroupsFromStringArr(request.Attributes.Groups, *namespace)

	file := models.File{
		Groups:    groups,
		Tags:      tags,
		Namespace: namespace,
		Name:      request.Name,
	}

	//Ensure localname is not already in use
	uniqueNameFound := false
	for i := 0; i < 5; i++ {
		file.LocalName = gaw.RandString(40)
		var c int
		handlerData.db.Model(&models.File{}).Where(&models.File{LocalName: file.LocalName}).Count(&c)
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

	//Write file
	err := ioutil.WriteFile(handlerData.config.GetStorageFile(file.LocalName), request.Data, 0700)
	if err != nil {
		sendServerError(w)
		log.Error(err)
		return
	}

	//Get filesize
	s, _ := os.Stat(handlerData.config.GetStorageFile(file.LocalName))
	file.FileSize = s.Size()

	err = file.Insert(handlerData.db, handlerData.user)
	if err != nil {
		sendServerError(w)
		log.Error(err)
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

	loaded := handlerData.db.Preload("Tags").Preload("Groups").Preload("Namespace").Where("namespace_id = ?", namespace.ID)

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
		sendResponse(w, models.ResponseError, "Namespace not found", 404)
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
	if err != nil {
		log.Error(err)
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

	if err != nil {
		sendServerError(w)
		return
	}

	sendResponse(w, models.ResponseSuccess, "success", nil)
}
