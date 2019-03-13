package main

import (
	json "encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/ezzarghili/recaptcha-go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"time"
	"sync"
)

var captcha recaptcha.ReCAPTCHA

// Settings are instance-specific settings.
type Settings struct {
	// Title is the name of this instance.
	SiteName string `json:"siteName"`
	// ReCaptcha tells if Google's reCaptcha should be used for registration.
	ReCaptcha bool `json:"reCaptcha"`
	// ReCaptchaPubkey is the public key for the reCaptcha API.
	ReCaptchaPubkey string `json:"reCaptchaPubKey"`
	// ReCaptchaPrivkey is the private key for the reCaptcha API.
	ReCaptchaPrivkey string `json:"reCaptchaPrivKey"`
	// Rules are the instance specific rules for this instance.
	Rules string `json:"rules"`
}

// DBType is the type at which all things are stored in the database.
type DBType struct {
	// SetupCompleted is used to know when to run setup page.
	SetupCompleted bool `json:"init"`
	// LockedTags is a map of tags which are locked and the user ID
	// of the user that can use them.
	LockedTags map[string]int64 `json:"lockedTags"`
	// Sessions is a map of session token to the user ID the session
	// belongs to.
	Sessions map[string]int64 `json:"sessions"`
	// Passwords is a map of user IDs to their bcrypt2 encrypted
	// hashes.
	Passwords map[int64]string `json:"passwords"`
	// Users is a map of user ID to their user data.
	Users map[int64]User `json:"users"`
	// Posts is a map of post ID to the post.
	Posts map[int64]Post `json:"posts"`
	// SearchCache is a cache of search strings and the post IDs
	// that match the result.
	SearchCache     map[string][]int64 `json:"searchCache"`
	searchCacheLock sync.Mutex
	// UsernameToID is used to easily fetch the user ID from a username
	// to make it so you dont need to itterate over every user to see if
	// a username exists.
	UsernameToID map[string]int64 `json:"usernameToID"`
	// searchCacheTimes is a map of the search strings to the time at which
	// they where last searched.
	searchCacheTimes map[string]int64
	// Settings contains instance-specific settings for this instance.
	Settings Settings `json:"settings"`
}

// Save saves the database.
func (db *DBType) Save() {
	log.Info("Saving DB.")
	data, _ := json.Marshal(db)
	ioutil.WriteFile("db.json", data, 0644)
}

// numOfPostsForTags returns the total number of posts for a list of tags.
func (db *DBType) numOfPostsForTags(searchTags []string) int {
	return len(db.cacheSearch(searchTags))
}

// numOfPagesForTags returns the total number of pages for a list of tags.
func (db *DBType) numOfPagesForTags(searchTags []string) int {
	return int(math.Ceil(float64(db.numOfPostsForTags(searchTags)) / float64(DefaultPostsPerPage)))
}

// cacheCleaner runs in the background to remove expired searches from the cache.
func (db *DBType) cacheCleaner() {
	for true {
		db.searchCacheLock.Lock()
		log.Error(db.SearchCache)
		for tags, _ := range db.SearchCache {
			val, ok := db.searchCacheTimes[tags]			
			if !ok || (time.Unix(val, 0).Add(CacheExpirationTime).After(time.Now())) {
				log.Info(tags + " has expired, removing from cache.")
				delete(db.SearchCache, tags)
				delete(db.searchCacheTimes, tags)
			}
		}
		db.searchCacheLock.Unlock()

		time.Sleep(CacheTimer)
	}
}

// init creates all the database fields and adds a user for testing.
func (db *DBType) init() {
	snowflake.Epoch = 1551864242
	if db.Users == nil {
		db.Users = make(map[int64]User, 0)
	}
	if db.Passwords == nil {
		db.Passwords = make(map[int64]string, 0)
	}
	if db.Posts == nil {
		db.Posts = make(map[int64]Post, 0)
	}
	if db.SearchCache == nil {
		db.SearchCache = make(map[string][]int64, 0)
	}
	if db.searchCacheTimes == nil {
		db.searchCacheTimes = make(map[string]int64, 0)
	}
	if db.UsernameToID == nil {
		db.UsernameToID = make(map[string]int64, 0)
	}
	if db.Sessions == nil {
		db.Sessions = make(map[string]int64, 0)
	}
	if !db.SetupCompleted {
		log.Warning("You need to go to /setup in web browser to setup this imageboard.")
	}
	if db.Settings.ReCaptcha {
		captcha, _ = recaptcha.NewReCAPTCHA(db.Settings.ReCaptchaPrivkey, recaptcha.V3, 10*time.Second)
	}

	go db.cacheCleaner()

}

// LoadDB loads the database from the db.json file and initializes it.
func LoadDB() DBType {
	var db DBType
	data, _ := ioutil.ReadFile("db.json")
	json.Unmarshal(data, &db)
	db.init()
	db.Save()
	return db
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DBType) AddPost(post Post, postID, userID int64) int64 {
	user := DB.Users[userID]
	user.Posts = append(user.Posts, postID)
	DB.Users[userID] = user
	DB.Posts[postID] = post
	return postID
}

func (db *DBType) DeletePost(postID int64) {
	authorID := DB.Posts[postID].PosterID
	delete(DB.Posts, postID)
	author := DB.Users[authorID]
	author.Posts = removeFromSlice(DB.Users[authorID].Posts, postID)
	DB.Users[authorID] = author
}

// cacheSearch searches for posts matching tags and returns a
// array of post IDs matching those tags.
func (db *DBType) cacheSearch(searchTags []string) []int64 {
	combinedTags := tagsListToString(searchTags)
	db.searchCacheLock.Lock()
	defer db.searchCacheLock.Unlock()

	if val, ok := db.SearchCache[combinedTags]; ok {
		return val
	} else {

		var matching []int64
		for i, item := range db.Posts {
			if doesMatchTags(searchTags, item) {
				matching = append(matching, i)
			}
		}
		sort.Slice(matching, func(i, j int) bool { return snowflake.ID(i).Time() < snowflake.ID(j).Time() })
		db.SearchCache[combinedTags] = matching
		db.searchCacheTimes[combinedTags] = time.Now().Unix()
		return db.SearchCache[combinedTags]
	}
}

// getSearchPage returns a paginated list of posts from a list of tags.
func (db *DBType) getSearchPage(searchTags []string, page int) []Post {
	matching := db.cacheSearch(searchTags)
	var matchingPosts []Post
	for _, post := range paginate(matching, page, DefaultPostsPerPage) {
		matchingPosts = append(matchingPosts, db.Posts[post])
	}

	return matchingPosts
}

// CheckForLoggedInUser is a helper function that is used to see if a
// HTTP request is from a logged in user.
// It returns a User struct and a bool to tell if there was a logged in
// user or not.
func (db *DBType) CheckForLoggedInUser(r *http.Request) (User, bool) {
	c, err := r.Cookie("sessionToken")
	if err == nil {
		if id, ok := DB.Sessions[c.Value]; ok {
			return DB.Users[id], true
		} else {
			return User{}, false
		}
	} else {
		log.Error(err)
	}
	return User{}, false
}

func (db *DBType) verifyRecaptcha(resp string) bool {
	if db.Settings.ReCaptcha {
		err := captcha.Verify(resp)
		if err != nil {
			log.Error(err)
			return false
		} else {
			return true
		}
	} else {
		return true
	}

}
