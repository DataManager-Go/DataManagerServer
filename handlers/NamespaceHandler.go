package handlers

import (
	"net/http"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

// NamespaceActionHandler handler for namespace actions (create/update/delete)
func NamespaceActionHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, hasAction := vars["action"]

	// validate action and attribute kind
	if !hasAction || !gaw.IsInStringArray(action, []string{"update", "delete", "create"}) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	var request models.NamespaceRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	// Check for empty field
	if gaw.HasEmptyString(request.Namespace) {
		sendResponse(w, models.ResponseError, "Bad request", nil, http.StatusBadRequest)
		return
	}

	// Check permissions
	if !handlerData.User.CanCreateNamespaces() {
		sendResponse(w, models.ResponseError, "Not allowed to create user namespaces", nil, http.StatusForbidden)
		return
	}

	// Find namespace
	namespace := models.FindNamespace(handlerData.Db, request.Namespace, handlerData.User)

	// Do action type related checks
	if action == "create" {
		// Check if namespace already exists
		if namespace != nil && namespace.ID != 0 {
			sendResponse(w, models.ResponseError, "namespace already exists", nil, http.StatusBadRequest)
			return
		}
	} else {
		// Error if namespace not found/valid
		if !namespace.IsValid() {
			sendResponse(w, models.ResponseError, "namespace not found", nil, http.StatusNotFound)
			return
		}

		// on update, check if new name is not empty
		if action == "update" && len(request.NewName) == 0 {
			sendResponse(w, models.ResponseError, "no new name provided", nil, http.StatusUnprocessableEntity)
			return
		}
	}

	var err error

	switch action {
	case "create":
		{
			// Create and insert namespaceo
			namespace = &models.Namespace{
				Name:   handlerData.User.GetNamespaceName(request.Namespace),
				User:   handlerData.User,
				UserID: handlerData.User.ID,
			}

			err = namespace.Create(handlerData.Db)
		}
	case "update":
		{
			newName := handlerData.User.GetNamespaceName(request.NewName)

			// Check if namespace already exists
			newNS := models.FindNamespace(handlerData.Db, newName, handlerData.User)
			if newNS != nil {
				sendResponse(w, models.ResponseError, "namespace already exists", nil, http.StatusBadRequest)
				return
			}

			// Update namespace
			namespace.Name = newName
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

	// On any errors
	if LogError(err) {
		sendServerError(w)
		return
	}

	// Send success
	sendResponse(w, models.ResponseSuccess, "", models.StringResponse{
		String: namespace.Name,
	})
}

// NamespaceListHandler lists namespaces
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
