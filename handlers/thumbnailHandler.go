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
	var cacheFile io.ReadCloser

	cacheFilename := fmt.Sprintf("%d.webp", postID)
	cacheFile, err = DB.ThumbnailsStorage.ReadFile(ctx, cacheFilename)
	if err != nil {
		post, _ := DB.Post(ctx, int64(postID))
		cacheFilename = DB.CreateThumbnail(ctx, post)
		cacheFile, err = DB.ThumbnailsStorage.ReadFile(ctx, cacheFilename)
		if err != nil {
			log.Error().Err(err).Msg("Open Cache File")
			return
		}
	}

	// Return early if no cache file could be created.
	if cacheFilename == "" {
		return
	}

	defer cacheFile.Close()

	w.Header().Set("Cache-Control", "public, immutable, only-if-cached, max-age=2592000")
	_, err = io.Copy(w, cacheFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache File Upload")
	}
}
