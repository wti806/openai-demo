package store

import (
	"time"

	"github.com/ReneKroon/ttlcache"
)

type SessionStore struct {
	sessionThreadCache *ttlcache.Cache
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessionThreadCache: ttlcache.NewCache(),
	}
}

func (s *SessionStore) AddSession(sessionID string, threadID string) {
	s.sessionThreadCache.SetWithTTL(sessionID, threadID, 30*time.Minute)
}

func (s *SessionStore) GetSession(sessionID string) (string, bool) {
	threadID, exists := s.sessionThreadCache.Get(sessionID)
	return threadID.(string), exists

}
