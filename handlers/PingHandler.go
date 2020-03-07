package handlers

import (
	"net/http"

	"github.com/JojiiOfficial/DataManagerServer/models"
)

//Ping handles ping request
func Ping(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	var request models.PingRequest
	if !parseUserInput(handlerData.config, w, r, &request) {
		return
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
}
