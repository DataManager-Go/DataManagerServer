package models

import (
	"gorm.io/gorm"
)

// Tag a filetag
type Tag struct {
	gorm.Model
	Name        string     `gorm:"not null"`
	NamespaceID uint       `sql:"index" gorm:"not null"`
	Namespace   *Namespace `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID      uint       `sql:"index" gorm:"not null"`
	User        *User      `gorm:"association_autoupdate:false;association_autocreate:false"`
}

// Insert inserts tag into DB
func (tag *Tag) Insert(db *gorm.DB, user *User) error {
	//Use default namespace if nil
	tag.Namespace = tag.GetNamespace()
	tag.User = user
	return db.Create(tag).Error
}

// GetNamespace return namespace of tag
func (tag Tag) GetNamespace() *Namespace {
	return tag.Namespace
}

// TagsFromStringArr return tag array from string array
func TagsFromStringArr(arr []string, namespace Namespace, user *User) []Tag {
	var tags []Tag
	for _, tag := range arr {
		tags = append(tags, Tag{
			Name:      tag,
			User:      user,
			UserID:    user.ID,
			Namespace: &namespace,
		})
	}
	return tags
}

// FindTags find namespace in DB
func FindTags(db *gorm.DB, sTags []string, namespace *Namespace) []Tag {
	var tags []Tag
	db.Model(&Tag{}).Where("name in (?) AND namespace_id = ?", sTags, namespace.ID).Find(&tags)
	return tags
}

// TagArrToStringArr return string arr from tags
func TagArrToStringArr(tags []Tag) []string {
	var str []string
	for _, tag := range tags {
		str = append(str, tag.Name)
	}
	return str
}

// GetTag returns or creates a tag
func GetTag(db *gorm.DB, name string, namespace *Namespace, user *User) *Tag {
	var tag Tag
	db.Where(&Tag{
		Name:        name,
		NamespaceID: namespace.ID,
		UserID:      user.ID,
	}).FirstOrCreate(&tag)

	return &tag
}

// FindTag finds a tag
func FindTag(db *gorm.DB, name string, namespace *Namespace, user *User) (*Tag, error) {
	var tag Tag

	err := db.Where(&Tag{
		Name:        name,
		NamespaceID: namespace.ID,
		UserID:      user.ID,
	}).First(&tag).Error

	if err != nil {
		return nil, err
	}

	return &tag, nil
}
