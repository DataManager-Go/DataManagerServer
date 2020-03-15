package handlers

import (
	"net/http"

	"github.com/JojiiOfficial/DataManagerServer/handlers/web"
	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

//AttributeHandler handler for attributes
func AttributeHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) {

	//Get vars
	vars := mux.Vars(r)
	attributeKind, hasAttribute := vars["attribute"]
	action, hasAction := vars["action"]

	//validate action and attribute kind
	if !hasAttribute || !hasAction || !gaw.IsInStringArray(action, []string{"update", "delete"}) || !gaw.IsInStringArray(attributeKind, []string{"tag", "group"}) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	//read request body
	var request models.UpdateAttributeRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	//Check for empty field
	if gaw.HasEmptyString(request.Name, request.Namespace) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	//Find namespace
	namespace := models.FindNamespace(handlerData.Db, request.Namespace, handlerData.User)

	// Handle namespace errors (not found || no access)
	if !handleNamespaceErorrs(namespace, handlerData.User, w) {
		return
	}

	if action == "update" && len(request.NewName) == 0 {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	if attributeKind == "tag" {
		//Find instance
		tag, err := models.FindTag(handlerData.Db, request.Name, namespace, handlerData.User)
		if tag == nil || err != nil {
			sendResponse(w, models.ResponseError, "Tag not found", nil, 404)
			return
		}

		//Do action for tag
		if action == "delete" {
			//Delete relations
			err = handlerData.Db.Unscoped().Table("files_tags").Where("tag_id=?", tag.ID).Delete(models.Tag{}).Error
			if LogError(err) {
				sendServerError(w)
				return
			}

			//Delete tags
			err = handlerData.Db.Delete(tag).Error
			if LogError(err) {
				sendServerError(w)
				return
			}
		} else if action == "update" {
			//Update tags name
			tag.Name = request.NewName
			err := handlerData.Db.Save(tag).Error

			if LogError(err) {
				sendServerError(w)
				return
			}
		}
	} else if attributeKind == "group" {
		//Find instance
		group, err := models.FindGroup(handlerData.Db, request.Name, namespace, handlerData.User)
		if group == nil || err != nil {
			sendResponse(w, models.ResponseError, "Group not found", nil, 404)
			return
		}

		//Do action for group
		if action == "delete" {
			//Delete relations
			err = handlerData.Db.Unscoped().Table("files_groups").Where("group_id=?", group.ID).Delete(models.Group{}).Error
			if LogError(err) {
				sendServerError(w)
				return
			}

			//Delete tags
			err = handlerData.Db.Delete(group).Error
			if LogError(err) {
				sendServerError(w)
				return
			}
		} else if action == "update" {
			//Update groups name
			group.Name = request.NewName
			err := handlerData.Db.Save(group).Error

			if LogError(err) {
				sendServerError(w)
				return
			}
		}
	}

	sendResponse(w, models.ResponseSuccess, "", nil)
}
