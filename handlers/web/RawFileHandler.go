package web

import (
	"net/http"
	"os"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
)

//RawFileHandler handler for previews
func RawFileHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID  := vars["fileID"]

	//Get requested file
	file, found, err := models.GetPublicFile(handlerData.Db, fileID)
	if !found {
		NotFoundHandler(handlerData, w, r)
		return
	}

	//Send error
	if LogError(err) {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	//Send not found if not public
	if !file.IsPublic {
		NotFoundHandler(handlerData, w, r)
		return
	}

	//Set content type header if available and valid
	if len(file.FileType) > 0 && filetype.IsMIMESupported(file.FileType) {
		setContentType(w, file.FileType)
	}

	//Open file
	f, err := os.Open(handlerData.Config.GetStorageFile(file.LocalName))
	if LogError(err) {
		if os.IsNotExist(err) {
			NotFoundHandler(handlerData, w, r)
			return
		}

		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	serveFileStream(handlerData.Config, f, w)
}
