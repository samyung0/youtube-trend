package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type YouTubeClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewYouTubeClient(apiKey string) *YouTubeClient {
	return &YouTubeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type YTVideo struct {
	ID      string `json:"id"`
	Snippet struct {
		Title        string `json:"title"`
		ChannelTitle string `json:"channelTitle"`
		ChannelID    string `json:"channelId"`
		Thumbnails   struct {
			Default  YTThumb  `json:"default"`
			Medium   YTThumb  `json:"medium"`
			High     YTThumb  `json:"high"`
			Standard *YTThumb `json:"standard"`
			Maxres   *YTThumb `json:"maxres"`
		} `json:"thumbnails"`
		CategoryID  string   `json:"categoryId"`
		Tags        []string `json:"tags"`
		PublishedAt string   `json:"publishedAt"`
	} `json:"snippet"`
	Statistics struct {
		ViewCount    string `json:"viewCount"`
		LikeCount    string `json:"likeCount"`
		CommentCount string `json:"commentCount"`
	} `json:"statistics"`
	ContentDetails struct {
		Duration string `json:"duration"`
	} `json:"contentDetails"`
}

type YTThumb struct {
	URL string `json:"url"`
}

type YTChannelInfo struct {
	ID      string `json:"id"`
	Snippet struct {
		Title      string `json:"title"`
		Thumbnails struct {
			Default YTThumb `json:"default"`
		} `json:"thumbnails"`
	} `json:"snippet"`
	Statistics struct {
		SubscriberCount string `json:"subscriberCount"`
		VideoCount      string `json:"videoCount"`
	} `json:"statistics"`
}

func (c *YouTubeClient) FetchTrending(categoryID *int, region string, maxResults int) ([]YTVideo, error) {
	params := url.Values{
		"part":       {"snippet,statistics,contentDetails"},
		"chart":      {"mostPopular"},
		"regionCode": {region},
		"maxResults": {strconv.Itoa(maxResults)},
		"key":        {c.apiKey},
	}
	if categoryID != nil && *categoryID > 0 {
		params.Set("videoCategoryId", strconv.Itoa(*categoryID))
	}

	resp, err := c.httpClient.Get("https://www.googleapis.com/youtube/v3/videos?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("youtube API error: %d", resp.StatusCode)
	}

	var result struct {
		Items []YTVideo `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *YouTubeClient) FetchChannels(channelIDs []string) ([]YTChannelInfo, error) {
	var all []YTChannelInfo
	for i := 0; i < len(channelIDs); i += 50 {
		end := i + 50
		if end > len(channelIDs) {
			end = len(channelIDs)
		}
		batch := channelIDs[i:end]

		ids := ""
		for j, id := range batch {
			if j > 0 {
				ids += ","
			}
			ids += id
		}

		params := url.Values{
			"part": {"snippet,statistics"},
			"id":   {ids},
			"key":  {c.apiKey},
		}

		resp, err := c.httpClient.Get("https://www.googleapis.com/youtube/v3/channels?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Items []YTChannelInfo `json:"items"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		all = append(all, result.Items...)
	}
	return all, nil
}

func (v *YTVideo) BestThumbnail() string {
	if v.Snippet.Thumbnails.Maxres != nil {
		return v.Snippet.Thumbnails.Maxres.URL
	}
	if v.Snippet.Thumbnails.Standard != nil {
		return v.Snippet.Thumbnails.Standard.URL
	}
	return v.Snippet.Thumbnails.High.URL
}

func (v *YTVideo) ValidateThumbnail() string {
	preferred := v.BestThumbnail()
	if checkURL(preferred) {
		return preferred
	}
	candidates := []string{
		fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", v.ID),
		fmt.Sprintf("https://i.ytimg.com/vi/%s/sddefault.jpg", v.ID),
		fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", v.ID),
		fmt.Sprintf("https://i.ytimg.com/vi/%s/mqdefault.jpg", v.ID),
	}
	for _, u := range candidates {
		if u != preferred && checkURL(u) {
			return u
		}
	}
	return ""
}

func checkURL(u string) bool {
	resp, err := http.Head(u)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func ParseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func ParseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
