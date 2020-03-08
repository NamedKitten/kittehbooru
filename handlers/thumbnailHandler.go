package handlers

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
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
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func createVideoThumbnail(ctx context.Context, post types.Post) (string, bool) {
	if !DB.Settings.VideoThumbnails {
		return "frontend/img/video.png", false
	}

	originalFilename := fmt.Sprintf("%s.%s", post.Filename, post.FileExtension)

	tmpFile, err := ioutil.TempFile("", "video_thumbnail_")
	if err != nil {
		log.Error().Err(err).Msg("Can't create temp file")
		return "frontend/img/video.png", false
	}
	vidTmpFile, err := ioutil.TempFile("", "video_")
	if err != nil {
		log.Error().Err(err).Msg("Can't create temp file")
		return "frontend/img/video.png", false
	}
	tmpFile.Close()
	defer os.Remove(vidTmpFile.Name())

	contentFile, _ := DB.ContentStorage.ReadFile(ctx, originalFilename)
	io.Copy(vidTmpFile, contentFile)
	vidTmpFile.Close()

	cmd := exec.CommandContext(ctx, "ffmpegthumbnailer", "-c", "png", "-i", vidTmpFile.Name(), "-o", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("Can't run ffmpegthumbnailer")
		return "frontend/img/video.png", false
	}
	return tmpFile.Name(), true
}

func createPDFThumbnail(ctx context.Context, post types.Post) (string, bool) {
	if !DB.Settings.PDFThumbnails {
		return "frontend/img/pdf.jpg", false
	}

	originalFilename := fmt.Sprintf("%s.%s", post.Filename, post.FileExtension)
	tmpFile, err := ioutil.TempFile("", "pdf_thumbnail_")
	if err != nil {
		log.Error().Err(err).Msg("Can't create temp file")
		return "frontend/img/pdf.jpg", false
	}
	contentFilename := tmpFile.Name()

	pdfTmpFile, err := ioutil.TempFile("", "pdf_")
	if err != nil {
		log.Error().Err(err).Msg("Can't create temp file")
		return "frontend/img/pdf.jpg", false
	}
	defer os.Remove(pdfTmpFile.Name())

	contentFile, _ := DB.ContentStorage.ReadFile(ctx, originalFilename)
	io.Copy(pdfTmpFile, contentFile)
	defer pdfTmpFile.Close()
	defer tmpFile.Close()

	defer os.Remove(tmpFile.Name())
	err = exec.Command("convert", "-format", "png", "-thumbnail", "x300", "-background", "white", "-alpha", "remove", originalFilename+"[0]", contentFilename).Run()
	if err != nil {
		return "frontend/img/pdf.jpg", false
	}
	return pdfTmpFile.Name(), true

}

func createThumbnail(ctx context.Context, post types.Post) string {
	log.Debug().Int64("postid", post.PostID).Msg("Creating Thumbnail")

	originalFilename := fmt.Sprintf("%s.%s", post.Filename, post.FileExtension)
	// The file where the generated thumbnail is stored.
	var contentFilename string
	thumbnailFile := fmt.Sprintf("%d.webp", post.PostID)
	isTmpFile := false

	if post.FileExtension == "swf" {
		contentFilename = "frontend/img/flash.jpg"
	} else if strings.HasPrefix(post.MimeType, "video/") {
		var success bool
		contentFilename, success = createVideoThumbnail(ctx, post)
		if success {
			defer os.Remove(contentFilename)
			isTmpFile = true
		}

	} else if post.FileExtension == "pdf" {
		var success bool
		contentFilename, success = createPDFThumbnail(ctx, post)
		if success {
			defer os.Remove(contentFilename)
			isTmpFile = true
		}

	} else if strings.HasPrefix(post.MimeType, "image/") {
		// Otherise just use the image file.
		contentFilename = originalFilename
	} else {
		// we can't create anything for this format yet
		contentFilename = "frontend/img/preview-not-available.jpg"
	}

	var contentFile io.ReadCloser
	var err error
	if strings.HasPrefix(contentFilename, "frontend/") || isTmpFile {
		contentFile, err = os.Open(contentFilename)
	} else {
		if !DB.ContentStorage.Exists(ctx, contentFilename) {
			log.Error().Msg("Content File Does Not Exist")
			return ""
		}
		contentFile, err = DB.ContentStorage.ReadFile(ctx, contentFilename)
	}
	defer contentFile.Close()
	if err != nil {
		log.Error().Err(err).Msg("Lost File?")
		return ""
	}

	image, _, err := image.Decode(contentFile)
	if err != nil {
		log.Error().Err(err).Msg("Image Decode")
		return "frontend/img/preview-not-available.jpg"
	}
	newCacheFile, err := DB.ThumbnailsStorage.WriteFile(ctx, thumbnailFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache Create")
		return ""
	}
	defer newCacheFile.Close()

	resizedImage := imaging.Resize(image, 300, 0, imaging.NearestNeighbor)

	err = webp.Encode(newCacheFile, resizedImage, &webp.Options{Quality: 70})
	if err != nil {
		log.Error().Err(err).Msg("Encode Fail")
		return ""
	}
	return thumbnailFile

}

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
		cacheFilename = createThumbnail(ctx, post)
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
