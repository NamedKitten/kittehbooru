package handlers

import (
	"fmt"
	"github.com/hqbobo/text2pic"
	"image"
	"image/jpeg"
	_ "image/png"
	_ "image/gif"

	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"regexp"

	"github.com/golang/freetype"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
	if !ok {
		cacheFile = "img/file-not-found.jpg"
	} else if post.FileExtension == "swf" {
		cacheFile = "img/flash.jpg"
	} else {
		cacheFile = "cache/" + post.Filename
	}

	if !fileExists(cacheFile) {
		var contentFilename string
		log.Error(post.MimeType)
		log.Error(post)
		originalFilename := fmt.Sprintf("content/%s.%s", post.Filename, post.FileExtension)

		if strings.HasPrefix(post.MimeType, "video/") {
			contentFilename = "/tmp/" + post.Filename + ".jpeg"
			cmd := exec.Command("ffmpegthumbnailer", "-i", originalFilename, "-o", contentFilename)
			cmd.Run()
		}  else if post.FileExtension == "pdf" {
			contentFilename = "/tmp/" + post.Filename + ".png"
			cmd := exec.Command("convert", "-thumbnail", "x300", "-background", "white", "-alpha", "remove", originalFilename + "[0]", contentFilename)
			cmd.Run()
		} else if post.FileExtension == "unknown" {
			data, _ := ioutil.ReadFile(originalFilename)
			dataSplit := strings.Split(string(data), "\n")
			fontBytes, err := ioutil.ReadFile("fonts/Lato.ttf")
			if err != nil {
				log.Error("Read Font Fail: ", err)
				return
			}
			f, err := freetype.ParseFont(fontBytes)
			if err != nil {
				log.Error("Parse FontErr: ", err)
				return
			}
			pic := text2pic.NewTextPicture(text2pic.Configure{Width: 4000, BgColor: text2pic.ColorWhite})
			pic.AddTextLine("Text File", 40, f, text2pic.ColorBlack, text2pic.Padding{})

			for ii, l := range dataSplit {
				if ii == 8 {
					break
				} else {
					log.Error(l)
					re := regexp.MustCompile("[[:^ascii:]]")
					nl := re.ReplaceAllLiteralString(l, "")
					pic.AddTextLine(nl, 30, f, text2pic.ColorBlack, text2pic.Padding{})
				}
			}
			contentFilename = "/tmp/" + post.Filename + ".jpeg"
			fil, err := os.Create(contentFilename)
			if err != nil {
				log.Error("Content file open fail: ", err)
				return
			}
			pic.Draw(fil, text2pic.TypeJpeg)
			fil.Close()

		} else {
			// Otherise just use the image file.
			log.Error(post.FileExtension)
			contentFilename = originalFilename
		}

		contentFile, err := os.Open(contentFilename)
		if err != nil {
			log.Error("Content file open fail: ", err)
			return
		}

		image, _, err := image.Decode(contentFile)
		if err != nil {
			log.Error("Error decoding file... ", err)
		}
		newCacheFile, err := os.Create(cacheFile)
		if err != nil {
			log.Error("Error creating cache file... ", err)
		}

		err = jpeg.Encode(newCacheFile, image, &jpeg.Options{70})
		if err != nil {
			log.Error("Error encoding file... ", err)
		}
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
