package handlers

import (
	"fmt"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/golang/freetype"
	"github.com/gorilla/mux"
	"github.com/hqbobo/text2pic"
	"github.com/rs/zerolog/log"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func createThumbnail(post types.Post) string {
	originalFilename := fmt.Sprintf("content/%s.%s", post.Filename, post.FileExtension)
	// The file where the generated thumbnail is stored.
	var contentFilename string
	thumbnailFile := fmt.Sprintf("cache/%d", post.PostID)

	if post.FileExtension == "swf" {
		return "img/flash.jpg"
	} else if strings.HasPrefix(post.MimeType, "video/") {
		if DB.Settings.VideoThumbnails {
			tmpFile, _ := ioutil.TempFile("", "video_thumbnail_")
			contentFilename = tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			err := exec.Command("ffmpegthumbnailer", "-c", "png", "-i", originalFilename, "-o", tmpFile.Name()).Run()
			if err != nil {
				return "img/video.jpg"
			}
		} else {
			return "img/video.jpg"
		}
	} else if post.FileExtension == "pdf" {
		if DB.Settings.PDFThumbnails {
			tmpFile, _ := ioutil.TempFile("", "pdf_thumbnail_")
			contentFilename = tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			err := exec.Command("convert", "-format", "png", "-thumbnail", "x300", "-background", "white", "-alpha", "remove", originalFilename+"[0]", contentFilename).Run()
			if err != nil {
				return "img/pdf.jpg"
			}
		} else {
			return "img/pdf.jpg"
		}

	} else if post.FileExtension == "unknown" {
		data, _ := ioutil.ReadFile(originalFilename)
		dataSplit := strings.Split(string(data), "\n")
		fontBytes, err := ioutil.ReadFile("fonts/Lato.ttf")
		if err != nil {
			log.Error().Err(err).Msg("Read Font Fail")
			return ""
		}
		f, err := freetype.ParseFont(fontBytes)
		if err != nil {
			log.Error().Err(err).Msg("Parse Font")
			return ""
		}
		pic := text2pic.NewTextPicture(text2pic.Configure{Width: 4000, BgColor: text2pic.ColorWhite})
		pic.AddTextLine("Text File", 40, f, text2pic.ColorBlack, text2pic.Padding{})

		for ii, l := range dataSplit {
			if ii == 8 {
				break
			} else {
				re := regexp.MustCompile("[[:^ascii:]]")
				nl := re.ReplaceAllLiteralString(l, "")
				pic.AddTextLine(nl, 30, f, text2pic.ColorBlack, text2pic.Padding{})
			}
		}
		tmpFile, _ := ioutil.TempFile("", "text_thumbnail_")
		contentFilename = tmpFile.Name()
		pic.Draw(tmpFile, text2pic.TypePng)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())
	} else {
		// Otherise just use the image file.
		contentFilename = originalFilename
	}

	contentFile, err := os.Open(contentFilename)
	if err != nil {
		log.Error().Err(err).Msg("Lost File?")
		return ""
	}

	image, _, err := image.Decode(contentFile)
	if err != nil {
		log.Error().Err(err).Msg("Image Decode")
		return ""
	}
	newCacheFile, err := os.Create(thumbnailFile)
	if err != nil {
		log.Error().Err(err).Msg("Cache Create")
		return ""
	}
	if DB.Settings.ThumbnailFormat == "png" {
		err = png.Encode(newCacheFile, image)
	} else {
		err = jpeg.Encode(newCacheFile, image, &jpeg.Options{70})
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
	post, ok := DB.Post(int64(postID))
	var cacheFile *os.File

	var cacheFilename string
	if !ok {
		cacheFilename = "img/file-not-found.jpg"
	} else {
		cacheFilename = "cache/" + post.Filename
	}

	if !fileExists(cacheFilename) {
		cacheFilename = createThumbnail(post)
	}

	cacheFile, err = os.Open(cacheFilename)
	if err != nil {
		log.Error().Err(err).Msg("Open Cache File")
		return
	}

	w.Header().Set("Cache-Control", "max-age=2592000")
	io.Copy(w, cacheFile)
}
