package handlers

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/chai2010/webp"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/nfnt/resize"
)

func createThumbnail(post types.Post, ext string) string {
	originalFilename := fmt.Sprintf("content/%s.%s", post.Filename, post.FileExtension)
	// The file where the generated thumbnail is stored.
	var contentFilename string
	thumbnailFile := fmt.Sprintf("cache/%d.%s", post.PostID, ext)

	if post.FileExtension == "swf" {
		return "frontend/img/flash.jpg"
	} else if strings.HasPrefix(post.MimeType, "video/") {
		if DB.Settings.VideoThumbnails {
			tmpFile, _ := ioutil.TempFile("", "video_thumbnail_")
			contentFilename = tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			err := exec.Command("ffmpegthumbnailer", "-c", "png", "-i", originalFilename, "-o", tmpFile.Name()).Run()
			if err != nil {
				contentFilename = "frontend/img/video.png"
			}
		} else {
			contentFilename = "frontend/img/video.png"
		}
	} else if post.FileExtension == "pdf" {
		if DB.Settings.PDFThumbnails {
			tmpFile, _ := ioutil.TempFile("", "pdf_thumbnail_")
			contentFilename = tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			err := exec.Command("convert", "-format", "png", "-thumbnail", "x300", "-background", "white", "-alpha", "remove", originalFilename+"[0]", contentFilename).Run()
			if err != nil {
				contentFilename = "frontend/img/pdf.jpg"
			}
		} else {
			contentFilename = "frontend/img/pdf.jpg"
		}

	} else if strings.HasPrefix(post.MimeType, "image/") {
		// Otherise just use the image file.
		contentFilename = originalFilename
	} else {
		// we can't create anything for this format yet
		contentFilename = "frontend/img/preview-not-available.jpg"
	}
	contentFile, err := os.Open(contentFilename)
	if err != nil {
		log.Error().Err(err).Msg("Lost File?")
		return ""
	}

	image, _, err := image.Decode(contentFile)
	if err != nil {
		log.Error().Err(err).Msg("Image Decode")
		return "frontend/img/preview-not-available.jpg"
	}
	newCacheFile, err := os.Create(thumbnailFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache Create")
		return ""
	}

	resizedImage := resize.Resize(200, 0, image, resize.Lanczos3)

	if ext == "webp" {
		err = webp.Encode(newCacheFile, resizedImage, &webp.Options{Quality: 70})
	} else {
		err = jpeg.Encode(newCacheFile, resizedImage, &jpeg.Options{Quality: 50})
	}

	if err != nil {
		log.Error().Err(err).Msg("Encode Fail")
		return ""
	}
	return thumbnailFile

}

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
	if err != nil {
		return
	}
	ext := vars["ext"]
	post, ok := DB.Post(int64(postID))
	var cacheFile *os.File

	var cacheFilename string
	if !ok {
		cacheFilename = "frontend/img/file-not-found.jpg"
	} else {
		cacheFilename = "cache/" + post.Filename
	}

	if !fileExists(cacheFilename) {
		cacheFilename = createThumbnail(post, ext)
	}

	// Return early if no cache file could be created.
	if cacheFilename == "" {
		return
	}

	cacheFile, err = os.Open(cacheFilename)
	if err != nil {
		log.Error().Err(err).Msg("Open Cache File")
		return
	}

	w.Header().Set("Cache-Control", "public, immutable, only-if-cached, max-age=2592000")
	_, err = io.Copy(w, cacheFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache File Write")
	}
}
