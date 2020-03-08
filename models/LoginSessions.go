package models

import (
	"github.com/jinzhu/gorm"
)

//LoginSession session for loggedin user
type LoginSession struct {
	gorm.Model
	User   *User `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID uint
	Token  string
}

//GetUserFromSession return user from session
func GetUserFromSession(db *gorm.DB, token string) (*User, error) {
	var session LoginSession
	err := db.Model(&LoginSession{}).Where(&LoginSession{
		Token: token,
	}).Preload("User").Preload("User.Role").Find(&session).Error

	if err != nil {
		return nil, err
	}

	return session.User, nil
}
