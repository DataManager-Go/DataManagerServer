package models

import (
	"github.com/jinzhu/gorm"
)

//File a file uploaded to the db
type File struct {
	gorm.Model
	Name        string `gorm:"not null"`
	LocalName   string `sql:"not null"`
	Namespace   *Namespace
	NamespaceID uint    `sql:"index" gorm:"not null"`
	Groups      []Group `gorm:"many2many:files_groups;association_autoupdate:false"`
	Tags        []Tag   `gorm:"many2many:files_tags;association_autoupdate:false"`
}

//FileAttributes attributes for a file
type FileAttributes struct {
	Tags      []string `json:"tags"`
	Groups    []string `json:"groups"`
	Namespace string   `json:"ns"`
}

//Insert inserts file into DB
func (file *File) Insert(db *gorm.DB) error {
	//Create groups
	for i := range file.Groups {
		if file.Groups[i].ID == 0 {
			if err := db.Where(&file.Groups[i]).Find(&file.Groups[i]).Error; err != nil {
				file.Groups[i].Insert(db)
			}
		}
	}

	//Create tags
	for i := range file.Tags {
		if file.Tags[i].ID == 0 {
			if err := db.Where(&file.Tags[i]).Find(&file.Tags[i]).Error; err != nil {
				file.Tags[i].Insert(db)
			}
		}
	}

	//Use default namespace if not specified
	file.Namespace = file.GetNamespace()

	//Create file
	if err := db.Debug().Create(file).Error; err != nil {
		return err
	}

	return nil
}

//GetNamespace return namespace of file
func (file File) GetNamespace() *Namespace {
	if file.Namespace == nil {
		return &DefaultNamespace
	}
	return file.Namespace
}

//IsInTagList return true if file has one of the specified tags
func (file File) IsInTagList(tags []Tag) bool {
	for _, tag := range file.Tags {
		for _, t1 := range tags {
			if tag.ID == t1.ID {
				return true
			}
		}
	}
	return false
}

//IsInGroupList return true if file is in one of the specified groups
func (file File) IsInGroupList(groups []Group) bool {
	for _, group := range file.Groups {
		for _, t1 := range groups {
			if group.ID == t1.ID {
				return true
			}
		}
	}
	return false
}
