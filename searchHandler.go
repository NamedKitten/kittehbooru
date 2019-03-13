package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
)

// SearchResultsTemplate contains data to be used in the template.
type SearchResultsTemplate struct {
	// The posts that match the search for a page.
	Results []Post
	// RealPage is the real page number for the current page.
	RealPage int
	// Page is RealPage + 1 and is used to show a 1-based page number index.
	Page int
	// TotalPages is the total number of pages for a given search query
	TotalPages int
	// NextLink is the link that will take you to the next page.
	NextLink string
	// PrevLink is the link that will take you to the previous page.
	PrevLink string
	// NumPosts is the total number of posts for a search query
	NumPosts int
	// Tags is the tags from the search query args, used to refill
	// the search bar.
	Tags string
	TemplateTemplate
}

// searchHandler is the search endpoint used for displaying results
// of a search query.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	vars := mux.Vars(r)
	tagsStr := vars["tags"]
	if len(tagsStr) == 0 {
		tagsStr = "*"
	}
	tags := splitTagsString(tagsStr)
	log.Error(tags)
	pageStr := vars["page"]
	if len(tagsStr) == 0 {
		tagsStr = "0"
	}
	page, _ := strconv.Atoi(pageStr)
	log.Error(page)
	matchingPosts := DB.getSearchPage(tags, page)
	var prevPage int
	if page <= 0 {
		prevPage = 0
	} else {
		prevPage = page - 1
	}
	query, _ := url.ParseQuery(r.URL.RawQuery)
	query.Set("page", strconv.Itoa(page+1))
	nextLink := "/search?" + query.Encode()
	query.Set("page", strconv.Itoa(prevPage))
	prevLink := "/search?" + query.Encode()

	searchResults := SearchResultsTemplate{
		Results:    matchingPosts,
		RealPage:   page,
		Page:       page + 1,
		NumPosts:   DB.numOfPostsForTags(tags),
		TotalPages: DB.numOfPagesForTags(tags),
		NextLink:   nextLink,
		PrevLink:   prevLink,
		Tags:       tagsStr,
		TemplateTemplate: TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}

	err := renderTemplate(w, "search.html", searchResults)
	if err != nil {
		panic(err)
	}
}
