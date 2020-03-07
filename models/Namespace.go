package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

//DefaultNamespace defalut namespace
var DefaultNamespace Namespace

//Namespace a namespace for files
type Namespace struct {
	gorm.Model
	Name   string `gorm:"not null"`
	UserID uint   `gorm:"column:uploader;index"`
	User   *User  `gorm:"association_autoupdate:false;association_autocreate:false"`
}

//GetNamespaceFromString return namespace from string
func GetNamespaceFromString(ns string) *Namespace {
	if len(ns) == 0 || strings.ToLower(ns) == "default" {
		return &DefaultNamespace
	}

	return &Namespace{
		Name: ns,
	}
}

//FindNamespace find namespace in DB
func FindNamespace(db *gorm.DB, ns string) *Namespace {
	namespace := GetNamespaceFromString(ns)
	if namespace.ID != 0 {
		return namespace
	}

	db.Where(&namespace).Find(&namespace)

	return namespace
}
