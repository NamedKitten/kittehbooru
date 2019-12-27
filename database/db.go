package database

import (
	"database/sql"
	json "encoding/json"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/bwmarrin/snowflake"
	"github.com/ezzarghili/recaptcha-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"math"
	"net/http"
	"os"
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
	sqldb *sql.DB
	// SetupCompleted is used to know when to run setup page.
	SetupCompleted bool `json:"init"`
	// LockedTags is a map of tags which are locked and the user ID
	// of the user that can use them.
	LockedTags map[string]string `json:"lockedTags"`
	// Sessions is a map of token IDs and session info.
	Sessions     map[string]types.Session `json:"sessions"`
	sessionsLock sync.Mutex
	// Passwords is a map of user IDs to their bcrypt2 encrypted
	// hashes.
	Passwords map[string]string `json:"passwords"`
	// types.Users is a map of user ID to their user data.
	Users map[string]types.User `json:"users"`
	// Posts is a map of post ID to the post.
	Posts map[int64]types.Post `json:"posts"`
	// SearchCache is a cache of search strings and the post IDs
	// that match the result.
	SearchCache SearchCache
	// Settings contains instance-specific settings for this instance.
	Settings Settings `json:"settings"`
}

// Save saves the database.
func (db *DBType) Save() {
	log.Info().Msg("Saving DB.")
	data, err := json.Marshal(db)
	if err != nil {
		log.Error().Err(err).Msg("Can't encode DB to json")
	}
	err = ioutil.WriteFile("db.json", data, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Can't save DB")
	}
}

// numOfPostsForTags returns the total number of posts for a list of tags.
func (db *DBType) NumOfPostsForTags(searchTags []string) int {
	return len(db.cacheSearch(searchTags))
}

// numOfPagesForTags returns the total number of pages for a list of tags.
func (db *DBType) NumOfPagesForTags(searchTags []string) int {
	return int(math.Ceil(float64(db.NumOfPostsForTags(searchTags)) / float64(20)))
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

// init creates all the database fields and starts cache and session management.
func (db *DBType) init() {
	snowflake.Epoch = 1551864242
	var err error
	db.sqldb, err = sql.Open("sqlite3", "file:db.sql")
	if err != nil {
		log.Warn().Err(err).Msg("SQL Error")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "users" (  "id"  INTEGER,  "avatarID"  INTEGER,  "owner"  BOOL,  "admin"  BOOL,  "username"  TEXT,  "description"  TEXT,  PRIMARY KEY("id"));`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Error")
	}
	if db.Users == nil {
		db.Users = make(map[string]types.User, 0)
	}
	if db.Passwords == nil {
		db.Passwords = make(map[string]string, 0)
	}
	if db.Posts == nil {
		db.Posts = make(map[int64]types.Post, 0)
	}
	if db.Sessions == nil {
		db.Sessions = make(map[string]types.Session, 0)
	}
	if !db.SetupCompleted {
		log.Warn().Msg("You need to go to /setup in web browser to setup this imageboard.")
	}
	if db.Settings.ReCaptcha {
		captcha, _ = recaptcha.NewReCAPTCHA(db.Settings.ReCaptchaPrivkey, recaptcha.V3, 10*time.Second)
	}

	go db.SearchCache.Start()
	go db.sessionCleaner()
}

// LoadDB loads the database from the db.json file and initializes it.
func LoadDB() *DBType {
	var db *DBType
	_, err := os.Stat("db.json")
	if err != nil {
		os.Create("db.json")
	}
	data, _ := ioutil.ReadFile("db.json")
	err = json.Unmarshal(data, &db)
	if err != nil {
		log.Error().Err(err).Msg("Cannot unmarshal DB")
		db = &DBType{}
	}
	db.init()
	db.Save()
	return db
}

func (db *DBType) AddUser(u types.User) {
	stmt, err := db.sqldb.Prepare(`INSERT INTO "users"("avatarID","owner","admin","username","description") VALUES (?,?,?,?,?);`)
	if err != nil {
		log.Warn().Err(err).Msg("AddUser SQL Error")
	}

	_, err = stmt.Exec(u.AvatarID, u.Owner, u.Admin, u.Username, u.Description)
	if err != nil {
		log.Warn().Err(err).Msg("AddUser SqlExec Error")
	}
}

func (db *DBType) User(username string) (u types.User, err error) {
	u = types.User{}

	rows, err := db.sqldb.Query(`select "avatarID","owner","admin","username","description" from users where username = ?`, username)
	if err != nil {
		log.Error().Err(err).Msg("User SQL Error")
		return u, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&u.AvatarID, &u.Owner, &u.Admin, &u.Username, &u.Description)
		if err != nil {
			log.Error().Err(err).Msg("User SQLScan Error")
			return u, err
		}
	}
	err = rows.Err()
	if err != nil {
		log.Error().Err(err).Msg("User RowsSQL Error")
		return u, err
	}
	return u, nil

}

func (db *DBType) EditUser(u types.User) (err error) {
	stmt, err := db.sqldb.Prepare(`update users set avatarID = ?, owner = ?, admin = ?, username = ?, description = ? where username = ?`)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser SQL Error")
		return err
	}

	_, err = stmt.Exec(u.AvatarID, u.Owner, u.Admin, u.Username, u.Description, u.Username)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser SQLExec Error")
		return err
	}
	return nil
}


func (db *DBType) DeleteUser(username string) {
	// redo
	user := types.User{}
	for _, postID := range user.Posts {
		db.DeletePost(postID)
	}
	delete(db.Passwords, user.Username)
}

func (db *DBType) CreateSession(username string) string {
	db.sessionsLock.Lock()
	sessionToken := utils.GenSessionToken()
	db.Sessions[sessionToken] = types.Session{Username: username, ExpirationTime: time.Now().Add(time.Hour * 3).Unix()}
	db.sessionsLock.Unlock()
	return sessionToken
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DBType) AddPost(post types.Post, postID int64, username string) int64 {
	user := db.Users[username]
	user.Posts = append(user.Posts, postID)
	db.Users[username] = user
	db.Posts[postID] = post
	return postID
}

func (db *DBType) DeletePost(postID int64) {
	//authorID := db.Posts[postID].PosterID
	delete(db.Posts, postID)
	/*author := db.Users[authorID]
	author.Posts = utils.RemoveFromSlice(db.Users[authorID].Posts, postID)
	db.Users[authorID] = author*/
}

// cacheSearch searches for posts matching tags and returns a
// array of post IDs matching those tags.
func (db *DBType) cacheSearch(searchTags []string) []int64 {
	combinedTags := utils.TagsListToString(searchTags)

	if val, ok := db.SearchCache.Get(combinedTags); ok {
		return val
	} else {

		var matching []int64
		for i, item := range db.Posts {
			if utils.DoesMatchTags(searchTags, item) {
				matching = append(matching, i)
			}
		}
		sort.Slice(matching, func(i, j int) bool { return snowflake.ID(i).Time() < snowflake.ID(j).Time() })
		db.SearchCache.Add(combinedTags, matching)
		return matching
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
			u, _ := db.User(sess.Username)
			return u, true
		} else {
			return types.User{}, false
		}
	}
	return types.User{}, false
}

func (db *DBType) VerifyRecaptcha(resp string) bool {
	if db.Settings.ReCaptcha {
		err := captcha.Verify(resp)
		if err != nil {
			log.Error().Err(err).Msg("Cannot verify recaptcha")
			return false
		} else {
			return true
		}
	} else {
		return true
	}

}
