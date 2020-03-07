package models

import "github.com/jinzhu/gorm"

//Tag a filetag
type Tag struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Namespace   *Namespace
	NamespaceID int64 `sql:"index" gorm:"not null"`
}

//Insert inserts tag into DB
func (tag *Tag) Insert(db *gorm.DB) error {
	//Use default namespace if nil
	tag.Namespace = tag.GetNamespace()
	return db.Create(tag).Error
}

//GetNamespace return namespace of tag
func (tag Tag) GetNamespace() *Namespace {
	if tag.Namespace == nil {
		return &DefaultNamespace
	}

	return tag.Namespace
}

//TagsFromStringArr return tag array from string array
func TagsFromStringArr(arr []string, namespace Namespace) []Tag {
	var tags []Tag
	for _, tag := range arr {
		tags = append(tags, Tag{
			Name:      tag,
			Namespace: &namespace,
		})
	}
	return tags
}
