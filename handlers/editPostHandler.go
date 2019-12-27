package handlers

import (
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

// EditPostHandler is the endpoint used to edit posts.
func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	postID, _ := strconv.Atoi(vars["postID"])

	post, postExists := DB.Posts[int64(postID)]
	if !postExists {
		http.Redirect(w, r, "/", 302)
		return
	}

	if !(user.Admin || post.Poster == user.Username) {
		http.Redirect(w, r, "/", 302)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	tags := utils.SplitTagsString(r.PostFormValue("tags"))

	newTags := make([]string, 0)
	for _, tag := range tags {
		hasPermissionToUseTag := true
		if v, ok := DB.LockedTags[tag]; ok {
			if v != user.Username {
				hasPermissionToUseTag = false
			}
		}

		if !strings.HasPrefix(tag, "user:") && hasPermissionToUseTag {
			newTags = append(newTags, tag)
		}
	}

	newTags = append(newTags, "user:"+user.Username)

	post.Tags = newTags
	post.Description = r.PostFormValue("description")

	DB.Posts[int64(postID)] = post

	http.Redirect(w, r, "/view/"+vars["postID"], 302)
}
