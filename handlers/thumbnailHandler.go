package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/h2non/bimg.v1"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// fileExists is a helper function to tell if a file exists or not.
func fileExists(fn string) bool {
	if _, err := os.Stat(fn); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// thumbnailHandler handles serving and downscaling post images as thumbnails.
func ThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	post, ok := DB.Posts[int64(postID)]
	var cacheFile string
	if ok {
		cacheFile = "cache/" + post.Filename
	} else {
		cacheFile = "img/file-not-found.jpg"
	}

	if !fileExists(cacheFile) {

		var contentFilename string
		log.Error(post.MimeType)
		log.Error(post)
		if strings.HasPrefix(post.MimeType, "video/") {
			// TODO: Implement this with FFMpegThumbnailer
			log.Error("Not implemented video thumbnail")
			contentFilename = "img/video.jpg"
		} else if post.FileExtension == "swf" {
			// For flash files we will use the flash.jpg image
			// because we can't generate a thumbnail for a flash
			// animation.
			contentFilename = "img/flash.jpg"
		} else {
			// Otherise just use the image file.
			contentFilename = fmt.Sprintf("content/%s.%s", post.Filename, post.FileExtension)
		}

		contentFile, err := bimg.Read(contentFilename)
		if err != nil {
			log.Error("Content file open fail: ", err)
			return
		}

		// Resize to a width or 600px
		newImage, err := bimg.NewImage(contentFile).Resize(600, 0)
		if err != nil {
			log.Error("Error resizing file... ", err)
		}
		bimg.Write(cacheFile, newImage)
	}
	// Load the cached thumbnail file.
	cachedFile, err := os.Open(cacheFile)
	if err != nil {
		log.Error("Cache Open: ", err)
		return
	}
	w.Header().Set("Cache-Control", "max-age=2592000")

	io.Copy(w, cachedFile)
}
