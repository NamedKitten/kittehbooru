package database

import (
	"context"
	"encoding/json"

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

	for _, tag := range post.Tags {
		rows, err := db.sqldb.QueryContext(ctx, `select "posts" from tags where tag = $1`, tag)
		if err != nil {
			log.Error().Err(err).Msg("AddPost can't select tags")
			return err
		}
		defer rows.Close()

		var postsString string
		for rows.Next() {
			err = rows.Scan(&postsString)
			if err != nil {
				log.Error().Err(err).Msg("AddPost can't scan row")
			}
		}

		var posts []int64
		err = json.Unmarshal([]byte(postsString), &posts)
		if err != nil {
			posts = []int64{post.PostID}
		} else {
			posts = append(posts, post.PostID)
		}
		x, err := json.Marshal(posts)
		if err != nil {
			log.Error().Err(err).Msg("AddPost can't marshal posts list")
			return err
		}

		_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "tags"("tag", "posts") VALUES ($1, $2) ON CONFLICT (tag) DO UPDATE SET posts = EXCLUDED.posts`, tag, string(x))
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't execute insert tags statement")
			return err
		}
	}

	return nil
}

func (db *DB) EditPost(ctx context.Context, postID int64, post types.Post) {
	defer trace.StartRegion(ctx, "DB/EditPost").End()

	err := db.DeletePost(ctx, postID)
	if err != nil {
		log.Error().Err(err).Msg("EditPost can't delete post")
		return
	}
	err = db.AddPost(ctx, post)
	if err != nil {
		log.Error().Err(err).Msg("EditPost can't add post")
		return
	}
}

func (db *DB) DeletePost(ctx context.Context, postID int64) error {
	defer trace.StartRegion(ctx, "DB/DeletePost").End()

	p, _ := db.Post(ctx, postID)
	for _, tag := range p.Tags {
		rows, err := db.sqldb.QueryContext(ctx, `select "posts" from tags where tag = $1`, tag)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost can't select tags")
			return err
		}
		defer rows.Close()

		var posts []int64
		newPosts := make([]int64, 0)
		var postsString string

		for rows.Next() {
			err = rows.Scan(&postsString)
			if err != nil {
				log.Error().Err(err).Msg("DeletePost can't scan rows")
				return err
			}
		}
		err = json.Unmarshal([]byte(postsString), &posts)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost Json Unmarshal Error")
			return err
		}

		posts = utils.RemoveFromSlice(posts, postID)
		x, err := json.Marshal(newPosts)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost can't unmarshal posts list")
			return err
		}

		_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "tags"("tag", "posts") VALUES ($1, $2) ON CONFLICT (tag) DO UPDATE SET posts = EXCLUDED.posts`, tag, string(x))
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't execute insert tags statement")
			return err
		}
	}

	_, err := db.sqldb.ExecContext(ctx, `delete from posts where postid = $1`, postID)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't execute delete post statement")
		return err
	}
	return nil
}
