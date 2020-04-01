package web

import (
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/JojiiOfficial/DataManagerServer/constants"
	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/gorilla/mux"
	"github.com/sbani/go-humanizer/units"
)

//Static files
const (
	NotFoundFile = "404.html"
	IndexFile    = "index.html"
	PreviewFile  = "Preview.html"
	FavIconFile  = "favicon.ico"
	ContentFile  = "Content.html"
)

//PrevievFileHandler handler for previews
func PrevievFileHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	//Return raw file if useragent in the list of raw useragents
	if handlerData.Config.IsRawUseragent(strings.ToLower(r.UserAgent())) {
		RawFileHandler(handlerData, w, r)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["fileID"]

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
		Host:           r.Host,
		FileSizeStr:    units.BinarySuffix(float64(file.FileSize)),
		Encrypted:      (file.Encryption.Valid && constants.EncryptionIValid(file.Encryption.Int32)),
	}

	//Serve preview
	LogError(servePreviewTemplate(handlerData.Config, w, templateData))
}

func servePreviewTemplate(config *models.Config, w http.ResponseWriter, data interface{}) error {
	PreviewFile := config.GetTemplateFile(PreviewFile)
	ContentFile := config.GetTemplateFile(ContentFile)

	templateName := path.Base(PreviewFile)

	//Create template
	t := template.New("")
	t.Funcs(template.FuncMap{
		"IsImagePreview":   models.IsImagePreview,
		"IsTextPreview":    models.IsTextPreview,
		"IsDefaultPreview": models.IsDefaultPreview,
	})

	t, err := t.ParseFiles(PreviewFile, ContentFile)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, templateName, data)
}
