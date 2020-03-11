package handlers

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

//Static files
const (
	NotFoundFile = "404.html"
	IndexFile    = "index.html"
)

//PrevievFileHandler handler for previews
func PrevievFileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID, _ := vars["fileID"]

	//search file
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

	//Set content type header if available
	if len(file.FileType) > 0 {
		w.Header().Set("Content-Type", file.FileType)
	}

	//Open file
	f, err := os.Open(handlerData.config.GetStorageFile(file.LocalName))
	if LogError(err) {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	//Copy stream
	gaw.BufferedCopy(handlerData.config.Webserver.DownloadFileBuffer, w, f)

	if LogError(err) {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	//Close file
	LogError(f.Close())
}

//IndexPageHandler show index/main page
func IndexPageHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	handleBrowserServeError(
		//Try to serve index file
		serveSingleFile(handlerData.config.GetHTMLFile(IndexFile), w),
		handlerData, w, r)
}

//NotFoundHandler 404 not found handler
func NotFoundHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	err := serveSingleFile(handlerData.config.GetHTMLFile(NotFoundFile), w)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Error("Can't find 404.html!")
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}

//Handles errors and respond with 404 if this caused the error
func handleBrowserServeError(err error, handerData handlerData, w http.ResponseWriter, r *http.Request) {
	if err != nil {
		if os.IsNotExist(err) {
			NotFoundHandler(handerData, w, r)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}

//Serve static file
func serveSingleFile(file string, w http.ResponseWriter) error {
	page, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mime.TypeByExtension(file)))
	_, err = w.Write(page)
	return err
}
