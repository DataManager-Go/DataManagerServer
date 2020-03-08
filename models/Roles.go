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

//HasUploadLimit gets upload limit
func (user User) HasUploadLimit() bool {
	return user.Role.MaxURLcontentSize > -1
}

//AllowedToUploadURLs gets upload limit
func (user User) AllowedToUploadURLs() bool {
	return user.Role.MaxURLcontentSize != 0
}

//CanUploadFiles return true if user can upload files
func (user User) CanUploadFiles() bool {
	return user.Role.CanUploadFiles
}

//CanWriteForeignNamespace return true if user is allowed to write in foreign namespaces
func (user User) CanWriteForeignNamespace() bool {
	return user.Role.AccesForeignNamespaces&Writepermission == Writepermission
}

//CanReadForeignNamespace return true if user is allowed to read in foreign namespaces
func (user User) CanReadForeignNamespace() bool {
	return user.Role.AccesForeignNamespaces&ReadPermission == ReadPermission
}
