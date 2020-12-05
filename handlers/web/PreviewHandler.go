package web

import (
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/DataManager-Go/DataManagerServer/models"
	libdm "github.com/DataManager-Go/libdatamanager"

	"github.com/gorilla/mux"
	"github.com/sbani/go-humanizer/units"
)

//Static files
const (
	NotFoundFile = "index.html"
	IndexFile    = "index.html"
	PreviewFile  = "Preview.html"
	FavIconFile  = "favicon.ico"
	ContentFile  = "Content.html"
)

//PrevievFileHandler handler for previews
func PrevievFileHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) error {
	//Return raw file if useragent in the list of raw useragents
	if handlerData.Config.IsRawUseragent(strings.ToLower(r.UserAgent())) {
		RawFileHandler(handlerData, w, r)
		return nil
	}

	vars := mux.Vars(r)
	fileID := vars["fileID"]

	//Get requested file
	file, found, err := models.GetPublicFile(handlerData.Db, fileID)
	if !found {
		NotFoundHandler(handlerData, w, r)
		return nil
	}

	//Send error
	if LogError(err) {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return nil
	}

	//Send not found if not public
	if !file.IsPublic {
		NotFoundHandler(handlerData, w, r)
		return nil
	}

	scheme := "http"
	if len(handlerData.Config.Webserver.SchemeOverwrite) > 0 {
		scheme = handlerData.Config.Webserver.SchemeOverwrite

		// Prevent selecting other schemes than http(s)
		switch scheme {
		case "http", "https":
		default:
			scheme = "http"
		}
	} else if r.TLS != nil {
		scheme = "https"
	}

	templateData := models.PreviewTemplate{
		Filename:       file.Name,
		PublicFilename: file.PublicFilename.String,
		PreviewType:    models.PreviewTypeFromMime(file.FileType),
		Host:           r.Host,
		FileSizeStr:    units.BinarySuffix(float64(file.FileSize)),
		Encrypted:      (file.Encryption.Valid && libdm.EncryptionIValid(file.Encryption.Int32)),
		MimeType:       file.FileType,
		Scheme:         scheme,
		AceTheme:       handlerData.Config.Webserver.AceTheme,
	}

	//Serve preview
	LogError(servePreviewTemplate(handlerData.Config, w, templateData))
	return nil
}

func servePreviewTemplate(config *models.Config, w http.ResponseWriter, data interface{}) error {
	PreviewFile := config.GetTemplateFile(PreviewFile)
	ContentFile := config.GetTemplateFile(ContentFile)

	templateName := path.Base(PreviewFile)

	//Create template
	t := template.New("")
	t.Funcs(template.FuncMap{
		"IsImagePreview":   models.IsImagePreview,
		"IsVideoPreview":   models.IsVideoPreview,
		"IsTextPreview":    models.IsTextPreview,
		"IsDefaultPreview": models.IsDefaultPreview,
	})

	t, err := t.ParseFiles(PreviewFile, ContentFile)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, templateName, data)
}
