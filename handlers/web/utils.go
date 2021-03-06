package web

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/JojiiOfficial/gaw"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//HandlerData handlerData for web
type HandlerData struct {
	Config *models.Config
	Db     *gorm.DB
	User   *models.User
}

//LogError returns true on error
func LogError(err error) bool {
	if err == nil {
		return false
	}

	log.Error(err.Error())

	return true
}

//Copy stream
func serveFileStream(config *models.Config, reader io.Reader, w http.ResponseWriter) {
	_ = gaw.BufferedCopy(config.Webserver.DownloadFileBuffer, w, reader)
}

//Detect and set Content-Type by extension
func autoSetContentType(w http.ResponseWriter, file string) {
	setContentType(w, mime.TypeByExtension(file))
}

//Set Content-Type
func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set(libdm.HeaderContentType, fmt.Sprintf("%s; charset=utf-8", contentType))
}

//Serve static file
func serveStaticFile(config *models.Config, file string, w http.ResponseWriter, contentType ...string) error {
	//Open file
	f, err := os.Open(config.GetHTMLFile(file))
	if LogError(err) {
		return err
	}
	defer f.Close()

	//Set contentType
	if len(contentType) == 0 || len(contentType[0]) == 0 {
		autoSetContentType(w, file)
	} else {
		w.Header().Set(libdm.HeaderContentType, contentType[0])
	}

	serveFileStream(config, f, w)
	return nil
}

//Handles errors and respond with 404 if this caused the error
func handleBrowserServeError(err error, handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	if err != nil {
		fmt.Println(err)
		if os.IsNotExist(err) {

			NotFoundHandler(handlerData, w, r)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}
