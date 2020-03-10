package database

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// Thumbnail Scanner scans for posts missing thumbnails every 15 minutes
func (db *DB) thumbnailScanner() {
	for {

		ctx := context.TODO()
		log.Info().Msg("thumbnailScanner scanning for posts with missing thumbnials.")

		rows, err := db.sqldb.Query(`select "postid" from posts where true`)
		if err != nil {
			log.Error().Err(err).Msg("thumbnailScanner can't select all posts")
		}
		defer rows.Close()
		var postID int64
		for rows.Next() {
			err = rows.Scan(&postID)
			if err != nil {
				log.Error().Err(err).Msg("thumbnailScanner can't scan row")
				break
			}
			if !db.ThumbnailsStorage.Exists(ctx, fmt.Sprintf("%d.webp", postID)) {
				log.Debug().Int64("postID", postID).Msg("Missing thumbnail, generating new.")
				p, exists := db.Post(ctx, postID)
				if !exists {
					log.Debug().Int64("postID", postID).Msg("thumbnailScanner can't fetch post.")
					continue
				}

				db.CreateThumbnail(ctx, p)
			}
		}

		time.Sleep(time.Minute * 15)
	}

}
