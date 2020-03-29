package database

import (
	"context"
	"encoding/json"

	"runtime/trace"

	"github.com/NamedKitten/kittehbooru/types"
	"github.com/rs/zerolog/log"
)

// AddUser adds a user to the database.
func (db *DB) AddUser(ctx context.Context, u types.User) {
	defer trace.StartRegion(ctx, "DB/AddUser").End()

	_, err := db.sqldb.ExecContext(ctx, `INSERT INTO "users"("avatarID","owner","admin","username","description","theme") VALUES ($1,$2,$3,$4,$5,$6)`, u.AvatarID, u.Owner, u.Admin, u.Username, "", "dark")
	if err != nil {
		log.Warn().Err(err).Msg("AddUser can't execute statement")
	}
}

// User fetches a user from the database.
func (db *DB) User(ctx context.Context, username string) (u types.User, err error) {
	defer trace.StartRegion(ctx, "DB/User").End()

	if val, ok := userCache.Get(ctx, username); ok {
		u = val.(types.User)
	} else {
		err = db.sqldb.QueryRowContext(ctx, `select "avatarID","owner","admin","username","description", "theme" from users where username = $1`, username).Scan(&u.AvatarID, &u.Owner, &u.Admin, &u.Username, &u.Description, &u.Theme)
		if err != nil {
			log.Error().Err(err).Msg("User can't query statement")
		} else {
			userCache.Add(ctx, u.Username, u, 0)
		}
	}
	return u, err
}

// EditUser edits a user in the database.
func (db *DB) EditUser(ctx context.Context, u types.User) (err error) {
	defer trace.StartRegion(ctx, "DB/EditUser").End()
	userCache.Delete(ctx, u.Username)
	_, err = db.sqldb.ExecContext(ctx, `update users set "avatarID"=$1, owner=$2, admin=$3, description=$4, theme=$5 where username = $6`, u.AvatarID, u.Owner, u.Admin, u.Description, u.Theme, u.Username)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser can't execute statement")
		return err
	}
	userCache.Delete(ctx, u.Username)
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
