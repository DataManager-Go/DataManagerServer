package handlers

import (
	"net/http"
	"strings"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

//NamespaceActionHandler handler for namespace actions (create/update/delete)
func NamespaceActionHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) {
	//Get vars
	vars := mux.Vars(r)
	action, hasAction := vars["action"]

	//validate action and attribute kind
	if !hasAction || !gaw.IsInStringArray(action, []string{"update", "delete", "create"}) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	var request models.NamespaceRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	//Check for empty field
	if gaw.HasEmptyString(request.Namespace) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	namespaceName := request.Namespace

	// Check permissions
	switch request.Type {
	case models.UserNamespaceType:
		{
			if !handlerData.User.CanCreateUserNamespaces() {
				sendResponse(w, models.ResponseError, "Not allowed to create user namespaces", nil, http.StatusForbidden)
				return
			}

			//Set namespace name to usernamespace name if action is create
			if action == "create" && !strings.HasPrefix(namespaceName, handlerData.User.Username+"_") {
				namespaceName = handlerData.User.Username + "_" + namespaceName
			}

		}
	case models.CustomNamespaceType:
		{
			if !handlerData.User.CanCreateCustomNamespaces() {
				sendResponse(w, models.ResponseError, "Not allowed to create custom namespaces", nil, http.StatusForbidden)
				return
			}
		}
	default:
		{
			sendResponse(w, models.ResponseError, "Invalid Type", nil, http.StatusBadRequest)
			return
		}
	}

	//Find namespace
	namespace := models.FindNamespace(handlerData.Db, namespaceName, handlerData.User)

	if action == "create" {
		//Check if namespace already exists
		if namespace.ID != 0 {
			sendResponse(w, models.ResponseError, "namespace already exists", nil, http.StatusBadRequest)
			return
		}
	} else {
		//Error if namespace not found
		if namespace.ID == 0 {
			sendResponse(w, models.ResponseError, "namespace not found", nil, http.StatusNotFound)
			return
		}
	}

	var err error

	switch action {
	case "create":
		{
			// Create namespaceo
			err = handlerData.Db.Model(&models.Namespace{}).Create(&models.Namespace{
				Name:   namespaceName,
				User:   handlerData.User,
				UserID: handlerData.User.ID,
			}).Error
		}
	case "update":
		{
			// Update namespace
			namespace.Name = request.NewName
			err = handlerData.Db.Model(&models.Namespace{}).Save(namespace).Error
		}
	case "delete":
		{
			// Delete namespace
			handlerData.Db.Delete(namespace)
			handlerData.Db.Delete(&models.Tag{}, "namespace_id=?", namespace.ID)
			handlerData.Db.Delete(&models.Group{}, "namespace_id=?", namespace.ID)
			handlerData.Db.Delete(&models.File{}, "namespace_id=?", namespace.ID)
		}
	}

	if LogError(err) {
		sendServerError(w)
		return
	}

	//Send success
	sendResponse(w, models.ResponseSuccess, "", models.StringResponse{
		String: namespaceName,
	})
}

//NamespaceListHandler lists namespaces
func NamespaceListHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) {
	namespaces, err := models.FindUserNamespaces(handlerData.Db, handlerData.User)
	if LogError(err) {
		sendServerError(w)
		return
	}

	var snamespaces []string
	for _, namespace := range namespaces {
		snamespaces = append(snamespaces, namespace.Name)
	}

	sendResponse(w, models.ResponseSuccess, "", models.StringSliceResponse{
		Slice: snamespaces,
	})
}
