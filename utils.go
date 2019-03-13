package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"sort"
	"strings"
)

func encryptPassword(password string) string {
	passwordBytes := []byte(password)
	bytes, _ := bcrypt.GenerateFromPassword(passwordBytes, 12)
	return string(bytes)
}

func checkPassword(encryptedPassword, attemptPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(attemptPassword))
	return err == nil
}

func genSessionToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func sha256Bytes(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func splitTagsString(tags string) []string {
	tags = strings.Replace(tags, "+", " ", -1)
	tags = strings.Replace(tags, ", ", " ", -1)
	return strings.Split(tags, " ")
}

func tagsListToString(tags []string) string {
	sort.Strings(tags)
	return strings.Join(tags, "+")
}

func doesMatchTags(searchTags []string, post Post) bool {
	if len(searchTags) == 0 {
		return false
	}
	tagMatched := false

	for _, searchTag := range searchTags {
		reverse := strings.HasPrefix(searchTag, "-")
		if reverse {
			searchTag = searchTag[1:]
		}

		if searchTag == "*" {
			tagMatched = true
		}
		if !tagMatched {
			for _, itemTag := range post.Tags {
				if searchTag == itemTag {
					tagMatched = true
				}
			}
		}

		if tagMatched && reverse {
			return false
		} else if !tagMatched {
			return false
		}
	}
	return tagMatched
}

func paginate(x []int64, page int, pageSize int) []int64 {
	skip := pageSize * page
	numItems := len(x)
	limit := func() int {

		if skip+pageSize > numItems {
			return numItems
		} else {
			return skip + pageSize
		}

	}

	start := func() int {
		if skip > numItems {
			return numItems
		} else {
			return skip
		}

	}
	return x[start():limit()]
}

func removeFromSlice(slice []int64, toDelete int64) []int64 {
	var newSlice []int64
	for _, v := range slice {
		if v == toDelete {
			continue
		} else {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}
