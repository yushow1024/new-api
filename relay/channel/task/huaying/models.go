package huaying

import "encoding/json"

const (
	ChannelName = "huaying-video"

	huayingSubmitPath = "/videos/generations"
	xingHeSubmitPath  = "/api/generate-video"
)

// VideoGenerationRequest matches the public parameters of Huaying /v1/videos/generations.
// Duration uses RawMessage so both numeric values and strings such as "5s" are preserved.
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

	// Seconds and Size are XingHe-compatible aliases. The public contract remains
	// Huaying-compatible; accepting these aliases keeps existing XingHe callers
	// working and they are normalized before validation and upstream forwarding.
	Seconds json.RawMessage `json:"seconds,omitempty"`
	Size    string          `json:"size,omitempty"`
}
