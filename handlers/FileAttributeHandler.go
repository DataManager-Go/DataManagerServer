package handlers

import (
	"net/http"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

//AttributeHandler handler for attributes
func AttributeHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {

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
	if !readRequestLimited(w, r, &request, handlerData.config.Webserver.MaxRequestBodyLength) {
		return
	}

	//Check for empty field
	if gaw.HasEmptyString(request.Name, request.Namespace) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	//Find namespace
	namespace := models.FindNamespace(handlerData.db, request.Namespace)
	if namespace == nil {
		sendResponse(w, models.ResponseError, "namespace not found", nil, http.StatusNotFound)
		return
	}

	//Check if user can access this namespace
	if !namespace.IsOwnedBy(handlerData.user) && !handlerData.user.CanWriteForeignNamespace() {
		sendResponse(w, models.ResponseError, "Write permission denied for foreign namespaces", nil, http.StatusForbidden)
		return
	}

	if action == "delete" {
		//Delete
		if attributeKind == "tag" {
			tag, err := models.FindTag(handlerData.db, request.Name, namespace, handlerData.user)
			if tag == nil || err != nil {
				sendResponse(w, models.ResponseError, "Tag not found", nil, 404)
				return
			}

			//Delete relations
			err = handlerData.db.Unscoped().Table("files_tags").Where("tag_id=?", tag.ID).Delete(models.Tag{}).Error
			if LogError(err) {
				sendServerError(w)
				return
			}

			//Delete tags
			err = handlerData.db.Delete(tag).Error
			if LogError(err) {
				sendServerError(w)
				return
			}
		} else {
			group, err := models.FindGroup(handlerData.db, request.Name, namespace, handlerData.user)
			if group == nil || err != nil {
				sendResponse(w, models.ResponseError, "Group not found", nil, 404)
				return
			}

			//Delete relations
			err = handlerData.db.Unscoped().Table("files_groups").Where("group_id=?", group.ID).Delete(models.Group{}).Error
			if LogError(err) {
				sendServerError(w)
				return
			}

			//Delete tags
			err = handlerData.db.Delete(group).Error
			if LogError(err) {
				sendServerError(w)
				return
			}
		}
	} else {
		//Update

		//Ensure newName is not empty
		if len(request.NewName) == 0 {
			sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
			return
		}

	}

	sendResponse(w, models.ResponseSuccess, "", nil)
}
