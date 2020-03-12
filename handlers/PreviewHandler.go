package handlers

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
)

//Static files
const (
	NotFoundFile = "404.html"
	IndexFile    = "index.html"
	FavIconFile  = "favicon.ico"
)

//PrevievRawFileHandler handler for previews
func PrevievRawFileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
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

	http.Error(w, "Server error", http.StatusInternalServerError)
}

//PrevievFileHandler handler for previews
func PrevievFileHandler(handlerData handlerData, w http.ResponseWriter, r *http.Request) {
	serveStaticFile(handlerData.config, IndexFile, w)
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
	logrus.Info("Not found: ", r.URL.Path)

	err := serveStaticFile(handlerData.config, NotFoundFile, w)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Error("Can't find 404.html!")
			return
		}
	}
}

//Handles errors and respond with 404 if this caused the error
func handleBrowserServeError(err error, handerData handlerData, w http.ResponseWriter, r *http.Request) {
	if err != nil {
		fmt.Println(err)
		if os.IsNotExist(err) {
			NotFoundHandler(handerData, w, r)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}

//Serve static file
func serveStaticFile(config *models.Config, file string, w http.ResponseWriter, contentType ...string) error {
	//Open file
	f, err := os.Open(config.GetHTMLFile(file))
	defer f.Close()

	if LogError(err) {
		return err
	}

	//Set contenttype
	if len(contentType) == 0 || len(contentType[0]) == 0 {
		autoSetContentType(w, file)
	} else {
		w.Header().Set(models.HeaderContentType, contentType[0])
	}

	return serveFileStream(config, f, w)
}

//Copy stream
func serveFileStream(config *models.Config, reader io.Reader, w http.ResponseWriter) error {
	err := gaw.BufferedCopy(config.Webserver.DownloadFileBuffer, w, reader)
	//Ignore EOF
	if err == io.EOF {
		return nil
	}
	return err
}

//Detect and set Content-Type by extension
func autoSetContentType(w http.ResponseWriter, file string) {
	setContentType(w, mime.TypeByExtension(file))
}

//Set Content-Type
func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set(models.HeaderContentType, fmt.Sprintf("%s; charset=utf-8", contentType))
}
