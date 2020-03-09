package handlers

import (
	"net/http"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
)

//Login login handler
//-> /user/login
func Login(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.CredentialsRequest

	if !readRequestLimited(w, r, &request, handlerData.config.Webserver.MaxRequestBodyLength) {
		return
	}

	if isStructInvalid(request) {
		sendResponse(w, models.ResponseError, "input missing", nil, http.StatusUnprocessableEntity)
		return
	}

	user := models.User{
		Username: request.Username,
		Password: gaw.SHA512(request.Username + request.Password),
	}

	session, err := user.Login(handlerData.db)
	if err != nil {
		sendResponse(w, models.ResponseError, "Invalid credentials", nil)
		return
	}

	if session != nil {
		sendResponse(w, models.ResponseSuccess, "", models.LoginResponse{
			Token:     session.Token,
			Namespace: user.GetDefaultNamespaceName(),
		})
	} else {
		sendResponse(w, models.ResponseError, "Error logging in", nil, http.StatusUnauthorized)
	}
}

//Register register handler
//-> /user/create
func Register(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	if !handlerData.config.Server.AllowRegistration {
		sendResponse(w, models.ResponseError, "Server doesn't accept registrations", nil, http.StatusForbidden)
		return
	}

	var request models.CredentialsRequest

	if !readRequestLimited(w, r, &request, handlerData.config.Webserver.MaxRequestBodyLength) {
		return
	}

	if isStructInvalid(request) {
		sendResponse(w, models.ResponseError, "input missing", nil, http.StatusUnprocessableEntity)
		return
	}

	user := models.User{
		Username: request.Username,
		Password: request.Password,
	}

	err := user.Register(handlerData.db, handlerData.config)
	if err == models.ErrorUserAlreadyExists {
		sendResponse(w, models.ResponseError, "User already exists", nil)
	} else if err != nil {
		return
	}

	sendResponse(w, models.ResponseSuccess, "success", nil, http.StatusOK)
}
