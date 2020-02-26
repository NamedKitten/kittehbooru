package handlers

import (
	"github.com/bwmarrin/snowflake"
	"github.com/rs/zerolog/log"
	"io/ioutil"

	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
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
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	fileType, _ := filetype.Match(fileBytes)

	node, _ := snowflake.NewNode(1)
	postID := node.Generate()
	postIDInt64 := postID.Int64()
	fileName := strconv.Itoa(int(postIDInt64))
	tags := utils.SplitTagsString(r.PostFormValue("tags"))
	description := r.PostFormValue("description")

	newPath := filepath.Join("content/", fileName+"."+fileType.Extension)
	newFile, err := os.Create(newPath)
	if err != nil {
		log.Error().Err(err).Msg("File Create")
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		log.Error().Err(err).Msg("File Write")
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	newTags := make([]string, 0)

	for _, tag := range tags {
		if !strings.HasPrefix(tag, "user:") {
			newTags = append(newTags, tag)
		}
	}
	tags = newTags

	tags = append(tags, "user:"+user.Username)

	sha256sum := utils.Sha256Bytes(fileBytes)

	DB.AddPost(types.Post{
		PostID:        postIDInt64,
		Filename:      fileName,
		FileExtension: strings.TrimPrefix(fileType.Extension, "."),
		Tags:          tags,
		Description:   description,
		Poster:        user.Username,
		CreatedAt:     postID.Time(),
		Sha256:        sha256sum,
		MimeType:      fileType.MIME.Value,
	}, postIDInt64, user.Username)
	http.Redirect(w, r, "/view/"+fileName, 302)
}

// uploadPageHandler is the endpoint where the file upload page is served.
func UploadPageHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}
	err := templates.RenderTemplate(w, "upload.html", templates.TemplateTemplate{
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
