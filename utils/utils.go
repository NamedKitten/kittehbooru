package utils

import (
	"sort"
	"strings"

	"regexp"
)

func FilterString(s string) string {
	re := regexp.MustCompile(`[^\p{L}\p{Z}:_-]+`)
	s = re.ReplaceAllLiteralString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return s
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
