package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"runtime/trace"
	"sort"
	"strings"
	"time"

	"github.com/NamedKitten/kittehimageboard/storage"
	"gopkg.in/yaml.v2"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/bwmarrin/snowflake"
	"github.com/ezzarghili/recaptcha-go"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var captcha recaptcha.ReCAPTCHA

// Settings are instance-specific settings.
type Settings struct {
	// Title is the name of this instance.
	SiteName string `yaml:"siteName"`
	// ReCaptcha tells if Google's reCaptcha should be used for registration.
	ReCaptcha bool `yaml:"reCaptcha"`
	// ReCaptchaPubkey is the public key for the reCaptcha API.
	ReCaptchaPubkey string `yaml:"reCaptchaPubKey"`
	// ReCaptchaPrivkey is the private key for the reCaptcha API.
	ReCaptchaPrivkey string `yaml:"reCaptchaPrivKey"`
	// Rules are the instance specific rules for this instance.
	Rules string `yaml:"rules"`
	// VideoThumbnails is to enable/disable creating video thumbnails.
	// This requires ffmpegthumbnailer to be installed.
	VideoThumbnails bool `yaml:"videoThumbnails"`
	// PDFThumbnails is to enable/disable creating PDF thumbnails.
	// This requires imagemagick's convert tool to be installed.
	PDFThumbnails bool `yaml:"pdfThumbnails"`
	// PDFView is to enable/disable the viewing of PDFs in the browser using pdf.js.
	PDFView bool `yaml:"pdfView"`
	// Database URI
	DatabaseURI string `yaml:"databaseURI"`
	// Database Type
	DatabaseType string `yaml:"databaseType"`
	// Content Storage URI
	ContentStorage string `yaml:"contentStorage"`
	// Thumbnails Storage URI
	ThumbnailsStorage string `yaml:"thumbnailsStorage"`
	// Content URL
	ContentURL string `yaml:"contentURL"`
	// Thumbnail URL
	ThumbnailURL string `yaml:"thumbnailURL"`
	// Listen Address
	ListenAddress string `yaml:"listenAddress"`
}

// DB is the type at which all things are stored in the database.
type DB struct {
	sqldb *sql.DB
	// SetupCompleted is used to know when to run setup page.
	SetupCompleted bool `yaml:"init"`
	// Sessions handles logged in user sessions
	Sessions Sessions `yaml:"-"`
	// SearchCache is a cache of search strings and the post IDs
	// that match the result.
	SearchCache SearchCache `yaml:"-"`
	// Settings contains instance-specific settings for this instance.
	Settings Settings `yaml:"settings"`

	ContentStorage    types.Storage `yaml:"-"`
	ThumbnailsStorage types.Storage `yaml:"-"`
}

// Save saves the settings.
func (db *DB) Save() {
	log.Info().Msg("Saving settings.")
	data, err := yaml.Marshal(db)
	if err != nil {
		log.Error().Err(err).Msg("Can't encode settings to yaml")
	}
	err = ioutil.WriteFile("settings.yaml", data, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Can't save settings")
	}
}

// numOfPostsForTags returns the total number of posts for a list of tags.
func (db *DB) NumOfPostsForTags(ctx context.Context, searchTags []string) int {
	return len(db.cacheSearch(ctx, searchTags))
}

// numOfPagesForTags returns the total number of pages for a list of tags.
func (db *DB) NumOfPagesForTags(ctx context.Context, searchTags []string) int {
	return int(math.Ceil(float64(db.NumOfPostsForTags(ctx, searchTags)) / float64(20)))
}

