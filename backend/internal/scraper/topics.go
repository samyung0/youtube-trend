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
	VideoIndices []int  `json:"video_indices"`
}

var topicColors = []string{
	"#6366f1", "#ec4899", "#10b981", "#f59e0b", "#8b5cf6",
	"#06b6d4", "#ef4444", "#84cc16", "#f97316", "#14b8a6",
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

	batchSize := 50
	totalSaved := 0

	for i := 0; i < len(videos); i += batchSize {
		end := i + batchSize
		if end > len(videos) {
			end = len(videos)
		}
		batchNum := (i / batchSize) + 1

		existingTopics, err := ts.ListAllTopics(ctx)
		if err != nil {
			return fmt.Errorf("list existing topics: %w", err)
		}
		existingCatalog, canonicalNames, existingBySlug := topicLookup(existingTopics)
		log.Printf("  Batch %d: %d videos, %d existing topics in catalog", batchNum, end-i, len(existingTopics))

		var lines []string
		for j, v := range videos[i:end] {
			tags := v.Tags
			if len(tags) > 8 {
				tags = tags[:8]
			}
			lines = append(lines, fmt.Sprintf("[%d] \"%s\" tags: %s", j, v.Title, strings.Join(tags, ", ")))
		}

		clusters, err := callDeepSeek(ctx, apiKey, strings.Join(lines, "\n"), end-i, existingCatalog)
		if err != nil {
			log.Printf("  Batch %d LLM error: %v", batchNum, err)
			continue
		}

		saved, err := saveBatchTopics(ctx, ts, clusters, videos[i:end], canonicalNames, existingBySlug)
		if err != nil {
			log.Printf("  Batch %d save error: %v", batchNum, err)
			continue
		}
		totalSaved += saved
		log.Printf("  Batch %d saved %d topic assignments", batchNum, saved)
	}

	log.Printf("Topic clustering complete. Linked %d videos across batches.", totalSaved)
	return nil
}

type batchTopic struct {
	cluster  topicCluster
	parent   string
	videoIDs []int64
}

func topicLookup(topics []model.Topic) (catalog string, canonicalNames map[string]string, bySlug map[string]model.Topic) {
	canonicalNames = make(map[string]string, len(topics))
	bySlug = make(map[string]model.Topic, len(topics))
	for _, t := range topics {
		canonicalNames[strings.ToLower(t.Name)] = t.Name
		bySlug[t.Slug] = t
	}
	return formatExistingTopics(topics), canonicalNames, bySlug
}

func saveBatchTopics(
	ctx context.Context,
	ts *store.TopicStore,
	clusters []topicCluster,
	videos []model.Video,
	canonicalNames map[string]string,
	existingBySlug map[string]model.Topic,
) (int, error) {
	batchTopics := make(map[string]*batchTopic)

	for ci := range clusters {
		if canon, ok := canonicalNames[strings.ToLower(clusters[ci].Name)]; ok {
			clusters[ci].Name = canon
		}
		c := clusters[ci]
		slug := slugify(c.Name)
		if slug == "" {
			continue
		}

		var vids []int64
		for _, idx := range c.VideoIndices {
			if idx >= 0 && idx < len(videos) {
				vids = append(vids, videos[idx].ID)
			}
		}
		if len(vids) == 0 {
			continue
		}

		parent := ""
		idx := c.VideoIndices[0]
		if idx >= 0 && idx < len(videos) {
			if name, ok := CategoryNames[videos[idx].CategoryID]; ok {
				parent = name
			}
		}

		if existing, ok := batchTopics[slug]; ok {
			existing.videoIDs = append(existing.videoIDs, vids...)
		} else {
			batchTopics[slug] = &batchTopic{cluster: c, parent: parent, videoIDs: vids}
		}
	}

	linked := 0
	snapshotDate := time.Now().Format("2006-01-02")
	for slug, t := range batchTopics {
		color := topicColorForSlug(slug)
		if existing, ok := existingBySlug[slug]; ok {
			color = existing.Color
		}
		topic := &model.Topic{
			Name:           t.cluster.Name,
			Slug:           slug,
			Description:    &t.cluster.Description,
			Color:          color,
			ParentCategory: &t.parent,
			SnapshotDate:   snapshotDate,
		}
		topicID, err := ts.UpsertTopic(ctx, topic)
		if err != nil {
			return linked, fmt.Errorf("upsert %q: %w", slug, err)
		}
		if err := ts.LinkTopicVideos(ctx, topicID, t.videoIDs); err != nil {
			return linked, fmt.Errorf("link %q: %w", slug, err)
		}
		linked += len(t.videoIDs)
	}
	return linked, nil
}

func formatExistingTopics(topics []model.Topic) string {
	if len(topics) == 0 {
		return "(none yet — you may define new topics)"
	}
	var b strings.Builder
	for _, t := range topics {
		desc := ""
		// if t.Description != nil {
		// 	desc = *t.Description
		// }
		fmt.Fprintf(&b, "- %q", t.Name)
		if desc != "" {
			fmt.Fprintf(&b, ": %s", desc)
		}
		b.WriteByte('\n')
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
				"content": `You analyze YouTube trending video titles and tags to extract topics. A topic can be a person, a name, a location, a thing, etc., but it should be a concrete subject rather than a concept. Remove the actions/adjectives/fillers from the topics. Examples: rather than "Kevin Hart Roast", it is just "Kevin Hart". Rather than "Cruise News and Tips", it should be just "Cruise" Rather than "Roblox Gameplay", it should just be "Roblox"
Return JSON: { "topic": [{ "name": "A subject (1-3 words)", "description": "One sentence description", "video_indices": [indices] }] }

Existing topics:
` + existingCatalog + `
Rules:
- Assign every video to exactly one topic
- Prefer existing topics: if a video fits an existing topic, use that topic's exact name (character-for-character)
- Only add a new topic when no existing topic fits the video
- Focus on the content of the video rather than the channel or creator.`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Assign these %d trending YouTube videos to topics (reuse existing names when they fit):\n\n%s", count, videoList),
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
		Topics []topicCluster `json:"topic"`
	}
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &parsed); err != nil {
		return nil, err
	}
	return parsed.Topics, nil
}

func topicColorForSlug(slug string) string {
	var h uint32
	for i := 0; i < len(slug); i++ {
		h = h*31 + uint32(slug[i])
	}
	return topicColors[h%uint32(len(topicColors))]
}

var nonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlpha.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
