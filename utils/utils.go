package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/NamedKitten/kittehimageboard/types"
	"golang.org/x/crypto/bcrypt"
	"sort"
	"strings"
)

func AnythingToBytes(i interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(i)
	return buf.Bytes()
}

func AnythingFromBytes(x []byte, i interface{}) {
	var buf bytes.Buffer
	buf.Write(x)
	dec := gob.NewDecoder(&buf)
	dec.Decode(&i)
}

func EncryptPassword(password string) string {
	passwordBytes := []byte(password)
	bytes, _ := bcrypt.GenerateFromPassword(passwordBytes, 12)
	return string(bytes)
}

func CheckPassword(encryptedPassword, attemptPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(attemptPassword))
	return err == nil
}

func GenSessionToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func Sha256Bytes(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func SplitTagsString(tags string) []string {
	tags = strings.Replace(tags, "+", " ", -1)
	tags = strings.Replace(tags, ", ", " ", -1)
	return strings.Split(tags, " ")
}

func TagsListToString(tags []string) string {
	sort.Strings(tags)
	return strings.Join(tags, "+")
}

func DoesMatchTags(searchTags []string, post types.Post) bool {
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

func Paginate(x []int64, page int, pageSize int) []int64 {
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

func RemoveFromSlice(slice []int64, toDelete int64) []int64 {
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
