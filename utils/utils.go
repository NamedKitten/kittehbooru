package utils

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func FilterString(s string) string {
	re := regexp.MustCompile("[[:^ascii:]]")
	s = re.ReplaceAllLiteralString(s, "")
	s = strings.TrimSpace(s)
	return s
}

func EncryptPassword(password string) string {
	passwordBytes := []byte(password)
	bytes, err := bcrypt.GenerateFromPassword(passwordBytes, 12)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func CheckPassword(encryptedPassword, attemptPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(attemptPassword))
	return err == nil
}

func Sha256Bytes(b []byte) string {
	h := sha256.New()
	_, __ := h.Write(b)
	_ = __
	return fmt.Sprintf("%x", h.Sum(nil))
}

func SplitTagsString(tags string) []string {
	tags = strings.Replace(tags, "+", " ", -1)
	tags = strings.Replace(tags, ", ", " ", -1)
	return strings.Split(tags, " ")
}

func TagsListToString(tags []string) string {
	for i, s := range tags {
		tags[i] = FilterString(s)
	}
	sort.Strings(tags)
	return strings.Join(tags, "+")
}

func Paginate(x []int64, page int, pageSize int) []int64 {
	skip := pageSize * page
	if skip <= 0 {
		skip = 0
	}
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
