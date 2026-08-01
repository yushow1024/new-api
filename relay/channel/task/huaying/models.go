package huaying

import "encoding/json"

const (
	ChannelName         = "huaying-video"
	HongNiaoChannelName = "hongniao-video"

	huayingSubmitPath  = "/videos/generations"
	hongNiaoSubmitPath = "/videos"
	xingHeSubmitPath   = "/api/generate-video"
)

// VideoGenerationRequest matches the unified public parameters accepted by
// POST /v1/videos/generations. Duration and Seconds use RawMessage so numeric
// values and strings such as "5s" are both preserved during normalization.
type VideoGenerationRequest struct {
	Model       string          `json:"model"`
	RequestID   *string         `json:"requestId,omitempty"`
	Prompt      string          `json:"prompt"`
	Images      []string        `json:"images,omitempty"`
	Videos      []string        `json:"videos,omitempty"`
	Audios      []string        `json:"audios,omitempty"`
	Duration    json.RawMessage `json:"duration,omitempty"`
	Resolution  string          `json:"resolution,omitempty"`
	AspectRatio string          `json:"aspectRatio,omitempty"`
	FirstFrame  *string         `json:"firstFrame,omitempty"`
	LastFrame   *string         `json:"lastFrame,omitempty"`

	// XingHe-compatible aliases.
	Seconds json.RawMessage `json:"seconds,omitempty"`
	Size    string          `json:"size,omitempty"`

	// HongNiao-compatible aliases. They are accepted on the unified creation
	// endpoint and normalized before validation and upstream forwarding.
	AspectRatioSnake *string  `json:"aspect_ratio,omitempty"`
	ImageURLs        []string `json:"image_urls,omitempty"`
	VideoURLs        []string `json:"video_urls,omitempty"`
	VideoURL         *string  `json:"video_url,omitempty"`
	AudioURLs        []string `json:"audio_urls,omitempty"`
	AudioURL         *string  `json:"audio_url,omitempty"`
}
