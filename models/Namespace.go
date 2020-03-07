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
	Name string
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
