package database

import (
	"github.com/rs/zerolog/log"
)

// sqlInit initialises the database and makes sure all tables exist.
// Also provides migration between schema versions.
func (db *DB) sqlInit() {
	var err error
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "users" (  "avatarID"  bigint,  "owner"  BOOL,  "admin"  BOOL,  "username"  TEXT,  "description"  TEXT,  PRIMARY KEY("username"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Users Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "passwords" (  "username"  TEXT, "password"  TEXT,  PRIMARY KEY("username"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Passwords Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "tags" (  "tag"  TEXT, "posts"  TEXT,  PRIMARY KEY("tag"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Tags Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "posts" (  "postid" bigint, "filename"  TEXT, "ext" TEXT, "description" TEXT, "tags"  TEXT, "poster" TEXT, "timestamp" bigint, "mimetype" TEXT, PRIMARY KEY("postid"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Posts Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "sessions" (  "token" TEXT, "username" TEXT, "expiry" bigint, PRIMARY KEY("token"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Sessions Table")
	}
	// Delete sha256 column if exists from past versions.
	db.sqldb.Exec(`ALTER TABLE "posts" DROP COLUMN sha256`)
}