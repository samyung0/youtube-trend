package model

import "time"

type ThumbnailAnalysis struct {
	ID             int64    `json:"id" db:"id"`
	VideoID        int64    `json:"video_id" db:"video_id"`
	DominantColors []string `json:"dominant_colors" db:"dominant_colors"`
	HasFace        bool     `json:"has_face" db:"has_face"`
	FaceCount      int      `json:"face_count" db:"face_count"`
	OCRText        string   `json:"ocr_text" db:"ocr_text"`
	Brightness     float64  `json:"brightness" db:"brightness"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type AnalysisStats struct {
	TotalAnalyzed  int                `json:"total_analyzed"`
	FacePercentage int                `json:"face_percentage"`
	AvgBrightness  float64            `json:"avg_brightness"`
	AvgFaceCount   float64            `json:"avg_face_count"`
	ColorFrequency map[string]int     `json:"color_frequency"`
	OCRWords       map[string]int     `json:"ocr_words"`
}
