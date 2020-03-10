package database

import (
	"context"
	"encoding/json"

	"runtime/trace"

	"github.com/NamedKitten/kittehimageboard/utils"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

func (db *DB) AddPostTags(ctx context.Context, post types.Post) error {
	defer trace.StartRegion(ctx, "DB/AddPostTags").End()
	for _, tag := range post.Tags {
		posts, err := db.TagPosts(ctx, tag)
		if err != nil {
			posts = []int64{post.PostID}
		} else {
			posts = append(posts, post.PostID)
		}

		db.SetTagPosts(ctx, tag, posts)
	}

	return nil
}

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

func (db *DB) TagPosts(ctx context.Context, tag string) ([]int64, error) {
	defer trace.StartRegion(ctx, "DB/TagPosts").End()

	rows, err := db.sqldb.QueryContext(ctx, `select "posts" from tags where tag = $1`, tag)
	if err != nil {
		log.Error().Err(err).Msg("TagPosts can't select tags")
		return []int64{}, err
	}
	defer rows.Close()

	var posts []int64
	var postsString string

	for rows.Next() {
		err = rows.Scan(&postsString)
		if err != nil {
			log.Error().Err(err).Msg("TagPosts can't scan rows")
			return []int64{}, err
		}
	}
	err = json.Unmarshal([]byte(postsString), &posts)
	if err != nil {
		log.Error().Err(err).Msg("TagPosts Json Unmarshal Error")
		return []int64{}, err
	}
	return posts, nil

}



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