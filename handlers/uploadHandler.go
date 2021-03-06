package handlers

import (
	"github.com/bwmarrin/snowflake"
	"github.com/rs/zerolog/log"

	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/NamedKitten/kittehbooru/i18n"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/NamedKitten/kittehbooru/types"
	"github.com/NamedKitten/kittehbooru/utils"

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
	ctx := r.Context()

	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		log.Error().Msg("Not Logged In")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Error().Err(err).Msg("File Too Big")
		renderError(w, "FILE_TOO_BIG", err, http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		log.Error().Err(err).Msg("File can't be found in form.")
		renderError(w, "INVALID_FILE", err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBuf := bytes.NewBuffer([]byte{})

	_, err = io.CopyN(fileBuf, file, 261)
	if err != nil {
		log.Error().Err(err).Msg("Can't read header")
		renderError(w, "INVALID_FILE", err, http.StatusBadRequest)
		return
	}

	fileType, err := filetype.Match(fileBuf.Bytes())
	if err != nil {
		log.Error().Err(err).Msg("Can't match fileType")
		renderError(w, "INVALID_FILE", err, http.StatusBadRequest)
		return
	}

	_, err = io.Copy(fileBuf, file)
	if err != nil {
		log.Error().Err(err).Msg("Can't read rest of file")
		renderError(w, "INVALID_FILE", err, http.StatusBadRequest)
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
		renderError(w, "INVALID_FORMAT", errors.New("Invalid Format"), http.StatusBadRequest)
		return
	}

	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	postID := node.Generate()
	postIDInt64 := postID.Int64()
	fileName := strconv.Itoa(int(postIDInt64))

	newPath := fileName + "." + extension
	newFile, err := DB.ContentStorage.WriteFile(ctx, newPath)
	if err != nil {
		log.Error().Err(err).Msg("File Create")
		renderError(w, "CANT_WRITE_FILE", err, http.StatusInternalServerError)
		return
	}
	defer newFile.Close()
	if _, err = io.Copy(newFile, fileBuf); err != nil {
		log.Error().Err(err).Msg("File Write")
		renderError(w, "CANT_WRITE_FILE", err, http.StatusInternalServerError)
		return
	}

	err = newFile.Close()
	if err != nil {
		log.Error().Err(err).Msg("File Close")
		renderError(w, "CANT_CLOSE_FILE", err, http.StatusInternalServerError)
		return
	}

	tags := utils.SplitTagsString(r.PostFormValue("tags"))
	description := r.PostFormValue("description")

	newTags := make([]string, 0)

	for _, tag := range tags {
		if !strings.HasPrefix(tag, "user:") {
			newTags = append(newTags, tag)
		}
	}
	tags = newTags

	tags = append(tags, "user:"+user.Username)

	p := types.Post{
		PostID:        postIDInt64,
		Filename:      fileName,
		FileExtension: strings.TrimPrefix(extension, "."),
		Tags:          tags,
		Description:   description,
		Poster:        user.Username,
		CreatedAt:     postID.Time(),
		MimeType:      mimeType,
	}
	go DB.CreateThumbnail(ctx, p)

	err = DB.AddPost(ctx, p)
	if err != nil {
		log.Error().Err(err).Msg("Post Creation")
		renderError(w, "POST_CREATE_ERR", err, http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/view/"+fileName, http.StatusFound)
}

// uploadPageHandler is the endpoint where the file upload page is served.
func UploadPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
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
		renderError(w, "TEMPLATE_RENDER_ERROR", err, http.StatusBadRequest)
	}
}
