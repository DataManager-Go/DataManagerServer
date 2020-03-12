package web

import (
	"net/http"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/gorilla/mux"
)

//Static files
const (
	NotFoundFile = "404.html"
	IndexFile    = "index.html"
	PreviewFile  = "Preview.html"
	FavIconFile  = "favicon.ico"
)

//PrevievFileHandler handler for previews
func PrevievFileHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	//Return raw file if useragent is curl or wget
	if returnRawByUseragent(r.UserAgent()) {
		RawFileHandler(handlerData, w, r)
		return
	}

	vars := mux.Vars(r)
	fileID, _ := vars["fileID"]

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

	templateData := models.PreviewTemplate{
		Filename:       file.Name,
		PublicFilename: file.PublicFilename.String,
		PreviewType:    models.PreviewTypeFromMime(file.FileType),
	}

	//Serve preview
	LogError(serveTemplate(handlerData.Config, PreviewFile, w, templateData))
}
