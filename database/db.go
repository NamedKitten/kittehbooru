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
	"strings"
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
	// Sessions handles logged in user sessions
	Sessions Sessions
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

// init creates all the database fields and starts cache and session management.
func (db *DBType) init() {
	snowflake.Epoch = 1551864242
	var err error
	db.sqldb, err = sql.Open("sqlite3", "file:db.sql")
	if err != nil {
		log.Warn().Err(err).Msg("SQL Open")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "users" (  "avatarID"  INTEGER,  "owner"  BOOL,  "admin"  BOOL,  "username"  TEXT,  "description"  TEXT,  PRIMARY KEY("username"));`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Users Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "passwords" (  "username"  TEXT, "password"  TEXT,  PRIMARY KEY("username"));`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Passwords Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "tags" (  "tag"  TEXT, "posts"  string,  PRIMARY KEY("tag"));`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Tags Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "posts" (  "postid" INTEGER, "filename"  TEXT, "ext" string, "description" text, "tags"  string, "poster" string, "timestamp" integer, "sha256" string, "mimetype" string, PRIMARY KEY("postid"));`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Posts Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "sessions" (  "token" string, "username" string, "expiry" integer, PRIMARY KEY("token"));`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Posts Table")
	}
	if !db.SetupCompleted {
		log.Warn().Msg("You need to go to /setup in web browser to setup this imageboard.")
	}
	if db.Settings.ReCaptcha {
		captcha, _ = recaptcha.NewReCAPTCHA(db.Settings.ReCaptchaPrivkey, recaptcha.V3, 10*time.Second)
	}

	go db.SearchCache.Start()
	go db.Sessions.Start(db.sqldb)
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

func (db *DBType) SetPassword(username string, password string) (err error) {
	stmt, err := db.sqldb.Prepare(`INSERT OR REPLACE INTO "passwords"("username", "password") VALUES (?, ?);`)
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't prepare statement")
		return err
	}
	_, err = stmt.Exec(username, utils.EncryptPassword(password))
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't execute statement")
		return err
	}
	return nil
}

func (db *DBType) CheckPassword(username string, password string) bool {
	var encPasswd string
	row := db.sqldb.QueryRow(`select password from passwords where username=?`, username)
	switch err := row.Scan(&encPasswd); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return utils.CheckPassword(encPasswd, password)
	default:
		return false
	}
	return false
}

func (db *DBType) AddUser(u types.User) {
	stmt, err := db.sqldb.Prepare(`INSERT INTO "users"("avatarID","owner","admin","username","description") VALUES (?,?,?,?,?);`)
	if err != nil {
		log.Warn().Err(err).Msg("AddUser can't prepare statement")
	}

	_, err = stmt.Exec(u.AvatarID, u.Owner, u.Admin, u.Username, "")
	if err != nil {
		log.Warn().Err(err).Msg("AddUser can't execute statement")
	}
}

func (db *DBType) User(username string) (types.User, bool) {
	u := types.User{}

	rows, err := db.sqldb.Query(`select "avatarID","owner","admin","username","description" from users where username = ?`, username)
	if err != nil {
		log.Error().Err(err).Msg("User can't prepare statement")
		return u, false
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&u.AvatarID, &u.Owner, &u.Admin, &u.Username, &u.Description)
		if err != nil {
			log.Error().Err(err).Msg("User can't execute statementr")
			return u, false
		} else {
			return u, username == u.Username
		}
	}
	return u, false
}

func (db *DBType) EditUser(u types.User) (err error) {
	stmt, err := db.sqldb.Prepare(`update users set avatarID=?, owner=?, admin=?, description=? where username = ?`)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser can't prepare statement")
		return err
	}
	log.Warn().Str("desc", u.Description).Msg("EditUser")

	_, err = stmt.Exec(u.AvatarID, u.Owner, u.Admin, u.Description, u.Username)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser can't execute statement")
		return err
	}
	return nil
}

func (db *DBType) DeleteUser(username string) error {
	stmt, err := db.sqldb.Prepare(`delete from users where username = ?`)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't prepare delete user statement")
		return err
	}

	_, err = stmt.Exec(username)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't execute delete user statement")
		return err
	}

	stmt, err = db.sqldb.Prepare(`delete from passwords where username = ?`)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't prepare delete password statement")
		return err
	}

	_, err = stmt.Exec(username)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't execute delete password statement")
		return err
	}

	rows, err := db.sqldb.Query(`select "posts" from tags where tag = ?`, "user:" + username)
	if err != nil {
		log.Error().Err(err).Msg("DeleteUser can't prepare select tags statement")
	}
	defer rows.Close()

	var posts []int64
	var postsString string

	for rows.Next() {
		rows.Scan(&postsString)
	}

	err = json.Unmarshal([]byte(postsString), &posts)
	for _, post := range posts {
		db.DeletePost(post)
	}

	db.Sessions.InvalidateSession(username)

	return nil
}

func (db *DBType) Post(postID int64) (types.Post, bool) {
	p := types.Post{}
	var tags string

	rows, err := db.sqldb.Query(`select "filename", "ext", "description", "tags", "poster", "timestamp", "sha256", "mimetype" from posts where postID = ?`, postID)
	if err != nil {
		log.Error().Err(err).Msg("Post can't query")
		return p, false
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&p.Filename, &p.FileExtension, &p.Description, &tags, &p.Poster, &p.CreatedAt, &p.Sha256, &p.MimeType)
		if err != nil {
			log.Error().Err(err).Msg("Post can't scan")
			return p, false
		} else {
			p.PostID = postID
			p.Tags = utils.SplitTagsString(tags)
			return p, true
		}
	}
	return p, false
}

// AddPost adds a post to the DB and adds it to the author's post list.
func (db *DBType) AddPost(post types.Post, postID int64, username string) int64 {
	post.Poster = username
	post.PostID = postID

	stmt, err := db.sqldb.Prepare(`INSERT INTO "posts"("postid", "filename", "ext", "description", "tags", "poster", "timestamp", "sha256", "mimetype") VALUES (?,?,?,?,?,?,?,?,?);`)
	if err != nil {
		log.Warn().Err(err).Msg("AddPost can't prepare insert post statement")
	}

	_, err = stmt.Exec(postID, post.Filename, post.FileExtension, post.Description, utils.TagsListToString(post.Tags), post.Poster, post.CreatedAt, post.Sha256, post.MimeType)
	if err != nil {
		log.Warn().Err(err).Msg("AddPost can't execute insert post statement")
	}

	for _, tag := range post.Tags {

		rows, err := db.sqldb.Query(`select "posts" from tags where tag = ?`, tag)
		if err != nil {
			log.Error().Err(err).Msg("AddPost can't prepare select tags statement")
		}
		defer rows.Close()

		var posts []int64
		var postsString string

		for rows.Next() {
			rows.Scan(&postsString)
		}

		err = json.Unmarshal([]byte(postsString), &posts)
		if err != nil {
			posts = make([]int64, 0)
		}
		posts = append(posts, postID)
		x, _ := json.Marshal(posts)

		stmt, err := db.sqldb.Prepare(`INSERT OR REPLACE INTO "tags"("tag", "posts") VALUES (?, ?);`)
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't prepare insert tags statement")
		}
		_, err = stmt.Exec(tag, string(x))
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't execute insert tags statement")
		}

	}

	//db.Posts[postID] = post
	return postID
}

func (db *DBType) EditPost(postID int64, post types.Post) {

}

func (db *DBType) DeletePost(postID int64) {

	p, _ := db.Post(postID)
	for _, tag := range p.Tags {
		rows, err := db.sqldb.Query(`select "posts" from tags where tag = ?`, tag)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost can't prepare select tags statement")
		}
		defer rows.Close()

		var posts []int64
		newPosts := make([]int64, 0)
		var postsString string

		for rows.Next() {
			err = rows.Scan(&postsString)
			if err != nil {
				log.Error().Err(err).Msg("DeletePost can't scan rows")
				return
			}
		}
		err = json.Unmarshal([]byte(postsString), &posts)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost Json Unmarshal Error")
			return
		}

		for _, post := range posts {
			if post != postID {
				newPosts = append(newPosts, post)
			}
		}

		x, _ := json.Marshal(newPosts)
		s, err := db.sqldb.Prepare(`INSERT OR REPLACE INTO "tags"("tag", "posts") VALUES (?, ?);`)
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't prepare insert tags statement")
		}
		_, err = s.Exec(tag, string(x))
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't execute insert tags statement")
		}

	}

	stmt, err := db.sqldb.Prepare(`delete from posts where postid = ?`)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't prepare delete post statement")
	}

	_, err = stmt.Exec(postID)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't execute delete post statement")
	}
}

// getPostsForTags gets posts matching tags from DB
// it uses a tags table which maps a tag to all the posts containing a tag
func (db *DBType) getPostsForTags(tags []string) []int64 {
	// we need to make sure to keep track of how many times the post
	// is seen and only get which posts appear for all of the positive posts
	// basically a simple way of getting the intersection of all positive tags
	// so that we only get the posts that match ALL of the positive tags
	posCount := 0
	posCounts := make(map[int64]int)
	negMatch := make(map[int64]bool)

	// lets first remove any duplicate tags
	tempTags := make(map[string]bool)
	// this will remove duplicate entrys
	for _, tag := range tags {
		tempTags[tag] = true
	}
	tags = make([]string, 0)
	for tag := range tempTags {
		shouldAdd := true
		// if there is a tag "foo" and also a tag "-foo", remove both of them to reduce database load
		if strings.HasPrefix(tag, "-") {
			if _, ok := tempTags[tag[1:]]; ok {
				shouldAdd = false
			}
		} else {
			if _, ok := tempTags["-"+tag]; ok {
				shouldAdd = false
			}
		}
		if shouldAdd {
			tags = append(tags, tag)
		}
	}

	for _, tag := range tags {
		// is it a positive tag or a negative tag?
		// true = positive, false = negative
		is := !strings.HasPrefix(tag, "-")

		if !is {
			// remove the - at start
			tag = tag[1:]
		} else {
			// increase the count of positive tags
			posCount += 1
		}

		// get what posts match tag
		var rows *sql.Rows
		var err error
		//posts will be all the posts that are tagged with `tag`
		posts := make([]int64, 0)

		if tag == "*" {
			rows, err = db.sqldb.Query(`select "postid" from posts where true`)
			if err != nil {
				log.Error().Err(err).Msg("GetPostsForTags can't query wildcard posts")
			}
			defer rows.Close()
			var pid int64
			for rows.Next() {
				rows.Scan(&pid)
				posts = append(posts, pid)
			}

		} else {
			rows, err = db.sqldb.Query(`select "posts" from tags where tag = ?`, tag)
			if err != nil {
				log.Error().Err(err).Msg("GetPostsForTags can't query tag posts")
			}
			var postsString string
			defer rows.Close()

			for rows.Next() {
				rows.Scan(&postsString)
			}
			// we store it as json just so its easy to store in the database
			json.Unmarshal([]byte(postsString), &posts)
		}


		if is {
			// add to counter if its a positive
			for _, post := range posts {
				if _, ok := posCounts[post]; ok {
					// increase the count
					posCounts[post] = posCounts[post] + 1
				} else {
					// add the count to map starting at 1 if not existing already
					posCounts[post] = 1
				}
			}
		} else {
			for _, post := range posts {
				// if its a negative match, aka post we DONT want, add it to this map instead
				negMatch[post] = true
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

// cacheSearch searches for posts matching tags and returns a
// array of post IDs matching those tags.
func (db *DBType) cacheSearch(searchTags []string) []int64 {
	combinedTags := utils.TagsListToString(searchTags)

	if val, ok := db.SearchCache.Get(combinedTags); ok {
		return val
	} else {
		matching := db.getPostsForTags(searchTags)
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
		p, _ := db.Post(post)
		matchingPosts = append(matchingPosts, p)
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
		if sess, ok := db.Sessions.CheckToken(c.Value); ok {
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
