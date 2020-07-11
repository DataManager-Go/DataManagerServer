package models

import (
	"sync"

	"github.com/JojiiOfficial/gaw"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//LoginSession session for loggedin user
type LoginSession struct {
	gorm.Model
	User      *User `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID    uint
	Token     string
	Requests  int64
	MachineID string

	mx sync.Mutex `gorm:"-"`
}

// SessionTokenLength length of session token
const SessionTokenLength = 64

//GetUserFromSession return user from session
func GetUserFromSession(db *gorm.DB, token string) (*User, error) {
	var err error

	// Try to get session from cache
	session := sessionCache.getSession(token)
	if session == nil {
		// Load from DB if not cached
		session, err = loadSession(token, db)
		if err != nil {
			return nil, err
		}

		// Add token to cache
		sessionCache.addSession(session, db)
	}

	// Increase request counter
	go func() {
		session.mx.Lock()
		defer session.mx.Unlock()

		session.Requests++
		err = db.Model(&LoginSession{}).Save(session).Error
		if err != nil {
			log.Error(err)
		}
	}()

	return session.User, nil
}

func loadSession(token string, db *gorm.DB) (*LoginSession, error) {
	var ss LoginSession

	// load session from db
	err := db.Model(&LoginSession{}).Where(&LoginSession{
		Token: token,
	}).Preload("User").Preload("User.Role").Find(&ss).Error

	if err != nil {
		return nil, err
	}

	return &ss, nil
}

// NewSession create new login session
func NewSession(user *User, machineID string) *LoginSession {
	if len(machineID) > 100 {
		machineID = ""
	}

	var token string
	var err error
	tries := 0

	// Try to generate a token
	for tries < 5 {
		token, err = gaw.GenRandString(SessionTokenLength)
		if err != nil {
			log.Error(err)
		} else {
			break
		}
		tries++
	}

	// Return nil if token can't be generated
	if len(token) != 64 {
		return nil
	}

	//Generate session
	return &LoginSession{
		Token:     token,
		UserID:    user.ID,
		User:      user,
		MachineID: machineID,
	}
}
