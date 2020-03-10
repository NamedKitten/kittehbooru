package database

import (
	"database/sql"

	"github.com/rs/zerolog/log"
)

func (db *DB) sqlInit(sqldb *sql.DB, dbType string) {
	var err error
	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS "users" (  "avatarID"  bigint,  "owner"  BOOL,  "admin"  BOOL,  "username"  TEXT,  "description"  TEXT,  PRIMARY KEY("username"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Users Table")
	}
	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS "passwords" (  "username"  TEXT, "password"  TEXT,  PRIMARY KEY("username"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Passwords Table")
	}
	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS "tags" (  "tag"  TEXT, "posts"  TEXT,  PRIMARY KEY("tag"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Tags Table")
	}
	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS "posts" (  "postid" bigint, "filename"  TEXT, "ext" TEXT, "description" TEXT, "tags"  TEXT, "poster" TEXT, "timestamp" bigint, "mimetype" TEXT, PRIMARY KEY("postid"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Posts Table")
	}
	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS "sessions" (  "token" TEXT, "username" TEXT, "expiry" bigint, PRIMARY KEY("token"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Sessions Table")
	}
	// Delete sha256 column if exists from past versions.
	sqldb.Exec(`ALTER TABLE "posts" DROP COLUMN sha256`)
}
