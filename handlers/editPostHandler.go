package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// EditPostHandler is the endpoint used to edit posts.
func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		log.Error().Err(err).Msg("Can't convert postID to string")
		return
	}
	post, postExists := DB.Post(int64(postID))
	if !postExists {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if !(user.Admin || post.Poster == user.Username) {
		http.Redirect(w, r, "/", http.StatusFound)
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
		if !strings.HasPrefix(tag, "user:") {
			newTags = append(newTags, tag)
		}
	}

	newTags = append(newTags, "user:"+user.Username)

	post.Tags = newTags
	post.Description = r.PostFormValue("description")

	DB.EditPost(int64(postID), post)

	http.Redirect(w, r, "/view/"+vars["postID"], http.StatusFound)
}
