package database

import (
	"context"
	"database/sql"

	"fmt"
	"runtime/trace"
	"strconv"
	"strings"

	"github.com/NamedKitten/kittehbooru/types"
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

func removeWildcard(s []string) []string {
	tags := make([]string, 0)
	for _, tag := range s {
		if tag != "*" {
			tags = append(tags, tag)
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
	db.RemovePostTags(ctx, post.PostID)

	tx, err := db.sqldb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO "tagMap"(postid, tag) VALUES( $1, $2 )`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, tag := range post.Tags {
		if _, err := stmt.Exec(post.PostID, tag); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// TagPosts returns a list of post IDs for a tag
func (db *DB) TagPosts(ctx context.Context, tag string) (posts []int64, err error) {
	defer trace.StartRegion(ctx, "DB/TagPosts").End()

	x, err := db.TagsPosts(ctx, []string{tag})
	if err == nil {
		return x[tag], err
	}
	return
}

// PostTags returns a list of tags for a post
func (db *DB) PostTags(ctx context.Context, pid int64) (tagsSlice []string, err error) {
	defer trace.StartRegion(ctx, "DB/PostTags").End()

	stmt, err := db.sqldb.PrepareContext(ctx, `select "tag" from "tagMap" where postid = $1`)
	tagsSlice = make([]string, 0)
	var tag string

	err = stmt.QueryRowContext(ctx, pid).Scan(&tag)
	switch {
	case err == sql.ErrNoRows:
		return
	case err != nil:
		log.Fatal().Err(err)
	default:
		tagsSlice = append(tagsSlice, tag)
	}

	return
}


// RemovePostTags removes all instances of a post from their tags
func (db *DB) RemovePostTags(ctx context.Context, postID int64) (err error) {
	defer trace.StartRegion(ctx, "DB/RemovePostTags").End()
	_, err = db.sqldb.ExecContext(ctx, `DELETE FROM "tagMap" WHERE postid = $1`, postID)
	if err != nil {
		log.Warn().Err(err).Msg("RemovePostTags can't execute delete statement")
		return
	}
	return
}

// TagPosts returns a map of tags for posts to their post IDs
func (db *DB) TagsPosts(ctx context.Context, tags []string) (result map[string][]int64, err error) {
	defer trace.StartRegion(ctx, "DB/TagsPosts").End()
	result = make(map[string][]int64)

	tags = removeWildcard(tags)

	posTags := make([]string, 0)
	negTags := make([]string, 0)
	args := make([]interface{}, 0)

	for _, tag := range tags {
		if strings.HasPrefix(tag, "-") {
			negTags = append(negTags, tag[1:])
		} else {
			posTags = append(posTags, tag)
		}
	}

	argsI := 1
	posArgs := make([]string, 0)
	negArgs := make([]string, 0)

	for _, tag := range posTags {
		posArgs = append(posArgs, "$"+strconv.Itoa(argsI))
		args = append(args, tag)
		argsI = argsI + 1
	}

	for _, tag := range negTags {
		negArgs = append(negArgs, "$"+strconv.Itoa(argsI))
		args = append(args, tag)
		argsI = argsI + 1
	}

	posArgsStr := strings.Join(posArgs, ",")
	negArgsStr := strings.Join(negArgs, ",")

	sStart := `SELECT postID, tag FROM "tagMap" WHERE`
	cond := ""
	sEnd := `GROUP BY postID, tag`

	if len(negTags) == 0 && len(posTags) == 0 {
		cond = "true"
	} else if len(posTags) != 0 && len(negTags) != 0 {
		cond = fmt.Sprintf(`(tag IN (%s)) AND postID NOT IN (SELECT postID FROM "tagMap" WHERE (tag IN (%s)))`, posArgsStr, negArgsStr)
		sEnd = sEnd + fmt.Sprintf(` HAVING COUNT( postID ) =%d`, len(posTags))
	} else if len(negTags) != 0 {
		cond = fmt.Sprintf(`NOT postID IN (SELECT postID FROM "tagMap" WHERE (tag IN (%s)))`, negArgsStr)
	} else if len(posTags) != 0 {
		sEnd = sEnd + fmt.Sprintf(` HAVING COUNT( postID ) =%d`, len(posTags))
		cond = fmt.Sprintf(`(tag IN (%s))`, posArgsStr)
	}

	s := fmt.Sprintf("%s %s %s", sStart, cond, sEnd)

	rows, err := db.sqldb.QueryContext(ctx, s, args...)
	if err != nil {
		log.Error().Err(err).Msg("TagsPosts can't query posts")
	}
	defer rows.Close()

	var pid int64
	var tag string
	for rows.Next() {
		err = rows.Scan(&pid, &tag)
		if err != nil {
			log.Error().Err(err).Msg("TagsPosts can't scan row")
			return
		}

		if result[tag] == nil {
			result[tag] = make([]int64, 0)
		}
		result[tag] = append(result[tag], pid)

	}

	return
}
