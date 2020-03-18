package database

import (
	"context"
	"fmt"

	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/NamedKitten/kittehbooru/types"
	"github.com/rs/zerolog/log"
)

// createVideoThumbnail creates a thumbnail from a video using ffmpegthumbnailer
func (db *DB) createVideoThumbnail(ctx context.Context, post types.Post) (string, bool) {
	if !db.Settings.VideoThumbnails {
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

	contentFile, _ := db.ContentStorage.ReadFile(ctx, originalFilename)
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

// createPDFThumbnail creates a thumbnail from a pdf using imagemagick
func (db *DB) createPDFThumbnail(ctx context.Context, post types.Post) (string, bool) {
	if !db.Settings.PDFThumbnails {
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

	contentFile, _ := db.ContentStorage.ReadFile(ctx, originalFilename)
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

// CreateThumbnail creates a thumbnail for a post.
func (db *DB) CreateThumbnail(ctx context.Context, post types.Post) string {
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
		contentFilename, success = db.createVideoThumbnail(ctx, post)
		if success {
			defer os.Remove(contentFilename)
			isTmpFile = true
		}

	} else if post.FileExtension == "pdf" {
		var success bool
		contentFilename, success = db.createPDFThumbnail(ctx, post)
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
	// If it is a frontend file, it is on the OS, not remote or accessable from the storage
	if strings.HasPrefix(contentFilename, "frontend/") || isTmpFile {
		contentFile, err = os.Open(contentFilename)
	} else {
		contentFile, err = db.ContentStorage.ReadFile(ctx, contentFilename)
		if err != nil {
			log.Error().Msg("Content File Does Not Exist")
			return ""
		}
	}
	defer contentFile.Close()
	if err != nil {
		log.Error().Err(err).Msg("Lost File?")
		return ""
	}

	tmpContentFile, err := ioutil.TempFile("", "content_")
	if err != nil {
		log.Error().Err(err).Msg("Can't create temp file")
		return ""
	}
	defer os.Remove(tmpContentFile.Name())

	io.Copy(tmpContentFile, contentFile)
	tmpContentFile.Close()

	tmpOutputFile, err := ioutil.TempFile("", "output_")
	if err != nil {
		log.Error().Err(err).Msg("Can't create temp file")
		return ""
	}
	defer tmpOutputFile.Close()
	defer os.Remove(tmpOutputFile.Name())

	cmd := exec.Command("convert", "-format", "webp", "-thumbnail", "x300", tmpContentFile.Name(), "webp:"+tmpOutputFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("Can't convert thumbnail")
		return ""
	}

	newCacheFile, err := db.ThumbnailsStorage.WriteFile(ctx, thumbnailFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache Create")
		return ""
	}
	io.Copy(newCacheFile, tmpOutputFile)
	newCacheFile.Close()
	return thumbnailFile
}
