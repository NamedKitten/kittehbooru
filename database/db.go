package database

import (
	json "encoding/json"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/bwmarrin/snowflake"
	"github.com/ezzarghili/recaptcha-go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"
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
	// ThumbnailFormat can be either "jpeg" or "png"
	ThumbnailFormat string `json:"thumbnailFormat"`
	// VideoThumbnails is to enable/disable creating video thumbnails.
	// This requires ffmpegthumbnailer to be installed.
	VideoThumbnails bool `json:"videoThumbnails"`
	// PDFThumbnails is to enable/disable creating PDF thumbnails.
	// This requires imagemagick's convert tool to be installed.
	PDFThumbnails bool `json:"pdfThumbnails"`
	// PDFView is to enable/disable the viewing of PDFs in the browser using pdf.js.
	PDFView bool `json:"pdfView"`
}

// DBType is the type at which all things are stored in the database.
type DBType struct {
	// SetupCompleted is used to know when to run setup page.
	SetupCompleted bool `json:"init"`
	// LockedTags is a map of tags which are locked and the user ID
	// of the user that can use them.
	LockedTags map[string]int64 `json:"lockedTags"`
	// Sessions is a map of token IDs and session info.
	Sessions     map[string]types.Session `json:"sessions"`
	sessionsLock sync.Mutex
	// Passwords is a map of user IDs to their bcrypt2 encrypted
	// hashes.
	Passwords map[int64]string `json:"passwords"`
	// types.Users is a map of user ID to their user data.
	Users map[int64]types.User `json:"users"`
	// Posts is a map of post ID to the post.
	Posts map[int64]types.Post `json:"posts"`
	// SearchCache is a cache of search strings and the post IDs
	// that match the result.
	SearchCache      map[string][]int64 `json:"searchCache"`
	searchCacheLock  sync.Mutex
	searchCacheTimes map[string]int64
	// types.UsernameToID is used to easily fetch the user ID from a username
	// to make it so you dont need to itterate over every user to see if
	// a username exists.
	UsernameToID map[string]int64 `json:"usernameToID"`
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
func (db *DBType) NumOfPostsForTags(searchTags []string) int {
	return len(db.cacheSearch(searchTags))
}

// numOfPagesForTags returns the total number of pages for a list of tags.
func (db *DBType) NumOfPagesForTags(searchTags []string) int {
	return int(math.Ceil(float64(db.NumOfPostsForTags(searchTags)) / float64(20)))
}

// cacheCleaner runs in the background to remove expired searches from the cache.
func (db *DBType) cacheCleaner() {
	for true {
		db.searchCacheLock.Lock()
		for tags := range db.SearchCache {
			val, ok := db.searchCacheTimes[tags]
			if !ok || (time.Unix(val, 0).Add(time.Second).After(time.Now())) {
				log.Info(tags + " has expired, removing from cache.")
				delete(db.SearchCache, tags)
				delete(db.searchCacheTimes, tags)
			}
		}
		db.searchCacheLock.Unlock()

		//time.Sleep(time.Second * 2)
	}
}

func (db *DBType) sessionCleaner() {
	for true {
		db.sessionsLock.Lock()
		for token, session := range db.Sessions {
			if time.Now().After(time.Unix(session.ExpirationTime, 0)) {
				delete(db.Sessions, token)
			}
		}
		db.sessionsLock.Unlock()

		time.Sleep(time.Second * 2)
	}
}

// init creates all the database fields and adds a user for testing.
func (db *DBType) init() {
	snowflake.Epoch = 1551864242
	if db.Users == nil {
		db.Users = make(map[int64]types.User, 0)
	}
	if db.Passwords == nil {
		db.Passwords = make(map[int64]string, 0)
	}
	if db.Posts == nil {
		db.Posts = make(map[int64]types.Post, 0)
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
		db.Sessions = make(map[string]types.Session, 0)
	}
	if !db.SetupCompleted {
		log.Warning("You need to go to /setup in web browser to setup this imageboard.")
	}
	if db.Settings.ReCaptcha {
		captcha, _ = recaptcha.NewReCAPTCHA(db.Settings.ReCaptchaPrivkey, recaptcha.V3, 10*time.Second)
	}

	go db.cacheCleaner()
	go db.sessionCleaner()
}

// LoadDB loads the database from the db.json file and initializes it.
func LoadDB() *DBType {
	var db *DBType
	data, _ := ioutil.ReadFile("db.json")
	json.Unmarshal(data, &db)
	db.init()
	db.Save()
	return db
}

func (db *DBType) DeleteUser(userID int64) {
	user := db.Users[userID]
	for _, postID := range user.Posts {
		db.DeletePost(postID)
	}
	delete(db.UsernameToID, user.Username)
	delete(db.Passwords, user.ID)
	delete(db.Users, user.ID)
}

func (db *DBType) CreateSession(userID int64) string {
	db.sessionsLock.Lock()
	sessionToken := utils.GenSessionToken()
	db.Sessions[sessionToken] = types.Session{UserID: userID, ExpirationTime: time.Now().Add(time.Hour * 3).Unix()}
	db.sessionsLock.Unlock()
	return sessionToken
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DBType) AddPost(post types.Post, postID, userID int64) int64 {
	user := db.Users[userID]
	user.Posts = append(user.Posts, postID)
	db.Users[userID] = user
	db.Posts[postID] = post
	return postID
}

func (db *DBType) DeletePost(postID int64) {
	authorID := db.Posts[postID].PosterID
	delete(db.Posts, postID)
	author := db.Users[authorID]
	author.Posts = utils.RemoveFromSlice(db.Users[authorID].Posts, postID)
	db.Users[authorID] = author
}

// cacheSearch searches for posts matching tags and returns a
// array of post IDs matching those tags.
func (db *DBType) cacheSearch(searchTags []string) []int64 {
	combinedTags := utils.TagsListToString(searchTags)
	db.searchCacheLock.Lock()
	defer db.searchCacheLock.Unlock()

	if val, ok := db.SearchCache[combinedTags]; ok {
		return val
	} else {

		var matching []int64
		for i, item := range db.Posts {
			if utils.DoesMatchTags(searchTags, item) {
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
func (db *DBType) GetSearchPage(searchTags []string, page int) []types.Post {
	matching := db.cacheSearch(searchTags)
	var matchingPosts []types.Post
	// TODO: add post per page for stuff
	for _, post := range utils.Paginate(matching, page, 20) {
		matchingPosts = append(matchingPosts, db.Posts[post])
	}

	sort.Slice(matchingPosts, func(i, j int) bool {
		return snowflake.ID(matchingPosts[i].PostID).Time() > snowflake.ID(matchingPosts[j].PostID).Time()
	})

	return matchingPosts
}

// CheckForLoggedIntypes.User is a helper function that is used to see if a
// HTTP request is from a logged in user.
// It returns a types.User struct and a bool to tell if there was a logged in
// user or not.
func (db *DBType) CheckForLoggedInUser(r *http.Request) (types.User, bool) {
	c, err := r.Cookie("sessionToken")
	if err == nil {
		if sess, ok := db.Sessions[c.Value]; ok {
			return db.Users[sess.UserID], true
		} else {
			return types.User{}, false
		}
	} else {
		log.Error(err)
	}
	return types.User{}, false
}

func (db *DBType) VerifyRecaptcha(resp string) bool {
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
