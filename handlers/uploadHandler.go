package handlers

import (
	"io/ioutil"

	"github.com/bwmarrin/snowflake"
	"github.com/rs/zerolog/log"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"

	"github.com/h2non/filetype"
)

var whitelistedTypes = [...]string{
	"image/jpeg",
	"image/png",
	"image/webp",
	"image/gif",
	"video/mp4",
	"video/x-matroska",
	"video/webm",
	"audio/mpeg",
	"audio/ogg",
	"audio/x-flac",
	"application/pdf",
	"application/x-shockwave-flash",
}

// maxUploadSize is the maximum filesize for a post.
// TODO: move to DB.Settings
// Default: 64Mb
const maxUploadSize = 64 * 1024 * 1024

// uploadHandler is the API endpoint for creating posts.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		log.Error().Msg("Not Logged In")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Error().Err(err).Msg("File Too Big")
		renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		log.Error().Err(err).Msg("File can't be found in form.")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("Can't read file")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	fileType, err := filetype.Match(fileBytes)
	if err != nil {
		log.Error().Err(err).Msg("Can't match fileType")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	mimeType := fileType.MIME.Value
	extension := fileType.Extension

	validType := false
	for _, t := range whitelistedTypes {
		if t == mimeType {
			validType = true
		}
	}

	if !validType {
		renderError(w, "INVALID_FORMAT", http.StatusBadRequest)
		return
	}

	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	postID := node.Generate()
	postIDInt64 := postID.Int64()
	fileName := strconv.Itoa(int(postIDInt64))
	tags := utils.SplitTagsString(r.PostFormValue("tags"))
	description := r.PostFormValue("description")



	newPath := filepath.Join("content/", fileName+"."+extension)
	newFile, err := os.Create(newPath)
	if err != nil {
		log.Error().Err(err).Msg("File Create")
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()
	if _, err = newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
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

	p := types.Post{
		PostID:        postIDInt64,
		Filename:      fileName,
		FileExtension: strings.TrimPrefix(extension, "."),
		Tags:          tags,
		Description:   description,
		Poster:        user.Username,
		CreatedAt:     postID.Time(),
		Sha256:        sha256sum,
		MimeType:      mimeType,
	}
	//go createThumbnails(p)

	err = DB.AddPost(p)
	if err != nil {
		log.Error().Err(err).Msg("Post Creation")
		renderError(w, "POST_CREATE_ERR", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/view/"+fileName, http.StatusFound)
}

// uploadPageHandler is the endpoint where the file upload page is served.
func UploadPageHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	err := templates.RenderTemplate(w, "upload.html", templates.T{
		LoggedIn:     loggedIn,
		LoggedInUser: user,
		Translator:   i18n.GetTranslator(r),
	})
	if err != nil {
		panic(err)
	}
}
