package handlers

import (
	"net/http"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"

	libdm "github.com/DataManager-Go/libdatamanager"
)

// ListFilesHandler handler for listing files
func ListFilesHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	var request libdm.FileListRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	var namespace *models.Namespace

	if !request.AllNamespaces {
		// Select namespace
		namespace = models.FindNamespace(handlerData.Db, request.Attributes.Namespace, handlerData.User)

		// Handle namespace errors (not found || no access)
		if !handleNamespaceErorrs(namespace, handlerData.User, w) {
			return nil
		}
	}

	var foundFiles []models.File

	loaded := handlerData.Db.Model(&foundFiles)
	if len(request.Attributes.Tags) > 0 || request.OptionalParams.Verbose > 1 {
		loaded = loaded.Preload("Tags")
	}

	if len(request.Attributes.Groups) > 0 || request.OptionalParams.Verbose > 1 {
		loaded = loaded.Preload("Groups")
	}

	if request.OptionalParams.Verbose > 1 || request.AllNamespaces {
		loaded = loaded.Preload("Namespace")
	}

	if len(request.Name) > 0 {
		loaded = loaded.Where("files.name LIKE ?", "%"+request.Name+"%")
	}

	if request.AllNamespaces {
		// Join to filter by namespace creator
		loaded = loaded.
			Joins("INNER JOIN namespaces ON namespaces.id = files.namespace_id").
			Where("namespaces.creator = ?", handlerData.User.ID)
	} else {
		// Just select the specified namespace
		loaded = loaded.Where("namespace_id = ?", namespace.ID)
	}

	// Search
	err := loaded.Find(&foundFiles).Error
	if err != nil {
		return err
	}

	// Convert to ResponseFile
	var retFiles []libdm.FileResponseItem
	for _, file := range foundFiles {
		// Filter tags
		if (len(request.Attributes.Tags) == 0 || (len(request.Attributes.Tags) > 0 && file.IsInTagList(request.Attributes.Tags))) &&
			// Filter groups
			(len(request.Attributes.Groups) == 0 || (len(request.Attributes.Groups) > 0 && file.IsInGroupList(request.Attributes.Groups))) {
			respItem := libdm.FileResponseItem{
				ID:           file.ID,
				Name:         file.Name,
				CreationDate: file.CreatedAt,
				Size:         file.FileSize,
				IsPublic:     file.IsPublic,
				Checksum:     file.Checksum,
			}

			// Set encryption
			if file.Encryption.Valid && libdm.EncryptionIValid(file.Encryption.Int32) {
				respItem.Encryption = libdm.ChiperToString(file.Encryption.Int32)
			}

			// Append public name if available
			if file.PublicFilename.Valid && len(file.PublicFilename.String) > 0 {
				respItem.PublicName = file.PublicFilename.String
			}

			// Return attributes on verbose
			if request.OptionalParams.Verbose > 1 || request.AllNamespaces {
				respItem.Attributes = file.GetAttributes()
			}

			// Add if matching filter
			retFiles = append(retFiles, respItem)
		}
	}

	sendResponse(w, libdm.ResponseSuccess, "", libdm.FileListResponse{
		Files: retFiles,
	})

	return nil
}
