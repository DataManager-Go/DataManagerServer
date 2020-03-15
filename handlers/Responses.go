package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JojiiOfficial/DataManagerServer/models"
)

func handleAndSendError(err error, w http.ResponseWriter, message string, statusCode int) bool {
	if !LogError(err) {
		return false
	}
	sendResponse(w, models.ResponseError, message, nil, statusCode)
	return true
}

func sendServerError(w http.ResponseWriter) {
	sendResponse(w, models.ResponseError, "internal server error", nil, http.StatusInternalServerError)
}

//Return true on success
func handleNamespaceErorrs(namespace *models.Namespace, user *models.User, w http.ResponseWriter) bool {
	// Check if namespace was found
	if !namespace.IsValid() {
		sendResponse(w, models.ResponseError, "Namespace not found", nil, http.StatusNotFound)
		return false
	}

	// Check if user can access this namespace
	if !user.HasAccess(namespace) {
		sendResponse(w, models.ResponseError, "Write permission denied for this namespace", nil, http.StatusForbidden)
		return false
	}

	return true
}

func sendResponse(w http.ResponseWriter, status models.ResponseStatus, message string, payload interface{}, params ...int) {
	statusCode := http.StatusOK
	s := "0"
	if status == 1 {
		s = "1"
	}

	w.Header().Set(models.HeaderStatus, s)
	w.Header().Set(models.HeaderStatusMessage, message)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if len(params) > 0 {
		statusCode = params[0]
		w.WriteHeader(statusCode)
	}

	var err error
	if payload != nil {
		err = json.NewEncoder(w).Encode(payload)
	} else if len(message) > 0 {
		_, err = fmt.Fprintln(w, message)
	}

	LogError(err)
}
