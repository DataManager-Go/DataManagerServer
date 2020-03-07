package models

import "github.com/jinzhu/gorm"

//Roles roles for user
type Roles struct {
	gorm.Model
	ForeignFiles      uint8
	ForeignNamespaces uint8
	CreateTags        bool
}

//Permission permission for roles
type Permission uint8

//Permissions
const (
	ReadPermission Permission = iota
	Writepermission
)
