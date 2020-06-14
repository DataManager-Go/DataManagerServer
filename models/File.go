package models

import (
	"database/sql"
	"os"
	"strings"
	"time"

	libdm "github.com/DataManager-Go/libdatamanager"

	"github.com/JojiiOfficial/gaw"
	"github.com/JojiiOfficial/shred"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var shredder = shred.Shredder{}

// File a file uploaded to the db
type File struct {
	gorm.Model
	Name           string `gorm:"not null"`
	LocalName      string `gorm:"not null"`
	User           *User  `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID         uint   `gorm:"column:uploader;index"`
	FileSize       int64
	FileType       string
	IsPublic       bool           `gorm:"default:false"`
	PublicFilename sql.NullString `gorm:"unique"`
	Groups         []Group        `gorm:"many2many:files_groups;association_autoupdate:false"`
	Tags           []Tag          `gorm:"many2many:files_tags;association_autoupdate:false"`
	Namespace      *Namespace     `gorm:"association_autoupdate:false;association_autocreate:false;"`
	NamespaceID    uint           `sql:"index" gorm:"not null"`
	Encryption     sql.NullInt32
	Checksum       string
}

// FileAttributes attributes for a file
type FileAttributes struct {
	Tags      []string `json:"tags,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Namespace string   `json:"ns"`
}

// GetAttributes get file attributes
func (file File) GetAttributes() FileAttributes {
	return FileAttributes{
		Groups:    GroupArrToStringArr(file.Groups),
		Tags:      TagArrToStringArr(file.Tags),
		Namespace: file.GetNamespace().Name,
	}
}

// Insert inserts file into DB
func (file *File) Insert(db *gorm.DB, user *User) error {
	// Create groups
	for i := range file.Groups {
		if file.Groups[i].ID == 0 {
			if err := db.Where(&Group{
				Name: file.Groups[i].Name,
			}).Find(&file.Groups[i]).Error; err != nil {
				err = file.Groups[i].Insert(db, user)
				if err != nil {
					log.Warn(err)
				}
			}
		}
	}

	// Create tags
	for i := range file.Tags {
		if file.Tags[i].ID == 0 {
			if err := db.Where(&Tag{
				Name: file.Tags[i].Name,
			}).Find(&file.Tags[i]).Error; err != nil {
				err = file.Tags[i].Insert(db, user)
				if err != nil {
					log.Warn(err)
				}
			}
		}
	}

	// Use default namespace if not specified
	file.Namespace = file.GetNamespace()
	file.User = user

	// Create file
	if err := db.Create(file).Error; err != nil {
		return err
	}

	return nil
}

// GetNamespace return namespace of file
func (file File) GetNamespace() *Namespace {
	return file.Namespace
}

// IsInTagList return true if file has one of the specified tags
func (file File) IsInTagList(tags []string) bool {
	for _, tag := range file.Tags {
		for _, t1 := range tags {
			if tag.Name == t1 {
				return true
			}
		}
	}
	return false
}

// IsInGroupList return true if file is in one of the specified groups
func (file File) IsInGroupList(groups []string) bool {
	for _, group := range file.Groups {
		for _, t1 := range groups {
			if group.Name == t1 {
				return true
			}
		}
	}
	return false
}

// FindFiles finds file
func FindFiles(db *gorm.DB, file File) ([]File, error) {
	var files []File
	a := db.Model(&File{})

	// Filter by filename
	if len(file.Name) > 0 {
		a = a.Where("name like ?", file.Name)
	}

	// Filter by ID
	if file.ID != 0 {
		a = a.Where("id = ?", file.ID)
	}

	// Filter by namespace ID and uploader
	if file.Namespace != nil {
		a = a.Where("namespace_id = ? AND uploader = ?", file.Namespace.ID, file.Namespace.UserID)
	}

	// Get file to delete
	err := a.
		Preload("Namespace").
		Preload("Namespace.User").
		Preload("Tags").
		Preload("Groups").
		Find(&files).Error
	if err != nil {
		return nil, err
	}

	return files, nil
}

// FindFile finds a file
func FindFile(db *gorm.DB, fileID, userID uint) (*File, error) {
	a := db.Model(&File{}).Where("uploader = ?", userID)

	// Include ID if set. Otherwise use namespace
	if fileID != 0 {
		a = a.Where("id = ?", fileID)
	}

	// Get file
	var file File
	err := a.Preload("Namespace").First(&file).Error
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// HasTag return true if file is in group
func (file File) HasTag(sTag string) bool {
	for _, tag := range file.Tags {
		if tag.Name == sTag {
			return true
		}
	}
	return false
}

// HasGroup return true if file is in group
func (file File) HasGroup(sGroup string) bool {
	for _, group := range file.Groups {
		if group.Name == sGroup {
			return true
		}
	}
	return false
}

// Delete deletes a file
func (file *File) Delete(db *gorm.DB, config *Config) error {
	// Remove public filename to free this keyword
	file.IsPublic = false
	file.PublicFilename = sql.NullString{
		Valid: false,
	}

	// Save new state
	err := file.Save(db)
	if err != nil {
		return err
	}

	localFile := config.GetStorageFile(file.LocalName)
	s, err := os.Stat(localFile)
	if err != nil {
		log.Warn(err)
	} else {
		// Shredder file in background
		go ShredderFile(localFile, s.Size())
	}

	// Delete from DB
	return db.Delete(&file).Error
}

// ShredderFile shreddres a file
func ShredderFile(localFile string, size int64) {
	var shredConfig *shred.ShredderConf
	if size < 0 {
		s, err := os.Stat(localFile)
		if err != nil {
			log.Warn("File to shredder not found")
			return
		}
		size = s.Size()
	}

	if size >= 1000000000 {
		// Size >= 1GB
		shredConfig = shred.NewShredderConf(&shredder, shred.WriteZeros, 1, true)
	} else if size >= 10000000 {
		// Size >= 10MB
		shredConfig = shred.NewShredderConf(&shredder, shred.WriteZeros|shred.WriteRand, 1, true)
	} else {
		// Size < 10MB
		shredConfig = shred.NewShredderConf(&shredder, shred.WriteZeros|shred.WriteRandSecure, 3, true)
	}

	// Shredder & Delete local file
	start := time.Now()
	err := shredConfig.ShredFile(localFile)
	if err != nil {
		log.Error(err)
		// Delete file if shredder didn't
		err = os.Remove(localFile)
		if err != nil {
			log.Warn(err)
		}
	}
	log.Debug("Shredding took ", time.Since(start).String())
}

// Rename renames a file
func (file *File) Rename(db *gorm.DB, newName string) error {
	file.Name = newName
	return file.Save(db)
}

// SetVilibility sets public/private
func (file *File) SetVilibility(db *gorm.DB, newVisibility bool) error {
	file.IsPublic = newVisibility
	return file.Save(db)
}

// AddTags adds tags to file
func (file *File) AddTags(db *gorm.DB, tagsToAdd []string, user *User) error {
	for _, sTag := range tagsToAdd {
		if file.HasTag(sTag) {
			continue
		}

		tag := GetTag(db, sTag, file.Namespace, user)
		file.Tags = append(file.Tags, *tag)
	}

	return file.Save(db)
}

// RemoveTags remove tags to file
func (file *File) RemoveTags(db *gorm.DB, tagsToRemove []string) error {
	if len(file.Tags) == 0 {
		return nil
	}

	var newTags []Tag
	for i := range file.Tags {
		if !gaw.IsInStringArray(file.Tags[i].Name, tagsToRemove) {
			newTags = append(newTags, file.Tags[i])
		}
	}

	//Only save if at least one tag was removed
	if len(newTags) < len(file.Tags) {
		db.Model(&file).Association("Tags").Clear()
		file.Tags = newTags
		return file.Save(db)
	}

	return nil
}

// AddGroups adds groups to file
func (file *File) AddGroups(db *gorm.DB, groupsToAdd []string, user *User) error {
	for _, sGroup := range groupsToAdd {
		if file.HasGroup(sGroup) {
			continue
		}

		group := GetGroup(db, sGroup, file.Namespace, user)
		file.Groups = append(file.Groups, *group)
	}

	return file.Save(db)
}

// RemoveGroups remove groups from file
func (file *File) RemoveGroups(db *gorm.DB, groupsToRemove []string) error {
	if len(file.Groups) == 0 {
		return nil
	}

	var newGroups []Group
	for i := range file.Groups {
		if !gaw.IsInStringArray(file.Groups[i].Name, groupsToRemove) {
			newGroups = append(newGroups, file.Groups[i])
		}
	}

	// Only save if at least one group was removed
	if len(newGroups) < len(file.Groups) {
		db.Model(&file).Association("Groups").Clear()
		file.Groups = newGroups
		return file.Save(db)
	}

	return file.Save(db)
}

// Save saves a file in DB
func (file *File) Save(db *gorm.DB) error {
	return db.Save(file).Error
}

// GetCount get count if file
func (file File) GetCount(db *gorm.DB, fileID uint) (uint, error) {
	var c uint

	fileFilter := File{
		Model: file.Model,
	}

	// Include namespace in search if not nil
	if file.Namespace != nil {
		fileFilter.NamespaceID = file.NamespaceID
	}

	// Create count statement
	del := db.Model(&File{}).Where(&fileFilter).Where("deleted_at is NULL")

	if len(file.Name) > 0 {
		// Allow searching for wildcards
		if strings.HasSuffix(file.Name, "%") || strings.HasPrefix(file.Name, "%") {
			// Apply wildcard
			del = del.Where("name like ?", file.Name)
		} else {
			del = del.Where("name = ?", file.Name)
		}
	}

	// Also use fileID if set
	if fileID > 0 {
		del = del.Where("id = ?", fileID)
	}

	// Execute statement
	err := del.Count(&c).Error

	return c, err
}

// GetPublicFile returns a file which is public
func GetPublicFile(db *gorm.DB, publicFilename string) (*File, bool, error) {
	var file File
	err := db.Model(&File{}).Where("public_filename = ? AND is_public=true", publicFilename).First(&file).Error
	if err != nil {
		// Check error. Send server error if error is not "not found"
		if gorm.IsRecordNotFoundError(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return &file, true, nil
}

// UpdateNamespace updates namespace for file
func (file *File) UpdateNamespace(db *gorm.DB, newNamespace *Namespace, user *User) error {
	// Set new namespace
	file.Namespace = newNamespace
	file.NamespaceID = newNamespace.ID

	// Update/move tags if available
	if len(file.Tags) > 0 {
		var newTags []Tag
		for _, tag := range file.Tags {
			newTag := GetTag(db, tag.Name, newNamespace, user)
			newTags = append(newTags, *newTag)
		}
		// remove old tags
		db.Model(&file).Association("Tags").Clear()
		// Set new tags
		file.Tags = newTags
	}

	// Update/move groups if available
	if len(file.Groups) > 0 {
		var newGroups []Group
		for _, group := range file.Groups {
			newGroup := GetGroup(db, group.Name, newNamespace, user)
			newGroups = append(newGroups, *newGroup)
		}
		// remove old groups
		db.Model(&file).Association("Groups").Clear()
		// Set new groups
		file.Groups = newGroups
	}

	// Save file
	return db.Save(&file).Error
}

// Publish publis a file
func (file *File) Publish(db *gorm.DB, publicName string) (bool, error) {
	// Determine public name
	if len(publicName) == 0 {
		publicName = gaw.RandString(25)
	}

	// Set file public name
	file.PublicFilename = sql.NullString{
		String: publicName,
		Valid:  true,
	}
	file.IsPublic = true

	// Check if public name already exists
	_, found, _ := GetPublicFile(db, publicName)
	if found {
		return true, nil
	}

	// Save new file
	return false, file.Save(db)
}

// SetEncryption set encryption
func (file *File) SetEncryption(encription string) *File {
	e := sql.NullInt32{
		Valid: false,
	}
	if len(encription) > 0 && libdm.IsValidCipher(encription) {
		e.Valid = true
		e.Int32 = libdm.ChiperToInt(encription)
	}

	file.Encryption = e
	return file
}

// SetUniqueFilename sets unique filename
func (file *File) SetUniqueFilename(db *gorm.DB) bool {
	var localName string

	for i := 0; i < 5; i++ {
		localName = gaw.RandString(40)
		var c int
		db.Model(&File{}).Where(&File{LocalName: localName}).Count(&c)
		if c == 0 {
			file.LocalName = localName
			return true
		}

		log.Warningf("Name collision found. Trying again (%d/%d)", i, 5)
	}

	return false
}

// GetPublicNameWithExtension return the public name ending with the real
// file extension
func (file *File) GetPublicNameWithExtension() string {
	if !file.IsPublic {
		return ""
	}

	if !strings.Contains(file.Name, ".") {
		return file.PublicFilename.String
	}

	splitted := strings.Split(file.Name, ".")
	return file.PublicFilename.String + "." + splitted[len(splitted)-1]
}
