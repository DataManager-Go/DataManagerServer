package handlers

import (
	"net/http"
	"strings"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
)

// NamespaceActionHandler handler for namespace actions (create/update/delete)
func NamespaceActionHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	action, hasAction := vars["action"]

	// validate action and attribute kind
	if !hasAction || !gaw.IsInStringArray(action, []string{"update", "delete", "create"}) {
		return RErrBadRequest
	}

	var request libdm.NamespaceRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	// Check for empty field
	if gaw.HasEmptyString(request.Namespace) {
		return RErrBadRequest
	}

	// Check permissions
	if !handlerData.User.CanCreateNamespaces() {
		return RErrNotAllowed.Append("to create user namespaces")
	}

	// Find namespace
	namespace := models.FindNamespace(handlerData.Db, request.Namespace, handlerData.User)

	// Do action type related checks
	if action == "create" {
		// Check if namespace already exists
		if namespace != nil && namespace.ID != 0 {
			return RErrAlreadyExists.Prepend("Namespace")
		}
	} else {
		// Error if namespace not found/valid
		if !namespace.IsValid() {
			return RErrNotFound.Prepend("Namespace")
		}

		// on update, check if new name is not empty
		if action == "update" && len(request.NewName) == 0 {
			return NewRequestError("no new name provided", http.StatusUnprocessableEntity)
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

			// Check if namespace already exists. If the name is equal no the newname (ignoring case), accept new name since it can
			// have different casing
			newNS := models.FindNamespace(handlerData.Db, newName, handlerData.User)
			if newNS != nil && strings.ToLower(request.NewName) != strings.ToLower(request.Namespace) {
				return RErrAlreadyExists.Prepend("Namespace")
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
	if err != nil {
		return err
	}

	sendResponse(w, libdm.ResponseSuccess, "", libdm.StringResponse{
		String: namespace.Name,
	})

	return nil
}

// NamespaceListHandler lists namespaces
func NamespaceListHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	namespaces, err := models.FindUserNamespaces(handlerData.Db, handlerData.User)
	if err != nil {
		return err
	}

	var snamespaces []string
	for _, namespace := range namespaces {
		snamespaces = append(snamespaces, namespace.Name)
	}

	sendResponse(w, libdm.ResponseSuccess, "", libdm.StringSliceResponse{
		Slice: snamespaces,
	})

	return nil
}
