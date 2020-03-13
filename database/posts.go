package database

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	"runtime/trace"

	"github.com/NamedKitten/kittehimageboard/utils"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

// A error to be returned if the post does not exist.
var PostNotExistError = errors.New("Post does not exist")

// Post fetches a post from the database given a post ID
func (db *DB) Post(ctx context.Context, postID int64) (types.Post, error) {
	defer trace.StartRegion(ctx, "DB/Post").End()

	p := types.Post{}
	var tags string

	rows, err := db.sqldb.QueryContext(ctx, `select "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype" from posts where postID = $1`, postID)
	if err != nil {
		log.Error().Err(err).Msg("Post can't query")
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&p.Filename, &p.FileExtension, &p.Description, &tags, &p.Poster, &p.CreatedAt, &p.MimeType)
		if err != nil {
			log.Error().Err(err).Msg("Post can't scan")
			return p, err
		} else {
			p.PostID = postID
			p.Tags = utils.SplitTagsString(tags)
			return p, nil
		}
	}
	return p, PostNotExistError
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DB) AddPost(ctx context.Context, post types.Post) error {
	defer trace.StartRegion(ctx, "DB/AddPost").End()

	_, err := db.sqldb.ExecContext(ctx, `INSERT INTO "posts"("postid", "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype") VALUES ($1,$2,$3,$4,$5,$6,$7, $8)`, post.PostID, post.Filename, post.FileExtension, post.Description, utils.TagsListToString(post.Tags), post.Poster, post.CreatedAt, post.MimeType)
	if err != nil {
		log.Warn().Err(err).Msg("AddPost can't execute insert post statement")
		return err
	}

	err = db.AddPostTags(ctx, post)
	return err
}

// EditPost edits a post give a new post object and a post ID
func (db *DB) EditPost(ctx context.Context, postID int64, p types.Post) error {
	defer trace.StartRegion(ctx, "DB/EditPost").End()

	var err error
	post, _ := db.Post(ctx, postID)
	db.RemovePostTags(ctx, post)

	tags := utils.TagsListToString(p.Tags)
	_, err = db.sqldb.ExecContext(ctx, `update posts set "filename"=$1, "ext"=$2, "description"=$3, "tags"=$4, "poster"=$5, "timestamp"=$6, "mimetype"=$7 where postid = $8`, p.Filename, p.FileExtension, p.Description, tags, p.Poster, p.CreatedAt, p.MimeType, postID)
	if err != nil {
		log.Warn().Err(err).Msg("EditPost can't execute statement")
		return err
	}

	err = db.AddPostTags(ctx, p)
	return err
}

// DeletePost deletes a post from the database
// TODO: add parameter for knowing if it should delete the post's files or not when editing
func (db *DB) DeletePost(ctx context.Context, postID int64) error {
	defer trace.StartRegion(ctx, "DB/DeletePost").End()

	p, _ := db.Post(ctx, postID)
	db.RemovePostTags(ctx, p)

	_, err := db.sqldb.ExecContext(ctx, `delete from posts where postid = $1`, postID)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't execute delete post statement")
		return err
	}
	db.ContentStorage.Delete(fmt.Sprintf("%d.%s", postID, p.FileExtension))
	db.ContentStorage.Delete(fmt.Sprintf("%d.webp", postID))
	fmt.Sprintf("%d.webp", postID)

	return nil
}

// AllPostIDs returns a list of the ID of all the posts in the database.
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

// PostsTagsCounts returns a map of tag to how many time the tag was encountered in all posts
func (db *DB) PostsTagsCounts(ctx context.Context, posts []int64) (res map[string]int, err error) {
	defer trace.StartRegion(ctx, "DB/PostsTagsCounts").End()

	res = make(map[string]int)
	stmt, err := db.sqldb.PrepareContext(ctx, `select "tags" from posts where postID = $1`)
	defer stmt.Close()

	var tags string
	var tagsSlice []string

	for _, p := range posts {
		err = stmt.QueryRowContext(ctx, p).Scan(&tags)
		switch {
		case err == sql.ErrNoRows:
			continue
		case err != nil:
			log.Fatal().Err(err)
		default:
			tagsSlice = utils.SplitTagsString(tags)
			for _, tag := range tagsSlice {
				if i, ok := res[tag]; ok {
					res[tag] = i + 1
				} else {
					res[tag] = 1
				}
			}
		}
	}

	return res, nil
}

// PostsTagsCounts returns a list of posts from their post IDs, does the same as Post but
// uses a prepared statement to do it all in one db connection
func (db *DB) Posts(ctx context.Context, posts []int64) (res []types.Post, err error) {
	defer trace.StartRegion(ctx, "DB/Posts").End()

	res = make([]types.Post, 0)
	stmt, err := db.sqldb.PrepareContext(ctx, `select "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype" from posts where postID = $1`)
	defer stmt.Close()

	var tags string
	var p types.Post

	for _, pid := range posts {
		err = stmt.QueryRowContext(ctx, pid).Scan(&p.Filename, &p.FileExtension, &p.Description, &tags, &p.Poster, &p.CreatedAt, &p.MimeType)
		switch {
		case err == sql.ErrNoRows:
			continue
		case err != nil:
			log.Fatal().Err(err)
		default:
			p.PostID = pid
			p.Tags = utils.SplitTagsString(tags)
			res = append(res, p)
		}
	}

	return res, nil
}