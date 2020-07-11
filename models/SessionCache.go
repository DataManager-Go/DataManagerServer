package models

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Cache sessions
var sessionCache SessionCache

// SessionCache cache sessions in ram
type SessionCache struct {
	cache map[string]*sessionCacheEntry

	mx sync.Mutex
}

type sessionCacheEntry struct {
	db      *gorm.DB
	session *LoginSession
	valid   bool

	lastCheck              int64
	requestsSinceLastCheck uint32

	mx sync.Mutex
}

// init the Cache
func (sc *SessionCache) init() {
	sc.mx.Lock()
	defer sc.mx.Unlock()

	if sc.cache == nil {
		sc.cache = make(map[string]*sessionCacheEntry)
	}
}

// Get Session from Cache
func (sc *SessionCache) getSession(token string) *LoginSession {
	sc.init()

	sc.mx.Lock()
	defer sc.mx.Unlock()

	// Get cache entry
	session, ok := sc.cache[token]
	if !ok {
		return nil
	}

	// Update request counter
	session.requestsSinceLastCheck++

	// Remove session from
	// cache if invalid
	if !session.valid {
		delete(sc.cache, token)
		return nil
	}

	// Update if neccessary
	if session.needUpdate() {
		if !session.update() {
			return nil
		}
	}

	return session.session
}

// Add LoginSession to cache
func (sc *SessionCache) addSession(session *LoginSession, db *gorm.DB) {
	sc.init()

	sc.mx.Lock()
	defer sc.mx.Unlock()

	sc.cache[session.Token] = &sessionCacheEntry{
		db:        db,
		valid:     true,
		session:   session,
		lastCheck: time.Now().Unix(),
	}
}

func (sce *sessionCacheEntry) update() bool {
	sce.mx.Lock()
	defer sce.mx.Unlock()

	// Reload LoginSession frm DB
	session, err := loadSession(sce.session.Token, sce.db)
	if err != nil {
		log.Warn("Session invalid: ", err)
		sce.valid = false
		return false
	}

	// Update cache entry
	sce.session = session
	sce.valid = true
	sce.lastCheck = time.Now().Unix()
	sce.requestsSinceLastCheck = 0

	return true
}

// Allow 60s cache
const (
	maxCacheValidTime     = int64(10)
	maxRequestsUntilCheck = uint32(10)
)

// Return true if session needs update from DB
func (sce *sessionCacheEntry) needUpdate() bool {
	sce.mx.Lock()
	defer sce.mx.Unlock()

	return time.Now().Unix()-maxCacheValidTime > sce.lastCheck || sce.requestsSinceLastCheck > maxRequestsUntilCheck
}
