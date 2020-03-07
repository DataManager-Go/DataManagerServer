package handlers

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/JojiiOfficial/DataManagerServer/models"
	gaw "github.com/JojiiOfficial/GoAw"
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

	err = file.Insert(handlerData.db)
	if err != nil {
		sendServerError(w)
		log.Error(err)
	} else {
		sendResponse(w, models.ResponseSuccess, "", models.UploadResponse{
			FileID: file.ID,
		})
	}
}

//ListFilesHandler handler for uploading files
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

	loaded := handlerData.db.Preload("Tags").Preload("Groups").Where("namespace_id = ?", namespace.ID)

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

			//Add if matching filter
			retFiles = append(retFiles, models.FileResponseItem{
				ID:   file.ID,
				Name: file.Name,
			})
		}
	}

	response := models.ListFileResponse{
		Files: retFiles,
	}
	sendResponse(w, models.ResponseSuccess, "", response)
}
