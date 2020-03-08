package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
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

func (s *Sessions) CreateSession(ctx context.Context, username string) string {
	sessionToken, err := GenSessionToken()
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't generate token")
		return ""
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO "sessions" ("token", "username", "expiry") VALUES ($1,$2,$3);`, sessionToken, username, time.Now().Add(time.Hour*3).Unix())
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't exec statement")
	}
	return sessionToken
}

func (s *Sessions) InvalidateSession(ctx context.Context, username string) {
	_, err := s.db.ExecContext(ctx, `delete from sessions where username = $1`, username)
	if err != nil {
		log.Error().Err(err).Msg("InvalidateSession can't exec statement")
	}
}

func (s *Sessions) CheckToken(ctx context.Context, token string) (types.Session, bool) {
	rows, err := s.db.QueryContext(ctx, `select "username", "expiry" from sessions where token = $1`, token)
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
		_, err := s.db.Exec(`DELETE FROM sessions WHERE expiry < $1`, time.Now().Unix())
		if err != nil {
			log.Error().Err(err).Msg("Sessions can't exec statement")
		}
		time.Sleep(time.Second * 2)
	}
}
