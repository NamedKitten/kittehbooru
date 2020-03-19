package database

import (
	"database/sql"
	"io/ioutil"
	"os"
	"time"

	"github.com/NamedKitten/kittehbooru/storage"
	"gopkg.in/yaml.v2"

	"github.com/NamedKitten/kittehbooru/types"
	"github.com/bwmarrin/snowflake"
	"github.com/ezzarghili/recaptcha-go"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// database-gobal captcha manager
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
	sqldb      *sql.DB
	configFile string `yaml:"-"`
	// SetupCompleted is used to know when to run setup page.
	SetupCompleted bool `yaml:"init"`
	// Settings contains instance-specific settings for this instance.
	Settings          Settings      `yaml:"settings"`
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
	err = ioutil.WriteFile(db.configFile, data, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Can't save settings")
	}
}

// init creates all the database fields and starts cache, thumbnail and session management
func (db *DB) init() {
	snowflake.Epoch = 1551864242
	var err error

	db.ContentStorage = storage.GetStorage(db.Settings.ContentStorage)
	db.ThumbnailsStorage = storage.GetStorage(db.Settings.ThumbnailsStorage)

	db.sqldb, err = sql.Open(db.Settings.DatabaseType, db.Settings.DatabaseURI)
	if err != nil {
		log.Warn().Err(err).Msg("SQL Open")
		panic(err)
	}

	db.sqlInit()

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
	go db.thumbnailScanner()
	go db.sessionCleaner()
}

// LoadDB loads the settings file and initializes the database
func LoadDB(configFile string) *DB {
	db := &DB{}
	_, err := os.Stat(configFile)
	if err != nil {
		if err != nil {
			log.Fatal().Err(err).Msg("Please copy over settings_example.yaml to settings.yaml.")
			panic(err)
		}
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Error().Err(err).Msg("Can't read settings")
	}

	err = yaml.Unmarshal(data, &db)
	if err != nil {
		log.Error().Err(err).Msg("Can't unmarshal settings")
		db = &DB{}
	}
	db.configFile = configFile
	db.init()
	db.Save()
	return db
}
