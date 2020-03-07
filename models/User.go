package models

import (
	gaw "github.com/JojiiOfficial/GoAw"
	"github.com/jinzhu/gorm"
)

//User user in db
type User struct {
	gorm.Model
	Username string
	Password string
}

//Login login user
func (user User) Login(db *gorm.DB) (*LoginSession, error) {
	token := gaw.RandString(64)

	//Return if user not exists
	if has, err := user.Has(db, true); !has {
		return nil, err
	}

	//Generate session
	session := LoginSession{
		Token:  token,
		UserID: user.ID,
		User:   &user,
	}

	//Save session
	if err := db.Create(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

//Register register user
func (user User) Register(db *gorm.DB) error {
	//Return if user already exists
	has, _ := user.Has(db, false)
	if has {
		return ErrorUserAlreadyExists
	}

	return db.Create(&User{
		Password: gaw.SHA512(user.Username + user.Password),
		Username: user.Username,
	}).Error
}

//Has return true if user exists
func (user *User) Has(db *gorm.DB, checkPass bool) (bool, error) {
	pass := ""
	if checkPass {
		pass = user.Password
	}
	//Check if user exists
	if err := db.Where(&User{
		Username: user.Username,
		Password: pass,
	}).First(user).Error; err != nil {
		return false, err
	}
	return true, nil
}
