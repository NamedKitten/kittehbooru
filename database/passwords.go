package database

import (
	"context"
	"database/sql"
	"runtime/trace"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// encryptPassword returns a bcrypt hash of the password
func encryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// SetPassword sets the password for a user, it takes in a unencrypted password.
func (db *DB) SetPassword(ctx context.Context, username string, password string) (err error) {
	defer trace.StartRegion(ctx, "DB/SetPassword").End()

	// Encrypt the password and get the byte array of the hash
	passwordBytes, err := encryptPassword(password)
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't encrypt password")
		return err
	}

	// Insert the password into the database
	_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "passwords" ("username", "password") VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password`, username, string(passwordBytes))
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't execute statement")
		return err
	}
	return nil
}

// CheckPassword checks if a password is valid for username, password is passed in unencrypted.
// Returns a bool of if the password is correct or not.
func (db *DB) CheckPassword(ctx context.Context, username string, password string) bool {
	defer trace.StartRegion(ctx, "DB/CheckPassword").End()

	var encPasswd string
	row := db.sqldb.QueryRowContext(ctx, `select password from passwords where username=$1`, username)
	switch err := row.Scan(&encPasswd); err {
	case sql.ErrNoRows:
		// If there is no password then we can assume the user doesn't exist and the login is incorrect.
		return false
	case nil:
		// If there is no error then the SQL statement executed correctly
		// It returns true or false depending on if the password is correct or not.
		return bcrypt.CompareHashAndPassword([]byte(encPasswd), []byte(password)) == nil
	default:
		// There is a error that we didn't account for so we log it.
		log.Warn().Err(err).Msg("CheckPassword got unexpected error")
		return false
	}
}
