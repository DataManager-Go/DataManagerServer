package models

import "github.com/jinzhu/gorm"

//LoginSession session for loggedin user
type LoginSession struct {
	gorm.Model
	User   *User `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID uint
	Token  string
}
