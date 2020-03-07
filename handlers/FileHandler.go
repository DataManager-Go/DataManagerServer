package handlers

import (
	"net/http"
	"strings"

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
	request.Attributes.Namespace = strings.ToLower(request.Attributes.Namespace)

	//Select namespace
	namespace := models.GetNamespaceFromString(request.Attributes.Namespace)

	//Gen Tags
	tags := models.TagsFromStringArr(request.Attributes.Tags, *namespace)

	//Gen Groups
	groups := models.GroupsFromStringArr(request.Attributes.Groups, *namespace)

	file := models.File{
		Groups:    groups,
		Tags:      tags,
		LocalName: gaw.RandString(40),
		Namespace: namespace,
		Name:      request.Name,
	}

	err := file.Insert(handlerData.db)
	if err != nil {
		sendServerError(w)
		log.Error(err)
	} else {
		sendResponse(w, models.ResponseSuccess, "success", nil)
	}
}
