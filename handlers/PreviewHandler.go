package handlers

import (
	"net/http"
	"os"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"
)

//Static files
const (
	NotFoundFile = "404.html"
	IndexFile    = "index.html"
	PreviewFile  = "Preview.html"
	FavIconFile  = "favicon.ico"
)

//RawFileHandler handler for previews
func RawFileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID, _ := vars["fileID"]

	//Get requested file
	file, found, err := models.GetPublicFile(handlerData.db, fileID)
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
	f, err := os.Open(handlerData.config.GetStorageFile(file.LocalName))
	if LogError(err) {
		if os.IsNotExist(err) {
			NotFoundHandler(handlerData, w, r)
			return
		}

		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = serveFileStream(handlerData.config, f, w)
	if LogError(err) {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

//PrevievFileHandler handler for previews
func PrevievFileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	//Return raw file if useragent is curl or wget
	if returnRawByUseragent(r.UserAgent()) {
		RawFileHandler(handlerData, w, r)
		return
	}

	vars := mux.Vars(r)
	fileID, _ := vars["fileID"]

	//Get requested file
	file, found, err := models.GetPublicFile(handlerData.db, fileID)
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

	//Serve preview
	LogError(serveTemplate(handlerData.config, PreviewFile, w, nil))
}

//FavIconHandler handle favicon
func FavIconHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	serveStaticFile(handlerData.config, FavIconFile, w)
}

//IndexPageHandler show index/main page
func IndexPageHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	handleBrowserServeError(
		//Try to serve index file
		serveStaticFile(handlerData.config, IndexFile, w),
		handlerData, w, r)
}

//NotFoundHandler 404 not found handler
func NotFoundHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	log.Info("Not found: ", r.URL.Path)

	err := serveStaticFile(handlerData.config, NotFoundFile, w)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("Can't find 404.html!")
			return
		}
	}
}
