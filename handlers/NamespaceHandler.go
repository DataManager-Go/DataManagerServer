package handlers

import (
	"net/http"
	"strings"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

//NamespaceActionHandler handler for namespace actions (create/update/delete)
func NamespaceActionHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	//Get vars
	vars := mux.Vars(r)
	action, hasAction := vars["action"]

	//validate action and attribute kind
	if !hasAction || !gaw.IsInStringArray(action, []string{"update", "delete", "create"}) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	var request models.NamespaceRequest
	if !readRequestLimited(w, r, &request, handlerData.config.Webserver.MaxRequestBodyLength) {
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
			if !handlerData.user.CanCreateUserNamespaces() {
				sendResponse(w, models.ResponseError, "Not allowed to create user namespaces", nil, http.StatusForbidden)
				return
			}

			//Set namespace name to usernamespace name if action is create
			if action == "create" && !strings.HasPrefix(namespaceName, handlerData.user.Username+"_") {
				namespaceName = handlerData.user.Username + "_" + namespaceName
			}

		}
	case models.CustomNamespaceType:
		{
			if !handlerData.user.CanCreateCustomNamespaces() {
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
	namespace := models.FindNamespace(handlerData.db, namespaceName)

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
			err = handlerData.db.Model(&models.Namespace{}).Create(&models.Namespace{
				Name:   namespaceName,
				User:   handlerData.user,
				UserID: handlerData.user.ID,
			}).Error
		}
	case "update":
		{
			// Update namespace
			namespace.Name = request.NewName
			err = handlerData.db.Model(&models.Namespace{}).Save(namespace).Error
		}
	case "delete":
		{
			// Delete namespace
			handlerData.db.Delete(namespace)
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
