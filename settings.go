package main

import "time"

// CacheTimer is how often the background thread looks for searches to remove from cache.
const CacheTimer = time.Second * 5

// CacheExpirationTime is how long before a search query can stay in the cache
// without being searched again for it to be removed from the cache to save memory
const CacheExpirationTime = time.Second * 30

// DefaultPostsPerPage is the default number of posts that are
// displayed from a search request.
const DefaultPostsPerPage = 40
