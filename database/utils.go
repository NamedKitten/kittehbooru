package database

import (
	"context"
	"math"
	"net/http"
	"runtime/trace"
	"sort"
	"strings"

	"github.com/NamedKitten/kittehimageboard/types"
)

// numOfPostsForTags returns the total number of posts for a list of tags.
func (db *DB) NumOfPostsForTags(ctx context.Context, searchTags []string) int {
	return len(db.cacheSearch(ctx, searchTags))
}

// numOfPagesForTags returns the total number of pages for a list of tags.
func (db *DB) NumOfPagesForTags(ctx context.Context, searchTags []string) int {
	// returns the smallest integer value greater than or equal to numPosts / 20
	// TODO: make variable page size
	return int(math.Ceil(float64(db.NumOfPostsForTags(ctx, searchTags)) / float64(20)))
}

// filterTags filters tags before searching using them
// Order of operations:
// 1. Remove duplicate tags
// 2. Removes both a positive and a negative tag if they are the same.
// 3. Adds * (wildcard operator) if there is only negative matches.
// 4. Sorts so positive tags come before negative tags.
func (db *DB) filterTags(tags []string) []string {
	// 1. Remove duplicate tags
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
		// 2. Removes both a positive and a negative tag if they are the same.
		if !ok && !(tag == " " || len(tag) == 0) {
			if !is {
				isOnlyNegatives = false
			}
			tags = append(tags, tag)
		}
	}

	// 3. Adds * (wildcard operator) if there is only negative matches.
	if isOnlyNegatives {
		tags = append(tags, "*")
	}

	// 4. Sorts so positive tags come before negative tags.
	// This is so we can fail early if one of the positive tags doesn't have any results.
	sort.Slice(tags, func(i, j int) bool {
		return !strings.HasPrefix(tags[i], "-") && strings.HasPrefix(tags[j], "-")
	})

	return tags
}

// VerifyRecaptcha verifies a recaptcha response
func (db *DB) VerifyRecaptcha(ctx context.Context, resp string) bool {
	defer trace.StartRegion(ctx, "DB/VerifyRecaptcha").End()

	// if we don't use reCaptcha, return true
	if !db.Settings.ReCaptcha {
		return true
	}

	// TODO: Add context support to recaptcha-go.
	// if there was a error verifying the response, return false
	if err := captcha.Verify(resp); err != nil {
		return false
	}

	// return true when the captcha is correct
	return true
}

// CheckForLoggedIntypes.User is a helper function that is used to see if a
// HTTP request is from a logged in user.
// It returns a types.User struct and a bool to tell if there was a logged in
// user or not.
func (db *DB) CheckForLoggedInUser(ctx context.Context, r *http.Request) (types.User, bool) {
	defer trace.StartRegion(ctx, "DB/CheckForLoggedInUser").End()

	// if there is no error in fetching the session's token, continue
	c, err := r.Cookie("sessionToken")
	if err == nil {
		// if there is no error checking the token, continue
		if sess, err := db.CheckToken(ctx, c.Value); err == nil {
			// fetch the user for the session
			u, err := db.User(ctx, sess.Username)
			if err == nil {
				// return the user
				return u, true
			}
		}
	}
	return types.User{}, false
}
