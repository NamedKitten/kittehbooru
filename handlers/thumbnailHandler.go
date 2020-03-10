package handlers

import (
	"fmt"

	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// thumbnailHandler handles serving and downscaling post images as thumbnails.
func ThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		return
	}
	post, ok := DB.Post(ctx, int64(postID))
	var cacheFile io.ReadCloser

	var cacheFilename string
	if !ok {
		cacheFilename = "frontend/img/file-not-found.jpg"
	} else {
		cacheFilename = fmt.Sprintf("%d.webp", post.PostID)
	}

	if !DB.ThumbnailsStorage.Exists(ctx, cacheFilename) {
		cacheFilename = DB.CreateThumbnail(ctx, post)
	}

	// Return early if no cache file could be created.
	if cacheFilename == "" {
		return
	}

	if !DB.ThumbnailsStorage.Exists(ctx, cacheFilename) {
		log.Error().Msg("Cache File Not Exist")
		return
	}
	cacheFile, err = DB.ThumbnailsStorage.ReadFile(ctx, cacheFilename)
	if err != nil {
		log.Error().Err(err).Msg("Open Cache File")
		return
	}

	defer cacheFile.Close()

	w.Header().Set("Cache-Control", "public, immutable, only-if-cached, max-age=2592000")
	_, err = io.Copy(w, cacheFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache File Upload")
	}
}
