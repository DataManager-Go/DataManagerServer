package handlers

import (
	"net/http"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

//AttributeHandler handler for attributes
func AttributeHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) {
	// Get vars
	vars := mux.Vars(r)
	attributeKind, hasAttribute := vars["attribute"]
	action, hasAction := vars["action"]

	// Validate action and attribute kind
	if !hasAttribute || !hasAction || !gaw.IsInStringArray(action, []string{"update", "delete", "get"}) || !gaw.IsInStringArray(attributeKind, []string{"tag", "group"}) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	// Read request body
	var request models.UpdateAttributeRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	// Check for empty field
	if gaw.HasEmptyString(request.Namespace) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	// Find namespace and handle namespace errors (not found || no access)
	namespace := models.FindNamespace(handlerData.Db, request.Namespace, handlerData.User)
	if !handleNamespaceErorrs(namespace, handlerData.User, w) {
		return
	}

	// check newName availability
	if action == "update" && len(request.NewName) == 0 {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	if attributeKind == "tag" {
		// Do action for tag
		switch action {
		case "delete", "update":
			{
				// Check required field availability
				if len(request.Name) == 0 {
					sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
					return
				}

				// Find instance
				tag, err := models.FindTag(handlerData.Db, request.Name, namespace, handlerData.User)
				if tag == nil || LogError(err) {
					sendResponse(w, models.ResponseError, "Tag not found", nil, 404)
					return
				}

				switch action {
				case "update":
					{
						// Update tags name
						tag.Name = request.NewName
						err := handlerData.Db.Save(tag).Error

						if LogError(err) {
							sendServerError(w)
							return
						}
					}
				case "delete":
					{
						// Delete relations
						err = handlerData.Db.Unscoped().Table("files_tags").Where("tag_id=?", tag.ID).Delete(models.Tag{}).Error
						if LogError(err) {
							sendServerError(w)
							return
						}

						// Delete tags
						err = handlerData.Db.Delete(tag).Error
						if LogError(err) {
							sendServerError(w)
							return
						}
					}
				}
			}
		case "get":
			{

				var tags []models.Tag
				err := handlerData.Db.Model(&models.Tag{}).Where("namespace_id=?", namespace.ID).Find(&tags).Error
				if err != nil {
					sendServerError(w)
					return
				}

				sendResponse(w, models.ResponseSuccess, "", models.TagArrToStringArr(tags))
				return
			}
		}
	} else if attributeKind == "group" {
		switch action {
		case "delete", "update":
			{
				// Check required field availability
				if len(request.Name) == 0 {
					sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
					return
				}

				// Find instance
				group, err := models.FindGroup(handlerData.Db, request.Name, namespace, handlerData.User)
				if group == nil || err != nil {
					sendResponse(w, models.ResponseError, "Group not found", nil, 404)
					return
				}

				// Do desired action
				switch action {
				case "delete":
					{

						// Delete relations
						err = handlerData.Db.Unscoped().Table("files_groups").Where("group_id=?", group.ID).Delete(models.Group{}).Error
						if LogError(err) {
							sendServerError(w)
							return
						}

						// Delete tags
						err = handlerData.Db.Delete(group).Error
						if LogError(err) {
							sendServerError(w)
							return
						}
					}
				case "update":
					{
						// Update groups name
						group.Name = request.NewName
						err := handlerData.Db.Save(group).Error
						if LogError(err) {
							sendServerError(w)
							return
						}
					}
				}
			}
		case "get":
			{

				var groups []models.Group
				err := handlerData.Db.Model(&models.Group{}).Where("namespace_id=?", namespace.ID).Find(&groups).Error
				if err != nil {
					sendServerError(w)
					return
				}

				sendResponse(w, models.ResponseSuccess, "", models.GroupArrToStringArr(groups))
				return
			}
		}
	}

	sendResponse(w, models.ResponseSuccess, "", nil)
}
