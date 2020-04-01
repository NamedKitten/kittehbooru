package database

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	"runtime/trace"
	"strconv"

	"github.com/NamedKitten/kittehbooru/utils"

	"github.com/NamedKitten/kittehbooru/types"
	"github.com/rs/zerolog/log"
)

// A error to be returned if the post does not exist.
var PostNotExistError = errors.New("Post does not exist")

// Post fetches a post from the database given a post ID
func (db *DB) Post(ctx context.Context, postID int64) (p types.Post, err error) {
	defer trace.StartRegion(ctx, "DB/Post").End()

	if postID == 0 {
		return types.Post{}, errors.New("Invalid ID")
	}

	var tags string

	// Query for the post
	err = db.sqldb.QueryRowContext(ctx, `select "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype" from posts where postID = $1`, postID).Scan(&p.Filename, &p.FileExtension, &p.Description, &tags, &p.Poster, &p.CreatedAt, &p.MimeType)
	if err != nil {
		log.Error().Err(err).Msg("Post can't select")
		return
	}
	// Set the postID and split the tags string, returning now filled in Post
	p.PostID = postID
	p.Tags = utils.SplitTagsString(tags)
	return
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DB) AddPost(ctx context.Context, post types.Post) (err error) {
	defer trace.StartRegion(ctx, "DB/AddPost").End()

	post.Tags = db.filterTags(post.Tags)

	tagCountsCache.Delete(ctx, "*")
	for _, tag := range post.Tags {
		tagCountsCache.Delete(ctx, tag)
	}

	_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "posts"("postid", "filename", "ext", "description", "tags", "poster", "timestamp", "mimetype") VALUES ($1,$2,$3,$4,$5,$6,$7, $8)`, post.PostID, post.Filename, post.FileExtension, post.Description, utils.TagsListToString(post.Tags), post.Poster, post.CreatedAt, post.MimeType)
	if err != nil {
		log.Warn().Err(err).Msg("AddPost can't execute insert post statement")
		return
	}

	err = db.AddPostTags(ctx, post)
	return
}

// EditPost edits a post give a new post object and a post ID
func (db *DB) EditPost(ctx context.Context, postID int64, p types.Post) (err error) {
	defer trace.StartRegion(ctx, "DB/EditPost").End()

	p.Tags = db.filterTags(p.Tags)

	tags := utils.TagsListToString(p.Tags)
	_, err = db.sqldb.ExecContext(ctx, `update posts set "filename"=$1, "ext"=$2, "description"=$3, "tags"=$4, "poster"=$5, "timestamp"=$6, "mimetype"=$7 where postid = $8`, p.Filename, p.FileExtension, p.Description, tags, p.Poster, p.CreatedAt, p.MimeType, postID)
	if err != nil {
		log.Warn().Err(err).Msg("EditPost can't execute statement")
		return
	}

	err = db.AddPostTags(ctx, p)
	return
}

// DeletePost deletes a post from the database
// TODO: add parameter for knowing if it should delete the post's files or not when editing
func (db *DB) DeletePost(ctx context.Context, postID int64) (err error) {
	defer trace.StartRegion(ctx, "DB/DeletePost").End()

	var p types.Post
	p, err = db.Post(ctx, postID)
	if err != nil {
		return
	}
	err = db.RemovePostTags(ctx, p.PostID)
	if err != nil {
		return
	}

	_, err = db.sqldb.ExecContext(ctx, `delete from posts where postid = $1`, postID)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't execute delete post statement")
		return
	}
	db.ContentStorage.Delete(fmt.Sprintf("%d.%s", postID, p.FileExtension))
	db.ContentStorage.Delete(fmt.Sprintf("%d.webp", postID))
	fmt.Sprintf("%d.webp", postID)
	return
}

// AllPostIDs returns a list of the ID of all the posts in the database.
func (db *DB) AllPostIDs(ctx context.Context) ([]int64, error) {
	defer trace.StartRegion(ctx, "DB/AllPostIDs").End()
	posts := make([]int64, 0)

	val, ok := searchCache.Get(ctx, "*")
	if ok {
		return val.([]int64), nil
	}

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

	searchCache.Add(ctx, "*", posts, 0)
	return posts, nil
}

// PostsTagsCounts returns a map of tag to how many time the tag was encountered in all posts
func (db *DB) PostsTagsCounts(ctx context.Context, posts []int64) (res map[string]int, err error) {
	defer trace.StartRegion(ctx, "DB/PostsTagsCounts").End()

	res = make(map[string]int)
	stmt, err := db.sqldb.PrepareContext(ctx, `select "tags" from posts where postID = $1`)
	defer stmt.Close()



	for _, p := range posts {
		var tags string
		var tagsSlice []string

		if val, ok := postTagsCache.Get(ctx, strconv.Itoa(int(p))); ok {
			for _, tag := range val.([]string) {
				if i, ok := res[tag]; ok {
					res[tag] = i + 1
				} else {
					res[tag] = 1
				}
			}
			continue
		}

		err = stmt.QueryRowContext(ctx, p).Scan(&tags)
		switch {
		case err == sql.ErrNoRows:
			return
		case err != nil:
			log.Fatal().Err(err)
		default:
			tagsSlice = utils.SplitTagsString(tags)
			postTagsCache.Set(ctx, strconv.Itoa(int(p)), tagsSlice, 0)
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

// Posts returns a list of posts from their post IDs, does the same as Post but
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
