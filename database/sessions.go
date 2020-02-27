package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
	"time"
)

type Sessions struct {
	db *sql.DB
}

func GenSessionToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Sessions) CreateSession(username string) string {
	sessionToken, err := GenSessionToken()
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't generate token")
		return ""
	}
	_, err = s.db.Exec(`INSERT INTO "sessions" ("token", "username", "expiry") VALUES (?,?,?);`, sessionToken, username, time.Now().Add(time.Hour*3).Unix())
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't exec statement")
	}
	return sessionToken
}

func (s *Sessions) InvalidateSession(username string) {
	_, err := s.db.Exec(`delete from sessions where username = ?`, username)
	if err != nil {
		log.Error().Err(err).Msg("InvalidateSession can't exec statement")
	}
}

func (s *Sessions) CheckToken(token string) (types.Session, bool) {
	rows, err := s.db.Query(`select "username", "expiry" from sessions where token = ?`, token)
	if err != nil {
		log.Error().Err(err).Msg("CheckToken can't query")
		return types.Session{}, false
	}

	defer rows.Close()

	var username string
	var expiry int64
	for rows.Next() {
		err = rows.Scan(&username, &expiry)
		if err != nil {
			log.Error().Err(err).Msg("CheckToken can't scan rows")
			return types.Session{}, false
		}
	}

	return types.Session{Username: username, ExpirationTime: expiry}, true
}

func (s *Sessions) Start(db *sql.DB) {
	s.db = db
	for {
		_, err := s.db.Exec(`delete from sessions where expiry < ?`, time.Now().Unix())
		if err != nil {
			log.Error().Err(err).Msg("Sessions can't exec statement")
		}
		time.Sleep(time.Second * 2)
	}
}
