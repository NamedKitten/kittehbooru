package database

import (
	"context"
	"fmt"
	"runtime/trace"
	"time"

	"github.com/rs/zerolog/log"
)

// Thumbnail Scanner scans for posts missing thumbnails every 15 minutes
func (db *DB) thumbnailScanner() {
	for {
		ctx, task := trace.NewTask(context.Background(), "thumbnailScanner")
		log.Info().Msg("thumbnailScanner scanning for posts with missing thumbnials.")
		posts := db.cacheSearch(ctx, []string{"*"})
		for _, postID := range posts {
			// If the thumbnail doesn't exist, generate a new thumbnail
			if !db.ThumbnailsStorage.Exists(ctx, fmt.Sprintf("%d.webp", postID)) {
				log.Debug().Int64("postID", postID).Msg("Missing thumbnail, generating new.")
				p, err := db.Post(ctx, postID)
				if err != nil {
					log.Debug().Err(err).Int64("postID", postID).Msg("thumbnailScanner can't fetch post.")
					continue
				}
				db.CreateThumbnail(ctx, p)
			}
		}
		task.End()
		time.Sleep(time.Minute * 15)
	}
}
