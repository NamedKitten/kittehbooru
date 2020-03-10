package database

import (
	"context"
	"math"
	"net/http"
	"runtime/trace"
	"strings"

	"github.com/NamedKitten/kittehimageboard/types"
)

// numOfPostsForTags returns the total number of posts for a list of tags.
func (db *DB) NumOfPostsForTags(ctx context.Context, searchTags []string) int {
	return len(db.cacheSearch(ctx, searchTags))
}

// numOfPagesForTags returns the total number of pages for a list of tags.
func (db *DB) NumOfPagesForTags(ctx context.Context, searchTags []string) int {
	return int(math.Ceil(float64(db.NumOfPostsForTags(ctx, searchTags)) / float64(20)))
}

func (db *DB) filterTags(tags []string) []string {
	// lets first remove any duplicate tags
	tempTags := make(map[string]bool)
	// this will remove duplicate entrys
	for _, tag := range tags {
		tempTags[tag] = true
	}

	isOnlyNegatives := true

	tags = make([]string, 0)
	for tag := range tempTags {
		// if there is a tag "foo" and also a tag "-foo", remove both of them to reduce database load
		is := strings.HasPrefix(tag, "-")
		var ok bool
		if is {
			_, ok = tempTags[tag[1:]]
		} else {
			_, ok = tempTags["-"+tag]
		}
		if !ok && !(tag == " " || len(tag) == 0) {
			if !is {
				isOnlyNegatives = false
			}
			tags = append(tags, tag)
		}
	}

	// if there is only negative tags, add wildcard
	if isOnlyNegatives {
		tags = append(tags, "*")
	}
	return tags
}

func (db *DB) VerifyRecaptcha(ctx context.Context, resp string) bool {
	defer trace.StartRegion(ctx, "DB/VerifyRecaptcha").End()

	if !db.Settings.ReCaptcha {
		return true
	}
	// TODO: Add context support to recaptcha-go.
	if err := captcha.Verify(resp); err != nil {
		return false
	}
	return true
}

// CheckForLoggedIntypes.User is a helper function that is used to see if a
// HTTP request is from a logged in user.
// It returns a types.User struct and a bool to tell if there was a logged in
// user or not.
func (db *DB) CheckForLoggedInUser(ctx context.Context, r *http.Request) (types.User, bool) {
	defer trace.StartRegion(ctx, "DB/CheckForLoggedInUser").End()

	c, err := r.Cookie("sessionToken")
	if err == nil {
		if sess, ok := db.Sessions.CheckToken(ctx, c.Value); ok {
			u, exists := db.User(ctx, sess.Username)
			if exists {
				return u, true
			}
		}
	}
	return types.User{}, false
}
