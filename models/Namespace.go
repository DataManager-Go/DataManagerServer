package models

import (
	"github.com/jinzhu/gorm"
)

// Namespace a namespace for files
type Namespace struct {
	gorm.Model
	Name   string `gorm:"not null"`
	UserID uint   `gorm:"column:creator;index"`
	User   *User  `gorm:"association_autoupdate:false;association_autocreate:false"`
}

// GetNamespaceFromString return namespace from string
func GetNamespaceFromString(ns string) *Namespace {
	return &Namespace{
		Name: ns,
	}
}

// FindNamespace find namespace in DB
func FindNamespace(db *gorm.DB, ns string, user *User) *Namespace {
	// Add username prefix if not provided
	ns = user.GetNamespaceName(ns)

	var namespace Namespace
	err := db.Where(&Namespace{
		UserID: user.ID,
	}).Where("LOWER(name)=LOWER(?)", ns).Limit(1).Find(&namespace).Error

	if err != nil {
		return nil
	}

	return &namespace
}

// IsOwnedBy returns true if namespace is users
func (namespace *Namespace) IsOwnedBy(user *User) bool {
	if user == nil || namespace == nil {
		return false
	}

	return namespace.UserID == user.ID
}

// FindUserNamespaces get all namespaces for user
func FindUserNamespaces(db *gorm.DB, user *User) ([]Namespace, error) {
	var namespaces []Namespace

	err := db.Model(&Namespace{}).Where("creator = ?", user.ID).Find(&namespaces).Error
	if err != nil {
		return []Namespace{}, err
	}

	return namespaces, nil
}

//IsValid return true if namespace is valid
func (namespace *Namespace) IsValid() bool {
	return (namespace != nil && namespace.ID > 0)
}

// Create a namespace
func (namespace *Namespace) Create(db *gorm.DB) error {
	return db.Model(&Namespace{}).Create(namespace).Error
}
