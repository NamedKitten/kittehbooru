package database

import (
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"sync"
	"time"
)

type Sessions struct {
	sessions map[string]types.Session
	lock     sync.Mutex
}

func (s *Sessions) init() {
	if s.sessions == nil {
		s.sessions = make(map[string]types.Session)
	}
}

func (s *Sessions) CreateSession(username string) string {
	s.lock.Lock()
	sessionToken := utils.GenSessionToken()
	s.sessions[sessionToken] = types.Session{Username: username, ExpirationTime: time.Now().Add(time.Hour * 3).Unix()}
	s.lock.Unlock()
	return sessionToken
}

func (s *Sessions) InvalidateSession(username string) {
	s.lock.Lock()
	for token, session := range s.sessions {
		if session.Username == username {
			delete(s.sessions, token)
		}
	}
	s.lock.Unlock()
}

func (s *Sessions) CheckToken(token string) (types.Session, bool) {
	s.lock.Lock()
	sess, ok := s.sessions[token]
	s.lock.Unlock()
	return sess, ok
}

func (s *Sessions) Start() {
	s.init()
	for true {
		s.lock.Lock()
		for token, session := range s.sessions {
			if time.Now().After(time.Unix(session.ExpirationTime, 0)) {
				delete(s.sessions, token)
			}
		}
		s.lock.Unlock()

		time.Sleep(time.Second * 2)
	}
}
