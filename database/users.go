package database

import (
	"context"
	"encoding/json"

	"runtime/trace"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

// AddUser adds a user to the database.
func (db *DB) AddUser(ctx context.Context, u types.User) {
	defer trace.StartRegion(ctx, "DB/AddUser").End()

	_, err := db.sqldb.ExecContext(ctx, `INSERT INTO "users"("avatarID","owner","admin","username","description") VALUES ($1,$2,$3,$4,$5)`, u.AvatarID, u.Owner, u.Admin, u.Username, "")
	if err != nil {
		log.Warn().Err(err).Msg("AddUser can't execute statement")
	}
}

// User fetches a user from the database.
func (db *DB) User(ctx context.Context, username string) (u types.User, err error) {
	defer trace.StartRegion(ctx, "DB/User").End()

	err = db.sqldb.QueryRowContext(ctx, `select "avatarID","owner","admin","username","description" from users where username = $1`, username).Scan(&u.AvatarID, &u.Owner, &u.Admin, &u.Username, &u.Description)
	if err != nil {
		log.Error().Err(err).Msg("User can't query statement")
	}
	return u, err
}

// EditUser edits a user in the database.
func (db *DB) EditUser(ctx context.Context, u types.User) (err error) {
	defer trace.StartRegion(ctx, "DB/EditUser").End()

	_, err = db.sqldb.ExecContext(ctx, `update users set "avatarID"=$1, owner=$2, admin=$3, description=$4 where username = $5`, u.AvatarID, u.Owner, u.Admin, u.Description, u.Username)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser can't execute statement")
		return err
	}
	return nil
}

// DeleteUser deletes a user and all their posts from the database
func (db *DB) DeleteUser(ctx context.Context, username string) error {
	defer trace.StartRegion(ctx, "DB/DeleteUser").End()

	_, err := db.sqldb.ExecContext(ctx, `delete from users where username = $1`, username)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't execute delete user statement")
		return err
	}

	_, err = db.sqldb.ExecContext(ctx, `delete from passwords where username = $1`, username)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't execute delete password statement")
		return err
	}

	rows, err := db.sqldb.QueryContext(ctx, `select "postid" from posts where poster = $1`, username)
	if err != nil {
		log.Error().Err(err).Msg("DeleteUser can't select posts")
	}
	defer rows.Close()

	var posts []int64
	var postsString string

	for rows.Next() {
		err = rows.Scan(&postsString)
		if err != nil {
			log.Error().Err(err).Msg("DeleteUser can't scan row")
			return err
		}
	}

	err = json.Unmarshal([]byte(postsString), &posts)
	if err != nil {
		log.Error().Err(err).Msg("Can't unmarshal posts list")
		return err
	}
	for _, post := range posts {
		err = db.DeletePost(ctx, post)
		if err != nil {
			log.Error().Err(err).Msg("Can't delete user's post")
			return err
		}
	}

	db.InvalidateSession(ctx, username)

	return nil
}
