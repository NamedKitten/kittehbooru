package database

import (
	"database/sql"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/rs/zerolog/log"
	"time"
)

type Sessions struct {
	db *sql.DB
}

func (s *Sessions) CreateSession(username string) string {
	sessionToken := utils.GenSessionToken()
	stmt, err := s.db.Prepare(`INSERT INTO "sessions" ("token", "username", "expiry") VALUES (?,?,?);`)
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't prepare statement")
	}
	_, err = stmt.Exec(sessionToken, username, time.Now().Add(time.Hour*3).Unix())
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't execute statement")
	}
	return sessionToken
}

func (s *Sessions) InvalidateSession(username string) {
	stmt, err := s.db.Prepare(`delete from sessions where username = ?`)
	if err != nil {
		log.Error().Err(err).Msg("InvalidateSession can't prepare statement")
	}
	_, err = stmt.Exec(username)
	if err != nil {
		log.Error().Err(err).Msg("InvalidateSession can't exec statement")
	}
}

func (s *Sessions) CheckToken(token string) (types.Session, bool) {
	rows, err := s.db.Query(`select "username", "expiry" from sessions where token = ?`, token)
	if err != nil {
		log.Error().Err(err).Msg("CheckToken can't prepare statement")
	}

	defer rows.Close()

	var username string
	var expiry int64
	for rows.Next() {
		rows.Scan(&username, &expiry)
	}

	return types.Session{Username: username, ExpirationTime: expiry}, username != ""
}

func (s *Sessions) Start(db *sql.DB) {
	s.db = db
	for true {
		stmt, err := s.db.Prepare(`delete from sessions where expiry < ?`)
		if err != nil {
			log.Error().Err(err).Msg("Sessions can't prepare statement")
		}
		_, err = stmt.Exec(time.Now().Unix())
		if err != nil {
			log.Error().Err(err).Msg("Sessions can't exec statement")
		}
		time.Sleep(time.Second * 2)
	}
}
