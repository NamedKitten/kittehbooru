package database

import (
	"context"
	"encoding/json"

	"runtime/trace"
	"strings"
	"database/sql"
	"github.com/NamedKitten/kittehimageboard/utils"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

func sliceContains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func removeWildcardAndNegatives(s []string) []string {
	tags := make([]string, 0)
    for _, tag := range s {
		if tag != "*" {
			if strings.HasPrefix(tag, "-") {
				tags = append(tags, tag[1:])
			} else {
				tags = append(tags, tag)
			}
		}
    }
    return tags
}

func removeItem(s []string, e string) []string {
	tags := make([]string, 0)
    for _, a := range s {
        if a != e {
            tags = append(tags, a)
        }
    }
    return tags
}

// AddPostTags adds a post's tags to the database for easy searching.
func (db *DB) AddPostTags(ctx context.Context, post types.Post) error {
	defer trace.StartRegion(ctx, "DB/addPostTags").End()
	for _, tag := range post.Tags {
		posts, err := db.TagPosts(ctx, tag)
		// If it doesn't exist in the database, make it with just the post ID
		// otherise append it to existing result
		if err != nil {
			posts = []int64{post.PostID}
		} else {
			posts = append(posts, post.PostID)
		}
		db.SetTagPosts(ctx, tag, posts)
	}
	return nil
}

// SetTagPosts performs the actual database operation of setting the tags.
func (db *DB) SetTagPosts(ctx context.Context, tag string, posts []int64) error {
	defer trace.StartRegion(ctx, "DB/SetTagPosts").End()
	x, err := json.Marshal(posts)
	if err != nil {
		log.Error().Err(err).Msg("SetTagPosts can't unmarshal posts list")
		return err
	}
	_, err = db.sqldb.ExecContext(ctx, `INSERT INTO "tags"("tag", "posts") VALUES ($1, $2) ON CONFLICT (tag) DO UPDATE SET posts = EXCLUDED.posts`, tag, string(x))
	if err != nil {
		log.Warn().Err(err).Msg("SetTagPosts can't execute insert tags statement")
		return err
	}
	return nil
}

// TagPosts returns a list of post IDs for a tag
func (db *DB) TagPosts(ctx context.Context, tag string) (posts []int64, err error) {
	defer trace.StartRegion(ctx, "DB/TagPosts").End()

	var postsString string

	err = db.sqldb.QueryRowContext(ctx, `select "posts" from tags where tag = $1`, tag).Scan(&postsString)
	if err != nil {
		if err == sql.ErrNoRows {
			return []int64{}, err
		}
		log.Error().Err(err).Msg("TagPosts can't select tags")
		return posts, err
	}

	err = json.Unmarshal([]byte(postsString), &posts)
	if err != nil {
		log.Error().Err(err).Msg("TagPosts Json Unmarshal Error")
		return posts, err
	}

	return posts, err
}

// RemovePostTags removes all instances of a post from their tags
func (db *DB) RemovePostTags(ctx context.Context, p types.Post) error {
	defer trace.StartRegion(ctx, "DB/RemovePostTags").End()
	for _, tag := range p.Tags {
		posts, err := db.TagPosts(ctx, tag)
		if err != nil {
			return err
		}
		posts = utils.RemoveFromSlice(posts, p.PostID)
		err = db.SetTagPosts(ctx, tag, posts)
		if err != nil {
			return err
		}
	}
	return nil
}

// TagPosts returns a map of tags to their of post IDs
func (db *DB) TagsPosts(ctx context.Context, tags []string) (result map[string][]int64, err error) {
	defer trace.StartRegion(ctx, "DB/TagsPosts").End()
	result = make(map[string][]int64)

	if sliceContains(tags, "*") {
		region := trace.StartRegion(ctx, "DB/TagsPosts/wildcard")
		var wildcardPosts []int64
		wildcardPosts, err = db.AllPostIDs(ctx)
		if err != nil {
			return
		}
		result["*"] = wildcardPosts
		region.End()
	}

	tags = removeWildcardAndNegatives(tags)

	cachedItems := make([]string, 0)

	for _, tag := range tags {
		val, ok := db.SearchCache.Get(ctx, tag)
		if ok {
			cachedItems = append(cachedItems, tag)
			result[tag] = val
		}
	}

	for _, tag := range cachedItems {
		tags = removeItem(tags, tag)
	}


	defer trace.StartRegion(ctx, "DB/TagsPosts/tags").End()
	stmt, err := db.sqldb.PrepareContext(ctx, `select "posts" from tags where tag = $1`)
	for _, tag := range tags {
		var postsString string
		var posts []int64
		err = stmt.QueryRowContext(ctx, tag).Scan(&postsString)
		switch {
		case err == sql.ErrNoRows:
			continue
		case err != nil:
			log.Fatal().Err(err).Msg("TagsPosts weird error")
		default:
			err = json.Unmarshal([]byte(postsString), &posts)
			if err != nil {
				log.Error().Err(err).Msg("TagsPosts Json Unmarshal Error")
				return
			}
			result[tag] = posts
		}
	}

	for _, tag := range tags {
		db.SearchCache.Add(ctx, tag, result[tag])
	}

	return
}