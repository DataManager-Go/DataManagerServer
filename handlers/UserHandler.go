package handlers

import (
	"net/http"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"

	"github.com/JojiiOfficial/gaw"
)

// Login login handler
func Login(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request models.CredentialsRequest

	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	if len(request.Password) == 0 || len(request.Username) == 0 {
		return RErrMissing.Prepend("Input")
	}

	user := models.User{
		Username: request.Username,
		Password: gaw.SHA512(request.Username + request.Password),
	}

	session, err := user.Login(handlerData.Db, request.MachineID)
	if err != nil {
		return RErrInvalid.Append("credentials")
	}

	if session != nil {
		sendResponse(w, models.ResponseSuccess, "", models.LoginResponse{
			Token:     session.Token,
			Namespace: user.GetDefaultNamespaceName(),
		})
	} else {
		return NewRequestError("Error logging in", http.StatusUnauthorized)
	}

	return nil
}

// Register register handler
func Register(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	if !handlerData.Config.Server.AllowRegistration {
		sendResponse(w, models.ResponseError, "Server doesn't accept registrations", nil, http.StatusForbidden)
		return nil
	}

	var request models.CredentialsRequest

	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	if len(request.Password) == 0 || len(request.Username) == 0 {
		return RErrMissing.Prepend("Input")
	}

	user := models.User{
		Username: request.Username,
		Password: request.Password,
	}

	err := user.Register(handlerData.Db, handlerData.Config)
	if err == models.ErrorUserAlreadyExists {
		return RErrAlreadyExists.Prepend("User")
	} else if err != nil {
		return err
	}

	sendResponse(w, models.ResponseSuccess, "success", nil, http.StatusOK)

	return nil
}
