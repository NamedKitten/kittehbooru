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
	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
)

func sizeToWidth(s string) int {
	switch s {
	case "small":
		return 200
	case "medium":
		return 400
	case "large":
		return 800
	default:
		return 400
	}
}

func sanitisedSize(s string) string {
	switch s {
	case "small":
		return "small"
	case "medium":
		return "medium"
	case "large":
		return "large"
	default:
		return "large"
	}
}

func createThumbnails(post types.Post) {
	createThumbnail(post, "jpeg", "small")
	createThumbnail(post, "jpeg", "medium")
	createThumbnail(post, "jpeg", "large")
	createThumbnail(post, "webp", "small")
	createThumbnail(post, "webp", "medium")
	createThumbnail(post, "webp", "large")
}

func createThumbnail(post types.Post, ext string, size string) string {
	log.Error().Msg("Creating Thumbnail")

	originalFilename := fmt.Sprintf("content/%s.%s", post.Filename, post.FileExtension)
	// The file where the generated thumbnail is stored.
	var contentFilename string
	thumbnailFile := fmt.Sprintf("cache/%d-%s.%s", post.PostID, size, ext)

	if post.FileExtension == "swf" {
		return "frontend/img/flash.jpg"
	} else if strings.HasPrefix(post.MimeType, "video/") {
		if DB.Settings.VideoThumbnails {
			tmpFile, err := ioutil.TempFile("", "video_thumbnail_")
			if err != nil {
				log.Error().Err(err).Msg("Can't create temp file")
				return "frontend/img/video.png"
			}
			contentFilename = tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			err = exec.Command("ffmpegthumbnailer", "-c", "png", "-i", originalFilename, "-o", tmpFile.Name()).Run()
			if err != nil {
				contentFilename = "frontend/img/video.png"
			}
		} else {
			contentFilename = "frontend/img/video.png"
		}
	} else if post.FileExtension == "pdf" {
		if DB.Settings.PDFThumbnails {
			tmpFile, err := ioutil.TempFile("", "pdf_thumbnail_")
			if err != nil {
				log.Error().Err(err).Msg("Can't create temp file")
				return "frontend/img/pdf.jpg"
			}
			contentFilename = tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			err = exec.Command("convert", "-format", "png", "-thumbnail", "x300", "-background", "white", "-alpha", "remove", originalFilename+"[0]", contentFilename).Run()
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
		DB.DeletePost(post.PostID)
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

	resizedImage := resize.Resize(uint(sizeToWidth(size)), 0, image, resize.Lanczos3)

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

	size := sanitisedSize(vars["size"])

	var cacheFilename string
	if !ok {
		cacheFilename = "frontend/img/file-not-found.jpg"
	} else {
		cacheFilename = fmt.Sprintf("cache/%d-%s.%s", post.PostID, size, ext)
	}

	if !fileExists(cacheFilename) {
		cacheFilename = createThumbnail(post, ext, size)
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
