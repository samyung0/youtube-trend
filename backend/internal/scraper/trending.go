package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thumbtrend/backend/internal/model"
	"github.com/thumbtrend/backend/internal/store"
)

var CategoryIDs = []int{0, 1, 10, 17, 20, 22, 23, 24, 25, 26, 28}

const (
	FetchCount = 80
	MaxValid   = 50
	Region     = "US"
)

func RunFetchTrending(ctx context.Context, yt *YouTubeClient, vs *store.VideoStore, cs *store.ChannelStore) error {
	log.Printf("Fetching trending videos for %d categories...", len(CategoryIDs))
	totalInserted := 0

	for _, catID := range CategoryIDs {
		catLabel := "All"
		if catID > 0 {
			catLabel = fmt.Sprintf("Category %d", catID)
		}
		log.Printf("\n--- %s ---", catLabel)

		var catPtr *int
		if catID > 0 {
			c := catID
			catPtr = &c
		}

		raw, err := yt.FetchTrending(catPtr, Region, FetchCount)
		if err != nil {
			log.Printf("  Error fetching %s: %v", catLabel, err)
			continue
		}
		if len(raw) == 0 {
			log.Printf("  No videos found")
			continue
		}

		type validVideo struct {
			yt       YTVideo
			thumbURL string
		}
		var validated []validVideo
		skipped := 0

		for _, v := range raw {
			if len(validated) >= MaxValid {
				break
			}
			thumbURL := v.ValidateThumbnail()
			if thumbURL != "" {
				validated = append(validated, validVideo{yt: v, thumbURL: thumbURL})
			} else {
				skipped++
				log.Printf("  Skipped %s (no working thumbnail)", v.ID)
			}
		}

		if skipped > 0 {
			log.Printf("  Validated: %d/%d (%d broken)", len(validated), len(raw), skipped)
		}
		if len(validated) == 0 {
			continue
		}

		// Create snapshot
		snap := &model.TrendingSnapshot{
			Region:     Region,
			CategoryID: catPtr,
			VideoCount: len(validated),
		}
		snapID, err := vs.CreateSnapshot(ctx, snap)
		if err != nil {
			log.Printf("  Error creating snapshot: %v", err)
			continue
		}

		var videoIDs []int64
		channelsSeen := map[string]bool{}

		for _, vv := range validated {
			v := vv.yt
			pubTime, _ := time.Parse(time.RFC3339, v.Snippet.PublishedAt)

			vid := &model.Video{
				YouTubeID:    v.ID,
				Title:        v.Snippet.Title,
				ChannelName:  v.Snippet.ChannelTitle,
				ChannelID:    v.Snippet.ChannelID,
				ThumbnailURL: vv.thumbURL,
				ViewCount:    ParseInt64(v.Statistics.ViewCount),
				LikeCount:    ParseInt64(v.Statistics.LikeCount),
				CommentCount: ParseInt64(v.Statistics.CommentCount),
				CategoryID:   ParseInt(v.Snippet.CategoryID),
				Tags:         v.Snippet.Tags,
				PublishedAt:  &pubTime,
				Duration:     v.ContentDetails.Duration,
			}
			if len(vid.Tags) > 20 {
				vid.Tags = vid.Tags[:20]
			}
			if vid.Tags == nil {
				vid.Tags = []string{}
			}

			dbID, err := vs.UpsertVideo(ctx, vid)
			if err != nil {
				log.Printf("  Error upserting video %s: %v", v.ID, err)
				continue
			}
			videoIDs = append(videoIDs, dbID)

			// Track channel
			if !channelsSeen[v.Snippet.ChannelID] {
				channelsSeen[v.Snippet.ChannelID] = true
				ch := &model.Channel{
					YouTubeChannelID: v.Snippet.ChannelID,
					Name:             v.Snippet.ChannelTitle,
				}
				chID, err := cs.UpsertChannel(ctx, ch)
				if err == nil {
					vs.LinkVideoChannel(ctx, dbID, chID)
				}
			}
		}

		if err := vs.LinkSnapshotVideos(ctx, snapID, videoIDs); err != nil {
			log.Printf("  Error linking snapshot videos: %v", err)
		}

		log.Printf("  Stored %d videos, snapshot #%d", len(videoIDs), snapID)
		totalInserted += len(videoIDs)

		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("\nDone. Total videos processed: %d", totalInserted)
	return nil
}
