package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

// genSessionToken generates a 32 byte long random session token
func genSessionToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a session for a user, returning a session token
func (db *DB) CreateSession(ctx context.Context, username string) string {
	sessionToken, err := genSessionToken()
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't generate token")
		return ""
	}
	_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "sessions" ("token", "username", "expiry") VALUES ($1,$2,$3);`, sessionToken, username, time.Now().Add(time.Hour*3).Unix())
	if err != nil {
		log.Error().Err(err).Msg("CreateSession can't exec statement")
	}
	return sessionToken
}

// InvalidateSession invalidates a session for a user
func (db *DB) InvalidateSession(ctx context.Context, username string) {
	_, err := db.sqldb.ExecContext(ctx, `delete from sessions where username = $1`, username)
	if err != nil {
		log.Error().Err(err).Msg("InvalidateSession can't exec statement")
	}
}

// CheckToken checks if there is a session for the user, and returns info on it if available
func (db *DB) CheckToken(ctx context.Context, token string) (types.Session, error) {
	rows, err := db.sqldb.QueryContext(ctx, `select "username", "expiry" from sessions where token = $1`, token)
	if err != nil {
		log.Error().Err(err).Msg("CheckToken can't query")
		return types.Session{}, err
	}

	defer rows.Close()

	var username string
	var expiry int64
	for rows.Next() {
		err = rows.Scan(&username, &expiry)
		if err != nil {
			log.Error().Err(err).Msg("CheckToken can't scan rows")
			return types.Session{}, err
		}
	}

	return types.Session{Username: username, ExpirationTime: expiry}, nil
}

// sessionCleaner removes any expired sessions from the database every 10 seconds
func (db *DB) sessionCleaner() {
	for {
		_, err := db.sqldb.Exec(`DELETE FROM sessions WHERE expiry < $1`, time.Now().Unix())
		if err != nil {
			log.Error().Err(err).Msg("Sessions can't exec statement")
		}
		time.Sleep(time.Second * 10)
	}
}
