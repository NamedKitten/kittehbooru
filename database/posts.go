package database

import (
	"context"

	"runtime/trace"

	"github.com/NamedKitten/kittehimageboard/utils"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

func (db *DB) Post(ctx context.Context, postID int64) (types.Post, bool) {
	defer trace.StartRegion(ctx, "DB/Post").End()

	p := types.Post{}
	var tags string

	rows, err := db.sqldb.QueryContext(ctx, `select "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype" from posts where postID = $1`, postID)
	if err != nil {
		log.Error().Err(err).Msg("Post can't query")
		return p, false
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&p.Filename, &p.FileExtension, &p.Description, &tags, &p.Poster, &p.CreatedAt, &p.MimeType)
		if err != nil {
			log.Error().Err(err).Msg("Post can't scan")
			return p, false
		} else {
			p.PostID = postID
			p.Tags = utils.SplitTagsString(tags)
			return p, true
		}
	}
	return p, false
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DB) AddPost(ctx context.Context, post types.Post) error {
	defer trace.StartRegion(ctx, "DB/AddPost").End()

	_, err := db.sqldb.ExecContext(ctx, `INSERT INTO "posts"("postid", "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, post.PostID, post.Filename, post.FileExtension, post.Description, utils.TagsListToString(post.Tags), post.Poster, post.CreatedAt, post.MimeType)
	if err != nil {
		log.Warn().Err(err).Msg("AddPost can't execute insert post statement")
		return err
	}

	err = db.AddPostTags(ctx, post)

	return err
}

func (db *DB) EditPost(ctx context.Context, postID int64, post types.Post) error {
	defer trace.StartRegion(ctx, "DB/EditPost").End()

	err := db.DeletePost(ctx, postID)
	if err != nil {
		return err
	}
	err = db.AddPost(ctx, post)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) DeletePost(ctx context.Context, postID int64) error {
	defer trace.StartRegion(ctx, "DB/DeletePost").End()

	p, _ := db.Post(ctx, postID)
	db.RemovePostTags(ctx, p)

	_, err := db.sqldb.ExecContext(ctx, `delete from posts where postid = $1`, postID)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't execute delete post statement")
		return err
	}
	return nil
}

func (db *DB) AllPostIDs(ctx context.Context) ([]int64, error) {
	defer trace.StartRegion(ctx, "DB/AllPostIDs").End()
	posts := make([]int64, 0)

	rows, err := db.sqldb.QueryContext(ctx, `select "postid" from posts where true`)
	if err != nil {
		log.Error().Err(err).Msg("AllPostIDs can't query wildcard posts")
		return posts, err
	}
	defer rows.Close()
	var pid int64
	for rows.Next() {
		err = rows.Scan(&pid)
		if err != nil {
			log.Error().Err(err).Msg("AllPostIDs can't scan row")
			return posts, err
		}
		posts = append(posts, pid)
	}
	return posts, nil
}
