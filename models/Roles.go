package models

//Role roles for user
type Role struct {
	ID                            uint       `gorm:"pk"`
	RoleName                      string     `gorm:"not null"`
	IsAdmin                       bool       `gorm:"default:false"`
	AccesForeignFiles             Permission `gorm:"type:smallint"` //TODO
	AccesForeignNamespaces        Permission `gorm:"type:smallint"` //TODO
	CreateTagsInForeignNamespaces bool       `gorm:"default:false"` //TODO
	CanUploadFiles                bool       `gorm:"default:true"`
	MaxURLcontentSize             int64
}

//Permission permission for roles
type Permission uint8

//Permissions
const (
	NoPermission Permission = iota
	ReadPermission
	Writepermission
)
