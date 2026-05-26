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
	22: "People & Blogs", 24: "Entertainment",
	25: "News & Politics", 26: "How-to & Style", 28: "Science & Tech",
}

type topicCluster struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	VideoIndices []int  `json:"video_indices"`
}

func RunClusterTopics(ctx context.Context, apiKey string, vs *store.VideoStore, ts *store.TopicStore) error {
	log.Println("Fetching videos not yet topic-clustered...")

	videos, err := vs.GetUnclustered(ctx)
	if err != nil {
		return fmt.Errorf("fetch unclustered videos: %w", err)
	}
	if len(videos) == 0 {
		log.Println("No unclustered videos. Skipping.")
		return nil
	}

	log.Printf("Found %d videos. Clustering...", len(videos))

	existingTopics, err := ts.ListAllTopics(ctx)
	if err != nil {
		return fmt.Errorf("list existing topics: %w", err)
	}
	existingCatalog := formatExistingGenres(existingTopics)
	canonicalNames := make(map[string]string, len(existingTopics))
	for _, t := range existingTopics {
		canonicalNames[strings.ToLower(t.Name)] = t.Name
	}
	if len(existingTopics) > 0 {
		log.Printf("Loaded %d existing micro-genres for reuse", len(existingTopics))
	}

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
		clusters, err := callDeepSeek(ctx, apiKey, strings.Join(lines, "\n"), len(batch), existingCatalog)
		if err != nil {
			log.Printf("  LLM error: %v", err)
			continue
		}

		for i := range clusters {
			if canon, ok := canonicalNames[strings.ToLower(clusters[i].Name)]; ok {
				clusters[i].Name = canon
			}
			c := clusters[i]
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

func formatExistingGenres(topics []model.Topic) string {
	if len(topics) == 0 {
		return "(none yet — you may define new micro-genres)"
	}
	var b strings.Builder
	for _, t := range topics {
		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}
		fmt.Fprintf(&b, "- %q", t.Name)
		if desc != "" {
			fmt.Fprintf(&b, ": %s", desc)
		}
		fmt.Fprintf(&b, " (color: %s)\n", t.Color)
	}
	return b.String()
}

func callDeepSeek(ctx context.Context, apiKey, videoList string, count int, existingCatalog string) ([]topicCluster, error) {
	body := map[string]interface{}{
		"model":           "deepseek-v4-flash",
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You analyze YouTube trending video titles and tags to assign micro-genres.
Return JSON: { "genres": [{ "name": "Short catchy name (2-4 words)", "description": "One sentence description", "color": "#hex color for UI", "video_indices": [indices] }] }

Existing micro-genres (catalog):
` + existingCatalog + `
Rules:
- Assign every video to exactly one genre
- Prefer existing genres: if a video fits an existing genre, use that genre's exact name (character-for-character) and color
- Only add a new genre when no existing genre fits the video
- When adding new genres, keep names specific and trendy (e.g. "Cozy Farming Sims" not "Gaming"); use vibrant distinct colors
- Do not create near-duplicate names for the same concept as an existing genre (synonyms, plural tweaks, reorderings)`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Assign these %d trending YouTube videos to micro-genres (reuse existing names when they fit):\n\n%s", count, videoList),
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
