package handlers

import (
	"net/http"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/JojiiOfficial/gaw"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// AttributeHandler handler for file attributes.
// Implements update, delete, get and create functions
// for tags and groups
func AttributeHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	// Get vars
	vars := mux.Vars(r)
	attributeKind, hasAttribute := vars["attribute"]
	action, hasAction := vars["action"]

	// Validate action and attribute kind
	if !hasAttribute ||
		!hasAction ||
		!gaw.IsInStringArray(action, []string{"update", "delete", "get", "create"}) ||
		!gaw.IsInStringArray(attributeKind, []string{"tag", "group"}) {

		return RErrBadRequest
	}

	// Read request body
	var request models.UpdateAttributeRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return nil
	}

	// Check for empty field
	if gaw.HasEmptyString(request.Namespace) {
		return RErrBadRequest
	}

	// Find namespace and handle namespace errors (not found || no access)
	namespace := models.FindNamespace(handlerData.Db, request.Namespace, handlerData.User)
	if !handleNamespaceErorrs(namespace, handlerData.User, w) {
		return nil
	}

	// check newName availability
	if action == "update" && len(request.NewName) == 0 {
		return RErrBadRequest
	}

	if attributeKind == "tag" {
		// Do action for tag
		switch action {
		case "delete", "update":
			{
				// Check required field availability
				if len(request.Name) == 0 {
					return RErrBadRequest
				}

				// Find instance
				tag, err := models.FindTag(handlerData.Db, request.Name, namespace, handlerData.User)
				if tag == nil || LogError(err) {
					return RErrNotFound.Prepend("Tag")
				}

				switch action {
				case "update":
					{
						// Update tags name
						tag.Name = request.NewName
						err := handlerData.Db.Save(tag).Error

						if LogError(err) {
							sendServerError(w)
							return nil
						}
					}
				case "delete":
					{
						// Delete relations
						err = handlerData.Db.Unscoped().Table("files_tags").Where("tag_id=?", tag.ID).Delete(models.Tag{}).Error
						if err != nil {
							return err
						}

						// Delete tags
						err = handlerData.Db.Delete(tag).Error
						if err != nil {
							return err
						}
					}
				}
			}
		case "get":
			{

				var tags []models.Tag
				err := handlerData.Db.Model(&models.Tag{}).Where("namespace_id=?", namespace.ID).Find(&tags).Error
				if err != nil {
					return err
				}

				sendResponse(w, models.ResponseSuccess, "", models.TagArrToStringArr(tags))
				return nil
			}
		case "create":
			{
				// Check if tag already exists
				tag, err := models.FindTag(handlerData.Db, request.Name, namespace, handlerData.User)
				if err == nil && tag != nil && tag.ID > 0 {
					return RErrAlreadyExists.Prepend("Tag")
				}

				tag = &models.Tag{
					NamespaceID: namespace.ID,
					Name:        request.Name,
					UserID:      handlerData.User.ID,
				}

				// Save tag
				err = tag.Insert(handlerData.Db, handlerData.User)
				if err != nil {
					return err
				}

				sendResponse(w, models.ResponseSuccess, "", nil)
				return nil
			}
		}
	} else if attributeKind == "group" {
		switch action {
		case "delete", "update":
			{
				// Check required field availability
				if len(request.Name) == 0 {
					return RErrBadRequest
				}

				// Find instance
				group, err := models.FindGroup(handlerData.Db, request.Name, namespace, handlerData.User)
				if group == nil || err != nil {
					return RErrNotFound.Prepend("Group")
				}

				// Do desired action
				switch action {
				case "delete":
					{

						// Delete relations
						err = handlerData.Db.Unscoped().Table("files_groups").Where("group_id=?", group.ID).Delete(models.Group{}).Error
						if err != nil {
							return err
						}

						// Delete tags
						err = handlerData.Db.Delete(group).Error
						if err != nil {
							return err
						}
					}
				case "update":
					{
						// Update groups name
						group.Name = request.NewName
						err := handlerData.Db.Save(group).Error
						if err != nil {
							return err
						}
					}
				}
			}
		case "get":
			{

				var groups []models.Group
				err := handlerData.Db.Model(&models.Group{}).Where("namespace_id=?", namespace.ID).Find(&groups).Error
				if err != nil {
					return err
				}

				sendResponse(w, models.ResponseSuccess, "", models.GroupArrToStringArr(groups))
				return nil
			}
		case "create":
			{
				// Check if tag already exists
				group, err := models.FindGroup(handlerData.Db, request.Name, namespace, handlerData.User)
				if err == nil && group != nil && group.ID > 0 {
					return RErrAlreadyExists.Prepend("Group")
				}

				group = &models.Group{
					NamespaceID: namespace.ID,
					Name:        request.Name,
					UserID:      handlerData.User.ID,
				}

				// Save tag
				err = group.Insert(handlerData.Db, handlerData.User)
				if err != nil {
					return err
				}
			}
		}
	}

	sendResponse(w, models.ResponseSuccess, "", nil)
	return nil
}

// UserAttributeHandler handler for getting user attribute informations
func UserAttributeHandler(handlerData web.HandlerData, w http.ResponseWriter, r *http.Request) error {
	// Get groups
	groups, err := handlerData.User.GetAllGroups(handlerData.Db)
	var nss []models.Namespace
	getNamespaces := make(chan error, 1)

	go func() {
		nss, err = models.FindUserNamespaces(handlerData.Db, handlerData.User)
		getNamespaces <- err
	}()

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return RErrNotFound
		}

		return err
	}

	nsMap := make(map[string][]models.Group)

	// Create map with namespace as key
	for i := range groups {
		t, ok := nsMap[groups[i].Namespace.Name]
		if !ok {
			t = []models.Group{}
		}

		nsMap[groups[i].Namespace.Name] = append(t, groups[i])
	}

	var response models.UserAttributeDataResponse
	response.Namespace = make([]models.Namespaceinfo, len(nsMap))

	i := 0
	// Loop map and build response
	for ns, groups := range nsMap {
		respItem := models.Namespaceinfo{Name: ns}
		respItem.Groups = make([]string, len(groups))
		for i := range groups {
			respItem.Groups[i] = groups[i].Name
		}

		response.Namespace[i] = respItem
		i++
	}

	err = <-getNamespaces
	if err != nil {
		return err
	}

	// Add namespaces which aren't assigned to any groups
	for i := range nss {
		_, ok := nsMap[nss[i].Name]
		if !ok {
			response.Namespace = append(response.Namespace, models.Namespaceinfo{
				Name: nss[i].Name,
			})
		}
	}

	// Send response
	sendResponse(w, models.ResponseSuccess, "", response)
	return nil
}
