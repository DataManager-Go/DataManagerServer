package handlers

import (
	"net/http"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"
)

// Ping handles ping request
func Ping(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request libdm.PingRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	payload := "pong"

	auth := NewAuthHandler(r)
	if len(auth.GetBearer()) > 0 {
		payload = "Authorized pong"
	}

	response := models.StringResponse{
		String: payload,
	}
	sendResponse(w, models.ResponseSuccess, "", response)

	return nil
}
