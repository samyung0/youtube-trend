package scraper

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ShortVideoMaxDuration = 5 * time.Minute

var youtubeDurPart = regexp.MustCompile(`(\d+)([HMS])`)

func parseYouTubeDuration(s string) (time.Duration, bool) {
	if !strings.HasPrefix(s, "PT") {
		return 0, false
	}
	parts := youtubeDurPart.FindAllStringSubmatch(s[2:], -1)
	if len(parts) == 0 {
		return 0, false
	}
	var total time.Duration
	for _, m := range parts {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, false
		}
		switch m[2] {
		case "H":
			total += time.Duration(n) * time.Hour
		case "M":
			total += time.Duration(n) * time.Minute
		case "S":
			total += time.Duration(n) * time.Second
		}
	}
	return total, true
}

// IsShortVideo returns true when duration is under ShortVideoMaxDuration or unparseable.
func IsShortVideo(duration string) bool {
	d, ok := parseYouTubeDuration(duration)
	return !ok || d < ShortVideoMaxDuration
}
