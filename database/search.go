package database

import (
	"context"
	"math"
	"runtime/trace"
	"sort"
	"strings"

	"github.com/bwmarrin/snowflake"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
)

// paginate paginates a list of int64s
// TODO: add more comments
func paginate(x []int64, page int, pageSize int) []int64 {
	var limit int
	var start int
	numItems := len(x)
	// skips forward N items where N is how many items in a page multiplied by the page number
	skip := pageSize * page
	// prevents integer overflow if skip becomes negative
	if skip <= 0 {
		skip = 0
	}
	if skip > numItems {
		start = numItems
	} else {
		start = skip
	}
	if skip + pageSize > numItems {
		limit = numItems
	} else {
		limit = skip + pageSize
	}
	return x[start:limit]
}

// searchTag is a wrapper around AllPostIDs and TagPosts 
// for searching for all results for a tag or a wildcard match.
func (db *DB) searchTag(ctx context.Context, tag string) ([]int64) {
	defer trace.StartRegion(ctx, "DB/searchTag").End()

	var posts []int64
	var err error 
	if val, ok := db.SearchCache.Get(tag); ok {
		posts = val
	} else {
		if tag == "*" {
			posts, err = db.AllPostIDs(ctx)
		} else {
			posts, err = db.TagPosts(ctx, tag)
		}
	}
	if err != nil {
		return []int64{}
	}
	db.SearchCache.Add(tag, posts)
	return posts
}

// getPostsForTags gets posts matching tags from DB
// it uses a tags table which maps a tag to all the posts containing a tag
func (db *DB) getPostsForTags(ctx context.Context, tags []string) []int64 {
	defer trace.StartRegion(ctx, "DB/getPostsForTags").End()
	// we need to make sure to keep track of how many times the post
	// is seen and only get which posts appear for all of the positive posts
	// basically a simple way of getting the intersection of all positive tags
	// so that we only get the posts that match ALL of the positive tags
	posCount := 0
	posCounts := make(map[int64]int)
	negMatch := make(map[int64]bool)

	tags = db.filterTags(tags)

	for _, tag := range tags {
		// is it a positive tag or a negative tag
		// true = positive, false = negative
		is := !strings.HasPrefix(tag, "-")

		if !is {
			// remove the - at start
			tag = tag[1:]
		} else {
			// increase the count of positive tags
			posCount += 1
		}

		//posts will be all the posts that are tagged with `tag`
		posts := db.searchTag(ctx, tag)

		for _, post := range posts {
			if !is {
				// if its a negative match, aka post we DONT want, add it to this map instead
				negMatch[post] = true
			} else if i, ok := posCounts[post]; ok {
				// add to counter of positive counts
				posCounts[post] = i + 1
			} else {
				// add the count to map starting at 1 if not existing already
				posCounts[post] = 1
			}
		}
	}

	finalPostIDs := make([]int64, 0)

	for posPost, posCountTimes := range posCounts {
		// so we only get the posts that match ALL positive tags
		if posCountTimes == posCount {
			found := false
			for negPost := range negMatch {
				// if there is a post that is a negative match, do not add this to the finalPostIDs array
				if posPost == negPost {
					found = true
				}
			}
			if !found {
				finalPostIDs = append(finalPostIDs, posPost)
			}
		}
	}
	return finalPostIDs
}

// TopNCommonTags returns the top N common tags for a search result
func (db *DB) TopNCommonTags(ctx context.Context, n int, tags []string) []types.TagCounts {
	defer trace.StartRegion(ctx, "DB/Top15CommonTags").End()

	posts := db.cacheSearch(ctx, tags)
	tagCounts := make(map[string]int)
	for _, p := range posts {
		post, err := db.Post(ctx, p)
		if err != nil {
			continue
		}
		for _, tag := range post.Tags {
			if i, ok := tagCounts[tag]; ok {
				tagCounts[tag] = i + 1
			} else {
				tagCounts[tag] = 1
			}
		}
	}

	tagCountsSlice := make([]types.TagCounts, 0, len(tagCounts))
	for k, v := range tagCounts {
		tagCountsSlice = append(tagCountsSlice, types.TagCounts{k, v})
	}
	// Sort by tag name first so equal value results are in alphabetical order
	sort.Slice(tagCountsSlice, func(i, j int) bool {
		return tagCountsSlice[i].Tag > tagCountsSlice[j].Tag
	})

	// Sort by how many posts a tag has
	sort.Slice(tagCountsSlice, func(i, j int) bool {
		return tagCountsSlice[i].Count > tagCountsSlice[j].Count
	})
	// Calculate the min between how many tags there are and N
	// Prevents panic when N > tag count
	x := math.Min(float64(n), float64(len(tagCountsSlice)))
	return tagCountsSlice[:int(x)]
}

// cacheSearch searches for posts matching tags and returns a
// array of post IDs matching those tags.
func (db *DB) cacheSearch(ctx context.Context, searchTags []string) []int64 {
	defer trace.StartRegion(ctx, "DB/cacheSearch").End()

	var result []int64
	searchTags = db.filterTags(searchTags)
	combinedTags := utils.TagsListToString(searchTags)
	// If it is in the cache then great! use the cached result
	// otherise search for them and add to the cache.
	if val, ok := db.SearchCache.Get(combinedTags); ok {
		result = val
	} else {
		matching := db.getPostsForTags(ctx, searchTags)
		go db.SearchCache.Add(combinedTags, matching)
		result = matching
	}
	// Sort by posted time
	// TODO add a parameter to have different sorting modes such as sorting from oldest to newest
	// or from filesize.
	sort.Slice(result, func(i, j int) bool {
		return snowflake.ID(result[i]).Time() > snowflake.ID(result[j]).Time()
	})
	return result
}

// GetSearchIDs returns a paginated list of Post IDs from a list of tags.
func (db *DB) GetSearchIDs(ctx context.Context, searchTags []string, page int) []int64 {
	defer trace.StartRegion(ctx, "DB/GetSearchIDs").End()

	matching := db.cacheSearch(ctx, searchTags)
	return paginate(matching, page, 20)
}

// getSearchPage returns a paginated list of posts from a list of tags.
func (db *DB) GetSearchPage(ctx context.Context, searchTags []string, page int) []types.Post {
	defer trace.StartRegion(ctx, "DB/GetSearchPage").End()

	matching := db.cacheSearch(ctx, searchTags)
	pageContent := paginate(matching, page, 20)
	matchingPosts := make([]types.Post, len(pageContent))
	for i, post := range pageContent {
		p, err := db.Post(ctx, post)
		if err != nil {
			continue
		}
		matchingPosts[i] = p
	}
	return matchingPosts
}
