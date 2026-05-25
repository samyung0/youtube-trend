package scraper

import (
	"context"
	"log"
	"time"

	"github.com/thumbtrend/backend/internal/model"
	"github.com/thumbtrend/backend/internal/store"
)

func RunTrackChannels(ctx context.Context, yt *YouTubeClient, cs *store.ChannelStore) error {
	log.Println("Tracking channel subscriber counts...")

	channels, err := cs.GetAll(ctx)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		log.Println("No channels to track.")
		return nil
	}

	log.Printf("Fetching stats for %d channels...", len(channels))

	// Collect all YouTube channel IDs
	ytIDs := make([]string, len(channels))
	idMap := map[string]int64{}
	for i, ch := range channels {
		ytIDs[i] = ch.YouTubeChannelID
		idMap[ch.YouTubeChannelID] = ch.ID
	}

	infos, err := yt.FetchChannels(ytIDs)
	if err != nil {
		return err
	}

	updated := 0
	for _, info := range infos {
		dbID, ok := idMap[info.ID]
		if !ok {
			continue
		}

		subCount := ParseInt64(info.Statistics.SubscriberCount)
		vidCount := ParseInt(info.Statistics.VideoCount)

		// Update channel record
		ch := &model.Channel{
			ID:               dbID,
			YouTubeChannelID: info.ID,
			Name:             info.Snippet.Title,
			AvatarURL:        info.Snippet.Thumbnails.Default.URL,
			SubscriberCount:  subCount,
			VideoCount:       vidCount,
		}
		cs.UpsertChannel(ctx, ch)

		// Insert daily snapshot
		snap := &model.ChannelSnapshot{
			ChannelID:       dbID,
			SubscriberCount: subCount,
		}
		if err := cs.InsertSnapshot(ctx, snap); err != nil {
			log.Printf("  Error snapshot for %s: %v", info.ID, err)
		} else {
			updated++
		}
	}

	// Prune old snapshots (>30 days)
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	cs.PruneOldSnapshots(ctx, cutoff)

	log.Printf("Updated %d/%d channels.", updated, len(channels))
	return nil
}
