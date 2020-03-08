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
	RoleID   uint  `sql:"index"`
	Role     *Role `gorm:"association_autoupdate:false;association_autocreate:false"`
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
func (user User) Register(db *gorm.DB, config *Config) error {
	//Return if user already exists
	has, _ := user.Has(db, false)
	if has {
		return ErrorUserAlreadyExists
	}

	user = User{
		Password: gaw.SHA512(user.Username + user.Password),
		Username: user.Username,
		RoleID:   config.GetDefaultRole().ID,
		Role:     config.GetDefaultRole(),
	}

	err := db.Create(&user).Error
	if err != nil {
		return err
	}

	//Create namespace for user
	_, err = user.CreateDefaultNamespace(db)
	return err
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

//GetDefaultNamespaceName return the name of the default namespace for a user
func (user *User) GetDefaultNamespaceName() string {
	return user.Username + "_default"
}

//CreateDefaultNamespace creates user namespace
func (user *User) CreateDefaultNamespace(db *gorm.DB) (*Namespace, error) {
	namespace := Namespace{
		Name: user.GetDefaultNamespaceName(),
		User: user,
	}

	//Crreate
	err := db.Create(&namespace).Error
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}
