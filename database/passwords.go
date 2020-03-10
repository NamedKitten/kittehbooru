package database

import (
	"context"
	"database/sql"
	"runtime/trace"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func encryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (db *DB) SetPassword(ctx context.Context, username string, password string) (err error) {
	defer trace.StartRegion(ctx, "DB/SetPassword").End()

	passwordBytes, err := encryptPassword(password)
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't encrypt password")
		return err
	}

	_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "passwords" ("username", "password") VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password`, username, string(passwordBytes))
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't execute statement")
		return err
	}
	return nil
}

func (db *DB) CheckPassword(ctx context.Context, username string, password string) bool {
	defer trace.StartRegion(ctx, "DB/CheckPassword").End()

	var encPasswd string
	row := db.sqldb.QueryRowContext(ctx, `select password from passwords where username=$1`, username)
	switch err := row.Scan(&encPasswd); err {
	case sql.ErrNoRows:
		return false
	case nil:

		return bcrypt.CompareHashAndPassword([]byte(encPasswd), []byte(password)) == nil
	default:
		return false
	}
}
