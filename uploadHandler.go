package main

import (
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"github.com/h2non/filetype"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// maxUploadSize is the maximum filesize for a post.
// TODO: move to DB.Settings
// Default: 64Mb
const maxUploadSize = 64 * 1024 * 1024

// uploadHandler is the API endpoint for creating posts.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		log.Error("Not logged in.")
		http.Redirect(w, r, "/login", 302)
		return
	}
	log.Error(r.Body)
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Error("File too big.")
		renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		log.Error("Not logged in.")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error("Invalid file.")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	fileType, _ := filetype.Match(fileBytes)

	node, _ := snowflake.NewNode(1)
	postID := node.Generate()
	postIDInt64 := postID.Int64()
	fileName := strconv.Itoa(int(postIDInt64))
	tags := splitTagsString(r.PostFormValue("tags"))
	log.Error("Tags: ", tags)
	description := r.PostFormValue("description")

	newPath := filepath.Join("content/", fileName+"."+fileType.Extension)
	newFile, err := os.Create(newPath)
	if err != nil {
		log.Error("Cant create file.")
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		log.Error("Cant create file.")
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	newTags := make([]string, 0)

	for _, tag := range tags {
		hasPermissionToUseTag := true
		if v, ok := DB.LockedTags[tag]; ok {
			if v != user.ID {
				hasPermissionToUseTag = false
			}
		}

		if !strings.HasPrefix(tag, "user:") && hasPermissionToUseTag {
			newTags = append(newTags, tag)
		}
	}
	tags = newTags

	tags = append(tags, "user:"+user.Username)

	sha256sum := sha256Bytes(fileBytes)

	DB.AddPost(Post{
		PostID:        postIDInt64,
		Filename:      fileName,
		FileExtension: strings.TrimPrefix(fileType.Extension, "."),
		Tags:          tags,
		Description:   description,
		PosterID:      user.ID,
		CreatedAt:     postID.Time(),
		Sha256:        sha256sum,
		MimeType:      fileType.MIME.Value,
	}, postIDInt64, user.ID)
	http.Redirect(w, r, "/view/"+fileName, 302)
}

// uploadPageHandler is the endpoint where the file upload page is served.
func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}
	err := renderTemplate(w, "upload.html", TemplateTemplate{
		LoggedIn:     loggedIn,
		LoggedInUser: user,
	})
	if err != nil {
		panic(err)
	}
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}
