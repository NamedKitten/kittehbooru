package database

import (
	"context"
	"fmt"

	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/h2non/bimg"
	"github.com/rs/zerolog/log"
)

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
	if strings.HasPrefix(contentFilename, "frontend/") || isTmpFile {
		contentFile, err = os.Open(contentFilename)
	} else {
		if !db.ContentStorage.Exists(ctx, contentFilename) {
			log.Error().Msg("Content File Does Not Exist")
			return ""
		}
		contentFile, err = db.ContentStorage.ReadFile(ctx, contentFilename)
	}
	defer contentFile.Close()
	if err != nil {
		log.Error().Err(err).Msg("Lost File?")
		return ""
	}

	buffer, err := ioutil.ReadAll(contentFile)
	if err != nil {
		log.Error().Err(err).Msg("Image Read")
		return "frontend/img/preview-not-available.jpg"
	}
	o := bimg.Options{
		Height:      0,
		Width:       300,
		Quality:     70,
		Compression: 100,
		Embed:       true,
		Type:        bimg.WEBP,
	}

	newImage, err := bimg.NewImage(buffer).Process(o)
	if err != nil {
		log.Error().Err(err).Msg("Resize Image")
	}

	newCacheFile, err := db.ThumbnailsStorage.WriteFile(ctx, thumbnailFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache Create")
		return ""
	}
	newCacheFile.Write(newImage)
	newCacheFile.Close()
	return thumbnailFile

}
