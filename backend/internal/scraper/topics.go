package scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/thumbtrend/backend/internal/model"
	"github.com/thumbtrend/backend/internal/store"
)

var CategoryNames = map[int]string{
	1: "Film & Animation", 10: "Music", 17: "Sports", 20: "Gaming",
	22: "People & Blogs", 23: "Comedy", 24: "Entertainment",
	25: "News & Politics", 26: "How-to & Style", 28: "Science & Tech",
}

type topicCluster struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	VideoIndices []int  `json:"video_indices"`
}

func RunClusterTopics(ctx context.Context, apiKey string, vs *store.VideoStore, ts *store.TopicStore) error {
	log.Println("Fetching recent videos for topic clustering...")

	since := time.Now().Add(-24 * time.Hour)
	videos, err := vs.GetTrending(ctx, since, nil, 200)
	if err != nil {
		return fmt.Errorf("fetch videos: %w", err)
	}
	if len(videos) == 0 {
		log.Println("No recent videos. Skipping.")
		return nil
	}

	log.Printf("Found %d videos. Clustering...", len(videos))

	batchSize := 50
	allTopics := map[string]*struct {
		cluster  topicCluster
		parent   string
		videoIDs []int64
	}{}

	for i := 0; i < len(videos); i += batchSize {
		end := i + batchSize
		if end > len(videos) {
			end = len(videos)
		}
		batch := videos[i:end]

		var lines []string
		for j, v := range batch {
			tags := v.Tags
			if len(tags) > 8 {
				tags = tags[:8]
			}
			lines = append(lines, fmt.Sprintf("[%d] \"%s\" tags: %s", i+j, v.Title, strings.Join(tags, ", ")))
		}

		log.Printf("  Processing batch %d...", (i/batchSize)+1)
		clusters, err := callDeepSeek(ctx, apiKey, strings.Join(lines, "\n"), len(batch))
		if err != nil {
			log.Printf("  LLM error: %v", err)
			continue
		}

		for _, c := range clusters {
			slug := slugify(c.Name)
			existing, ok := allTopics[slug]
			var vids []int64
			for _, idx := range c.VideoIndices {
				if idx >= 0 && idx < len(videos) {
					vids = append(vids, videos[idx].ID)
				}
			}

			parent := ""
			if len(vids) > 0 {
				catID := videos[c.VideoIndices[0]].CategoryID
				if name, ok := CategoryNames[catID]; ok {
					parent = name
				}
			}

			if ok {
				existing.videoIDs = append(existing.videoIDs, vids...)
			} else {
				allTopics[slug] = &struct {
					cluster  topicCluster
					parent   string
					videoIDs []int64
				}{cluster: c, parent: parent, videoIDs: vids}
			}
		}
	}

	log.Printf("\nDiscovered %d topics:", len(allTopics))
	for slug, t := range allTopics {
		log.Printf("  - %s (%d videos)", t.cluster.Name, len(t.videoIDs))
		topic := &model.Topic{
			Name:           t.cluster.Name,
			Slug:           slug,
			Description:    &t.cluster.Description,
			Color:          t.cluster.Color,
			ParentCategory: &t.parent,
			SnapshotDate:   time.Now().Format("2006-01-02"),
		}
		topicID, err := ts.UpsertTopic(ctx, topic)
		if err != nil {
			log.Printf("    Error: %v", err)
			continue
		}
		ts.LinkTopicVideos(ctx, topicID, t.videoIDs)
	}

	log.Println("Topic clustering complete.")
	return nil
}

func callDeepSeek(ctx context.Context, apiKey, videoList string, count int) ([]topicCluster, error) {
	body := map[string]interface{}{
		"model":       "deepseek-v4-flash",
		"temperature": 0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You analyze YouTube trending video titles and tags to discover micro-genres.
Return JSON: { "genres": [{ "name": "Short catchy name (2-4 words)", "description": "One sentence description", "color": "#hex color for UI", "video_indices": [indices] }] }
- Create 5-8 distinct micro-genres
- Every video must be assigned to exactly one genre
- Names should be specific and trendy (e.g. "Cozy Farming Sims" not "Gaming")
- Colors should be vibrant and distinct from each other`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Cluster these %d trending YouTube videos into micro-genres:\n\n%s", count, videoList),
			},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	var parsed struct {
		Genres []topicCluster `json:"genres"`
	}
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &parsed); err != nil {
		return nil, err
	}
	return parsed.Genres, nil
}

var nonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlpha.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
