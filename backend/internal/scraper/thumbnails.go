package scraper

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/thumbtrend/backend/internal/model"
	"github.com/thumbtrend/backend/internal/store"
)

const (
	AnalyzeBatchSize = 20
	AnalyzeDelay     = 150 * time.Millisecond
)

func RunAnalyzeThumbnails(ctx context.Context, vs *store.VideoStore, as *store.AnalysisStore) error {
	log.Println("Finding unanalyzed videos...")

	since := time.Now().Add(-24 * time.Hour)
	videos, err := as.GetUnanalyzedVideos(ctx, since, 100)
	if err != nil {
		return fmt.Errorf("fetch unanalyzed: %w", err)
	}
	if len(videos) == 0 {
		log.Println("All recent videos already analyzed.")
		return nil
	}

	log.Printf("Analyzing %d thumbnails...", len(videos))
	success := 0

	for i := 0; i < len(videos); i += AnalyzeBatchSize {
		end := i + AnalyzeBatchSize
		if end > len(videos) {
			end = len(videos)
		}
		batch := videos[i:end]
		log.Printf("  Batch %d/%d", (i/AnalyzeBatchSize)+1, (len(videos)+AnalyzeBatchSize-1)/AnalyzeBatchSize)

		for _, v := range batch {
			img, err := downloadAndDecode(v.ThumbnailURL)
			if err != nil {
				log.Printf("    Error %s: %v", v.YouTubeID, err)
				time.Sleep(AnalyzeDelay)
				continue
			}

			colors := extractDominantColors(img, 5)
			brightness := calcBrightness(img)
			hasFace, faceCount := detectFaceHeuristic(img)

			analysis := &model.ThumbnailAnalysis{
				VideoID:        v.ID,
				DominantColors: colors,
				HasFace:        hasFace,
				FaceCount:      faceCount,
				Brightness:     brightness,
			}

			if err := as.Upsert(ctx, analysis); err != nil {
				log.Printf("    DB error %s: %v", v.YouTubeID, err)
			} else {
				success++
				faceLabel := "no-face"
				if hasFace {
					faceLabel = fmt.Sprintf("faces:%d", faceCount)
				}
				log.Printf("    %s: colors=%d %s brightness=%.2f", v.YouTubeID, len(colors), faceLabel, brightness)
			}
			time.Sleep(AnalyzeDelay)
		}
	}

	log.Printf("Analysis complete. %d/%d succeeded.", success, len(videos))
	return nil
}

func downloadAndDecode(u string) (image.Image, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(strings.NewReader(string(data)))
	return img, err
}

func extractDominantColors(img image.Image, n int) []string {
	bounds := img.Bounds()
	step := 4
	bucketSize := 32

	type bucket struct {
		r, g, b int
		count   int
	}
	buckets := map[string]*bucket{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			ri, gi, bi := int(r>>8), int(g>>8), int(b>>8)
			kr := (ri / bucketSize) * bucketSize
			kg := (gi / bucketSize) * bucketSize
			kb := (bi / bucketSize) * bucketSize
			key := fmt.Sprintf("%d,%d,%d", kr, kg, kb)

			if b, ok := buckets[key]; ok {
				b.r += ri
				b.g += gi
				b.b += bi
				b.count++
			} else {
				buckets[key] = &bucket{r: ri, g: gi, b: bi, count: 1}
			}
		}
	}

	sorted := make([]*bucket, 0, len(buckets))
	for _, b := range buckets {
		sorted = append(sorted, b)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	if len(sorted) > n {
		sorted = sorted[:n]
	}
	var colors []string
	for _, b := range sorted {
		colors = append(colors, fmt.Sprintf("#%02x%02x%02x", b.r/b.count, b.g/b.count, b.b/b.count))
	}
	return colors
}

func calcBrightness(img image.Image) float64 {
	bounds := img.Bounds()
	step := 8
	var total float64
	count := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			lum := (0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)) / 255.0
			total += lum
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Round(total/float64(count)*100) / 100
}

func detectFaceHeuristic(img image.Image) (bool, int) {
	bounds := img.Bounds()
	step := 4
	skinPixels := 0
	totalPixels := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			ri, gi, bi := int(r>>8), int(g>>8), int(b>>8)
			totalPixels++
			if ri > 95 && gi > 40 && bi > 20 && ri > gi && ri > bi && ri-gi > 15 && ri-bi > 15 {
				skinPixels++
			}
		}
	}

	if totalPixels == 0 {
		return false, 0
	}
	ratio := float64(skinPixels) / float64(totalPixels)
	hasFace := ratio > 0.15
	count := 0
	if hasFace {
		count = 1
		if ratio > 0.35 {
			count = 2
		}
	}
	return hasFace, count
}