// init creates all the database fields and starts cache and session management.
func (db *DB) init() {
	snowflake.Epoch = 1551864242
	var err error

	db.ContentStorage = storage.GetStorage(db.Settings.ContentStorage)
	db.ThumbnailsStorage = storage.GetStorage(db.Settings.ThumbnailsStorage)

	db.sqldb, err = sql.Open(db.Settings.DatabaseType, db.Settings.DatabaseURI)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Open")
	}

	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "users" (  "avatarID"  bigint,  "owner"  BOOL,  "admin"  BOOL,  "username"  TEXT,  "description"  TEXT,  PRIMARY KEY("username"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Users Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "passwords" (  "username"  TEXT, "password"  TEXT,  PRIMARY KEY("username"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Passwords Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "tags" (  "tag"  TEXT, "posts"  TEXT,  PRIMARY KEY("tag"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Tags Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "posts" (  "postid" bigint, "filename"  TEXT, "ext" TEXT, "description" TEXT, "tags"  TEXT, "poster" TEXT, "timestamp" bigint, "sha256" TEXT, "mimetype" TEXT, PRIMARY KEY("postid"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Posts Table")
	}
	_, err = db.sqldb.Exec(`CREATE TABLE IF NOT EXISTS "sessions" (  "token" TEXT, "username" TEXT, "expiry" bigint, PRIMARY KEY("token"))`)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Create Sessions Table")
	}
	if !db.SetupCompleted {
		log.Warn().Msg("You need to go to /setup in web browser to setup this imageboard.")
	}
	if db.Settings.ReCaptcha {
		captcha, err = recaptcha.NewReCAPTCHA(db.Settings.ReCaptchaPrivkey, recaptcha.V3, 10*time.Second)
		if err != nil {
			log.Error().Err(err).Msg("Can't init ReCAPTCHA")
			panic(err)
		}
	}

	go db.SearchCache.Start()
	go db.Sessions.Start(db.sqldb)
}

// LoadDB loads the database from the db.json file and initializes it.
func LoadDB() *DB {
	db := &DB{}
	_, err := os.Stat("settings.yaml")
	if err != nil {
		if err != nil {
			log.Fatal().Err(err).Msg("Please copy over settings_example.yaml to settings.yaml.")
			panic(err)
		}
	}
	data, err := ioutil.ReadFile("settings.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Can't read settings")
	}

	err = yaml.Unmarshal(data, &db)
	if err != nil {
		log.Error().Err(err).Msg("Can't unmarshal settings")
		db = &DB{}
	}
	db.init()
	db.Save()
	return db
}

func (db *DB) SetPassword(ctx context.Context, username string, password string) (err error) {
	defer trace.StartRegion(ctx, "DB/SetPassword").End()
	var query string
	if db.Settings.DatabaseType == "postgres" {
		query =  `INSERT INTO "passwords" ("username", "password") VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password`
	} else {
		query =  `INSERT OR REPLACE INTO "passwords"("username", "password") VALUES ($1, $2)`
	}
	_, err = db.sqldb.ExecContext(ctx, query, username, utils.EncryptPassword(password))
	if err != nil {
		log.Warn().Err(err).Msg("SetPassword can't execute statement")
		return err
	}
	return nil
}

func (db *DB) CheckPassword(ctx context.Context, username string, password string) bool {
	defer trace.StartRegion(ctx, "DB/CheckPassword").End()

	var encPasswd string
	row := db.sqldb.QueryRowContext(ctx, `select password from passwords where username=$1`, username)
	switch err := row.Scan(&encPasswd); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return utils.CheckPassword(encPasswd, password)
	default:
		return false
	}
}

func (db *DB) AddUser(ctx context.Context, u types.User) {
	defer trace.StartRegion(ctx, "DB/AddUser").End()

	_, err := db.sqldb.ExecContext(ctx, `INSERT INTO "users"("avatarID","owner","admin","username","description") VALUES ($1,$2,$3,$4,$5)`, u.AvatarID, u.Owner, u.Admin, u.Username, "")
	if err != nil {
		log.Warn().Err(err).Msg("AddUser can't execute statement")
	}
}

func (db *DB) User(ctx context.Context, username string) (types.User, bool) {
	defer trace.StartRegion(ctx, "DB/User").End()

	u := types.User{}

	rows, err := db.sqldb.QueryContext(ctx, `select "avatarID","owner","admin","username","description" from users where username = $1`, username)
	if err != nil {
		log.Error().Err(err).Msg("User can't query statement")
		return u, false
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&u.AvatarID, &u.Owner, &u.Admin, &u.Username, &u.Description)
		if err != nil {
			log.Error().Err(err).Msg("User can't scan")
		} else {
			return u, username == u.Username
		}
	}
	return u, false
}

func (db *DB) EditUser(ctx context.Context, u types.User) (err error) {
	defer trace.StartRegion(ctx, "DB/EditUser").End()

	_, err = db.sqldb.ExecContext(ctx, `update users set avatarID=$1, owner=$2, admin=$3, description=$4 where username = $5`, u.AvatarID, u.Owner, u.Admin, u.Description, u.Username)
	if err != nil {
		log.Warn().Err(err).Msg("EditUser can't execute statement")
		return err
	}
	return nil
}

func (db *DB) DeleteUser(ctx context.Context, username string) error {
	defer trace.StartRegion(ctx, "DB/DeleteUser").End()

	_, err := db.sqldb.ExecContext(ctx, `delete from users where username = $1`, username)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't execute delete user statement")
		return err
	}

	_, err = db.sqldb.ExecContext(ctx, `delete from passwords where username = $1`, username)
	if err != nil {
		log.Warn().Err(err).Msg("DeleteUser can't execute delete password statement")
		return err
	}

	rows, err := db.sqldb.QueryContext(ctx, `select "postid" from posts where poster = $1`, username)
	if err != nil {
		log.Error().Err(err).Msg("DeleteUser can't select posts")
	}
	defer rows.Close()

	var posts []int64
	var postsString string

	for rows.Next() {
		err = rows.Scan(&postsString)
		if err != nil {
			log.Error().Err(err).Msg("DeleteUser can't scan row")
			return err
		}
	}

	err = json.Unmarshal([]byte(postsString), &posts)
	if err != nil {
		log.Error().Err(err).Msg("Can't unmarshal posts list")
		return err
	}
	for _, post := range posts {
		err = db.DeletePost(ctx, post)
		if err != nil {
			log.Error().Err(err).Msg("Can't delete user's post")
			return err
		}
	}

	db.Sessions.InvalidateSession(ctx, username)

	return nil
}

func (db *DB) Post(ctx context.Context, postID int64) (types.Post, bool) {
	defer trace.StartRegion(ctx, "DB/Post").End()

	p := types.Post{}
	var tags string

	rows, err := db.sqldb.QueryContext(ctx, `select "filename", "ext", "description", "tags", "poster", "timestamp", "sha256", "mimetype" from posts where postID = $1`, postID)
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
func (db *DB) AddPost(ctx context.Context, post types.Post) error {
	defer trace.StartRegion(ctx, "DB/AddPost").End()

	_, err := db.sqldb.ExecContext(ctx, `INSERT INTO "posts"("postid", "filename", "ext", "description", "tags", "poster", "timestamp", "sha256", "mimetype") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, post.PostID, post.Filename, post.FileExtension, post.Description, utils.TagsListToString(post.Tags), post.Poster, post.CreatedAt, post.Sha256, post.MimeType)
	if err != nil {
		log.Warn().Err(err).Msg("AddPost can't execute insert post statement")
		return err
	}

	for _, tag := range post.Tags {
		rows, err := db.sqldb.QueryContext(ctx, `select "posts" from tags where tag = $1`, tag)
		if err != nil {
			log.Error().Err(err).Msg("AddPost can't select tags")
			return err
		}
		defer rows.Close()

		var postsString string
		for rows.Next() {
			err = rows.Scan(&postsString)
			if err != nil {
				log.Error().Err(err).Msg("AddPost can't scan row")
			}
		}

		var posts []int64
		err = json.Unmarshal([]byte(postsString), &posts)
		if err != nil {
			posts = []int64{post.PostID}
		} else {
			posts = append(posts, post.PostID)
		}
		x, err := json.Marshal(posts)
		if err != nil {
			log.Error().Err(err).Msg("AddPost can't marshal posts list")
			return err
		}
		var query string
		if db.Settings.DatabaseType == "postgres" {
			query =  `INSERT INTO "tags"("tag", "posts") VALUES ($1, $2) ON CONFLICT (tag) DO UPDATE SET posts = EXCLUDED.posts`
		} else {
			query =  `INSERT OR REPLACE INTO "tags"("tag", "posts") VALUES ($1, $2)`
		}

		_, err = db.sqldb.ExecContext(ctx, query, tag, string(x))
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't execute insert tags statement")
			return err
		}
	}

	return nil
}

func (db *DB) EditPost(ctx context.Context, postID int64, post types.Post) {
	defer trace.StartRegion(ctx, "DB/EditPost").End()

	err := db.DeletePost(ctx, postID)
	if err != nil {
		log.Error().Err(err).Msg("EditPost can't delete post")
		return
	}
	err = db.AddPost(ctx, post)
	if err != nil {
		log.Error().Err(err).Msg("EditPost can't add post")
		return
	}
}

func (db *DB) DeletePost(ctx context.Context, postID int64) error {
	defer trace.StartRegion(ctx, "DB/DeletePost").End()

	p, _ := db.Post(ctx, postID)
	for _, tag := range p.Tags {
		rows, err := db.sqldb.QueryContext(ctx, `select "posts" from tags where tag = $1`, tag)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost can't select tags")
			return err
		}
		defer rows.Close()

		var posts []int64
		newPosts := make([]int64, 0)
		var postsString string

		for rows.Next() {
			err = rows.Scan(&postsString)
			if err != nil {
				log.Error().Err(err).Msg("DeletePost can't scan rows")
				return err
			}
		}
		err = json.Unmarshal([]byte(postsString), &posts)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost Json Unmarshal Error")
			return err
		}

		posts = utils.RemoveFromSlice(posts, postID)
		x, err := json.Marshal(newPosts)
		if err != nil {
			log.Error().Err(err).Msg("DeletePost can't unmarshal posts list")
			return err
		}
		var query string
		if db.Settings.DatabaseType == "postgres" {
			query =  `INSERT INTO "tags"("tag", "posts") VALUES ($1, $2) ON CONFLICT (tag) DO UPDATE SET posts = EXCLUDED.posts`
		} else {
			query =  `INSERT OR REPLACE INTO "tags"("tag", "posts") VALUES ($1, $2)`
		}

		_, err = db.sqldb.ExecContext(ctx, query, tag, string(x))
		if err != nil {
			log.Warn().Err(err).Msg("AddPost Tags can't execute insert tags statement")
			return err
		}
	}

	_, err := db.sqldb.ExecContext(ctx, `delete from posts where postid = $1`, postID)
	if err != nil {
		log.Warn().Err(err).Msg("DeletePost can't execute delete post statement")
		return err
	}
	return nil
}

func (db *DB) getPostsForTag(ctx context.Context, tag string) []int64 {
	defer trace.StartRegion(ctx, "DB/getPostsForTag").End()

	var posts []int64
	if val, ok := db.SearchCache.Get(tag); ok {
		posts = val
	} else {
		if tag == "*" {
			rows, err := db.sqldb.QueryContext(ctx, `select "postid" from posts where true`)
			if err != nil {
				log.Error().Err(err).Msg("GetPostsForTags can't query wildcard posts")
				return []int64{}
			}
			defer rows.Close()
			var pid int64
			for rows.Next() {
				err = rows.Scan(&pid)
				if err != nil {
					log.Error().Err(err).Msg("GetPostsForTags can't scan row")
					return []int64{}
				}
				posts = append(posts, pid)
			}

		} else {
			rows, err := db.sqldb.QueryContext(ctx, `select "posts" from tags where tag = $1`, tag)
			if err != nil {
				log.Error().Err(err).Msg("GetPostsForTags can't query tag posts")
				return []int64{}
			}
			var postsString string
			defer rows.Close()

			for rows.Next() {
				err = rows.Scan(&postsString)
				if err != nil {
					log.Error().Err(err).Msg("GetPostsForTags can't scan row")
					continue
				}
			}
			// we store it as json just so its easy to store in the database
			err = json.Unmarshal([]byte(postsString), &posts)
			if err != nil {
				return []int64{}
			}
		}
	}
	db.SearchCache.Add(tag, posts)
	return posts
}

func (db *DB) filterTags(tags []string) []string {
	// lets first remove any duplicate tags
	tempTags := make(map[string]bool)
	// this will remove duplicate entrys
	for _, tag := range tags {
		tempTags[tag] = true
	}

	isOnlyNegatives := true

	tags = make([]string, 0)
	for tag := range tempTags {
		// if there is a tag "foo" and also a tag "-foo", remove both of them to reduce database load
		is := strings.HasPrefix(tag, "-")
		var ok bool
		if is {
			_, ok = tempTags[tag[1:]]
		} else {
			_, ok = tempTags["-"+tag]
		}
		if !ok && !(tag == " " || len(tag) == 0) {
			if !is {
				isOnlyNegatives = false
			}
			tags = append(tags, tag)
		}
	}

	// if there is only negative tags, add wildcard
	if isOnlyNegatives {
		tags = append(tags, "*")
	}
	return tags
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
		posts := db.getPostsForTag(ctx, tag)

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

func (db *DB) Top15CommonTags(ctx context.Context, tags []string) []types.TagCounts {
	defer trace.StartRegion(ctx, "DB/Top15CommonTags").End()

	posts := db.cacheSearch(ctx, tags)
	tagCounts := make(map[string]int, 0)
	for _, p := range posts {
		post, exists := db.Post(ctx, p)
		if !exists {
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

	sort.Slice(tagCountsSlice, func(i, j int) bool {
		return tagCountsSlice[i].Tag > tagCountsSlice[j].Tag
	})

	sort.Slice(tagCountsSlice, func(i, j int) bool {
		return tagCountsSlice[i].Count > tagCountsSlice[j].Count
	})

	x := math.Min(float64(15), float64(len(tagCountsSlice)))
	return tagCountsSlice[:int(x)]
}

// cacheSearch searches for posts matching tags and returns a
// array of post IDs matching those tags.
func (db *DB) cacheSearch(ctx context.Context, searchTags []string) []int64 {
	defer trace.StartRegion(ctx, "DB/cacheSearch").End()

	var result []int64
	searchTags = db.filterTags(searchTags)
	combinedTags := utils.TagsListToString(searchTags)
	if val, ok := db.SearchCache.Get(combinedTags); ok {
		result = val
	} else {
		matching := db.getPostsForTags(ctx, searchTags)
		db.SearchCache.Add(combinedTags, matching)
		result = matching
	}

	sort.Slice(result, func(i, j int) bool {
		return snowflake.ID(result[i]).Time() > snowflake.ID(result[j]).Time()
	})
	return result
}

// GetSearchIDs returns a paginated list of Post IDs from a list of tags.
func (db *DB) GetSearchIDs(ctx context.Context, searchTags []string, page int) []int64 {
	defer trace.StartRegion(ctx, "DB/GetSearchIDs").End()

	matching := db.cacheSearch(ctx, searchTags)
	return utils.Paginate(matching, page, 20)
}

// getSearchPage returns a paginated list of posts from a list of tags.
func (db *DB) GetSearchPage(ctx context.Context, searchTags []string, page int) []types.Post {
	defer trace.StartRegion(ctx, "DB/GetSearchPage").End()

	matching := db.cacheSearch(ctx, searchTags)
	pageContent := utils.Paginate(matching, page, 20)
	matchingPosts := make([]types.Post, len(pageContent))
	for i, post := range pageContent {
		p, _ := db.Post(ctx, post)
		matchingPosts[i] = p
	}
	return matchingPosts
}

// CheckForLoggedIntypes.User is a helper function that is used to see if a
// HTTP request is from a logged in user.
// It returns a types.User struct and a bool to tell if there was a logged in
// user or not.
func (db *DB) CheckForLoggedInUser(ctx context.Context, r *http.Request) (types.User, bool) {
	defer trace.StartRegion(ctx, "DB/CheckForLoggedInUser").End()

	c, err := r.Cookie("sessionToken")
	if err == nil {
		if sess, ok := db.Sessions.CheckToken(ctx, c.Value); ok {
			u, exists := db.User(ctx, sess.Username)
			if exists {
				return u, true
			}
		}
	}
	return types.User{}, false
}

func (db *DB) VerifyRecaptcha(ctx context.Context, resp string) bool {
	defer trace.StartRegion(ctx, "DB/VerifyRecaptcha").End()

	if !db.Settings.ReCaptcha {
		return true
	}
	// TODO: Add context support to recaptcha-go.
	if err := captcha.Verify(resp); err != nil {
		return false
	}
	return true
}
