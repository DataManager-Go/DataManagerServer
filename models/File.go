package models

import (
	"github.com/jinzhu/gorm"
)

//File a file uploaded to the db
type File struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Namespace   *Namespace
	NamespaceID int64   `sql:"index" gorm:"not null"`
	LocalName   string  `sql:"not null"`
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
			if err := db.Find(&file.Groups[i]).Error; err != nil {
				file.Groups[i].Insert(db)
			}
		}
	}

	//Create tags
	for i := range file.Tags {
		if file.Tags[i].ID == 0 {
			if err := db.Find(&file.Tags[i]).Error; err != nil {
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
