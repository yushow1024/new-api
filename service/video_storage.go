package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// PersistGeneratedVideo resolves a completed upstream video result, stores the
// actual video in the configured file storage, and returns its public URL.
func PersistGeneratedVideo(ctx context.Context, ch *model.Channel, task *model.Task, result *relaycommon.TaskInfo) (string, error) {
	if ch == nil || task == nil || result == nil {
		return "", fmt.Errorf("missing channel, task, or task result")
	}

	if source := strings.TrimSpace(result.Url); source != "" {
		if strings.HasPrefix(source, "data:") {
			return SaveDataURL(ctx, FileCategoryVideo, source, "video")
		}
		headers := generatedVideoDownloadHeaders(ch, task, source)
		return saveRemoteVideo(ctx, ch, source, headers)
	}

	if source := strings.TrimSpace(result.RemoteUrl); source != "" {
		headers := make(http.Header)
		if ch.Type == constant.ChannelTypeGemini {
			if key := taskVideoKey(ch, task); key != "" {
				headers.Set("x-goog-api-key", key)
			}
		}
		return saveRemoteVideo(ctx, ch, source, headers)
	}

	switch ch.Type {
	case constant.ChannelTypeOpenAI, constant.ChannelTypeSora:
		baseURL := strings.TrimRight(ch.GetBaseURL(), "/")
		if baseURL == "" {
			baseURL = strings.TrimRight(constant.ChannelBaseURLs[ch.Type], "/")
		}
		if baseURL == "" {
			baseURL = "https://api.openai.com"
		}
		upstreamTaskID := strings.TrimSpace(task.GetUpstreamTaskID())
		if upstreamTaskID == "" {
			return "", fmt.Errorf("upstream video task ID is empty")
		}
		headers := make(http.Header)
		if key := taskVideoKey(ch, task); key != "" {
			headers.Set("Authorization", "Bearer "+key)
		}
		source := fmt.Sprintf("%s/v1/videos/%s/content", baseURL, upstreamTaskID)
		return saveRemoteVideo(ctx, ch, source, headers)
	default:
		return "", fmt.Errorf("completed video task does not contain a downloadable URL")
	}
}

// PersistGeneratedVideoMedia stores both the generated video and its optional
// cover image. The returned URLs always point to the configured file storage.
func PersistGeneratedVideoMedia(ctx context.Context, ch *model.Channel, task *model.Task, result *relaycommon.TaskInfo) (string, string, error) {
	videoURL, err := PersistGeneratedVideo(ctx, ch, task, result)
	if err != nil {
		return "", "", err
	}
	coverURL, err := PersistGeneratedVideoCover(ctx, ch, task, result)
	if err != nil {
		return "", "", fmt.Errorf("persist generated video cover failed: %w", err)
	}
	return videoURL, coverURL, nil
}

// PersistGeneratedVideoMediaBestEffort attempts to persist generated media.
// If persistence fails, it returns the original upstream URLs so callers can
// keep the response and task result unchanged instead of failing the task.
func PersistGeneratedVideoMediaBestEffort(ctx context.Context, ch *model.Channel, task *model.Task, result *relaycommon.TaskInfo) (videoURL, coverURL string, persisted bool, err error) {
	videoURL, coverURL = originalGeneratedVideoMediaURLs(task, result)
	persistedVideoURL, persistedCoverURL, persistErr := PersistGeneratedVideoMedia(ctx, ch, task, result)
	if persistErr != nil {
		return videoURL, coverURL, false, persistErr
	}
	return persistedVideoURL, persistedCoverURL, true, nil
}

func originalGeneratedVideoMediaURLs(task *model.Task, result *relaycommon.TaskInfo) (string, string) {
	videoURL := ""
	coverURL := ""
	if result != nil {
		videoURL = strings.TrimSpace(result.Url)
		if videoURL == "" {
			videoURL = strings.TrimSpace(result.RemoteUrl)
		}
		coverURL = strings.TrimSpace(result.CoverUrl)
	}
	if videoURL == "" && task != nil {
		videoURL = strings.TrimSpace(task.GetResultURL())
	}
	return videoURL, coverURL
}

// PersistGeneratedVideoCover stores an optional cover URL or base64 payload in
// the generated-image directory. An empty cover is valid and returns no URL.
func PersistGeneratedVideoCover(ctx context.Context, ch *model.Channel, task *model.Task, result *relaycommon.TaskInfo) (string, error) {
	if ch == nil || task == nil || result == nil {
		return "", fmt.Errorf("missing channel, task, or task result")
	}
	source := strings.TrimSpace(result.CoverUrl)
	if source == "" {
		return "", nil
	}
	if strings.HasPrefix(source, "data:") {
		return SaveDataURL(ctx, FileCategoryImage, source, "cover")
	}
	if parsed, err := url.Parse(source); err == nil && parsed != nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		headers := generatedVideoCoverDownloadHeaders(ch, task, source)
		proxy := ch.GetSetting().Proxy
		return SaveRemoteWithOptions(ctx, FileCategoryImage, source, headers, proxy)
	}
	return SaveBase64(ctx, FileCategoryImage, source, "cover", "")
}

// generatedVideoDownloadHeaders builds headers for downloading a generated video.
// A completed task may return a URL hosted by an unrelated CDN. The channel key
// must never be sent to that CDN, so Authorization is only added when the
// result URL and the channel base URL have the same HTTP origin.
func generatedVideoDownloadHeaders(ch *model.Channel, task *model.Task, source string) http.Header {
	return generatedMediaDownloadHeaders(ch, task, source, "video/mp4,video/*;q=0.9,application/octet-stream;q=0.8,*/*;q=0.5")
}

func generatedVideoCoverDownloadHeaders(ch *model.Channel, task *model.Task, source string) http.Header {
	return generatedMediaDownloadHeaders(ch, task, source, "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.9,*/*;q=0.5")
}

func generatedMediaDownloadHeaders(ch *model.Channel, task *model.Task, source, accept string) http.Header {
	if ch == nil || !constant.IsDedicatedVideoPollingChannel(ch.Type) {
		return nil
	}

	// These headers are safe for third-party CDNs; never forward the channel key
	// unless the media URL has the same origin as the channel endpoint.
	headers := make(http.Header)
	headers.Set("Accept", accept)
	headers.Set("User-Agent", "Mozilla/5.0 (compatible; new-api-media-fetcher/1.0)")
	if sameURLOrigin(source, ch.GetBaseURL()) {
		if key := taskVideoKey(ch, task); key != "" {
			headers.Set("Authorization", "Bearer "+key)
		}
	}
	return headers
}

func sameURLOrigin(left, right string) bool {
	leftURL, leftErr := url.Parse(strings.TrimSpace(left))
	rightURL, rightErr := url.Parse(strings.TrimSpace(right))
	if leftErr != nil || rightErr != nil || leftURL == nil || rightURL == nil {
		return false
	}
	if (leftURL.Scheme != "http" && leftURL.Scheme != "https") ||
		!strings.EqualFold(leftURL.Scheme, rightURL.Scheme) ||
		leftURL.Host == "" || rightURL.Host == "" {
		return false
	}
	return strings.EqualFold(leftURL.Host, rightURL.Host)
}
func saveRemoteVideo(ctx context.Context, ch *model.Channel, source string, headers http.Header) (string, error) {
	proxy := ""
	if ch != nil {
		proxy = ch.GetSetting().Proxy
	}
	return SaveRemoteWithOptions(ctx, FileCategoryVideo, source, headers, proxy)
}

func taskVideoKey(ch *model.Channel, task *model.Task) string {
	if task != nil {
		if key := strings.TrimSpace(task.PrivateData.Key); key != "" {
			return key
		}
	}
	if ch == nil {
		return ""
	}
	for _, key := range ch.GetKeys() {
		if key = strings.TrimSpace(key); key != "" {
			return key
		}
	}
	return strings.TrimSpace(ch.Key)
}

func UpdateGeneratedVideoLog(task *model.Task) error {
	if task == nil || task.LogId <= 0 {
		return nil
	}
	body, err := common.Marshal(task.ToOpenAIVideo())
	if err != nil {
		return fmt.Errorf("encode generated video log response failed: %w", err)
	}
	return model.UpdateLogResData(task.LogId, model.LogData(body))
}

func OverrideOpenAIVideoURL(body []byte, resultURL string) ([]byte, error) {
	return OverrideOpenAIVideoMediaURLs(body, resultURL, "")
}

func OverrideOpenAIVideoMediaURLs(body []byte, resultURL, coverURL string) ([]byte, error) {
	var video dto.OpenAIVideo
	if err := common.Unmarshal(body, &video); err != nil {
		return nil, fmt.Errorf("decode OpenAI video response failed: %w", err)
	}
	if strings.TrimSpace(resultURL) != "" {
		video.SetMetadata("url", resultURL)
	}
	if strings.TrimSpace(coverURL) != "" {
		video.SetMetadata("cover_url", coverURL)
	}
	result, err := common.Marshal(video)
	if err != nil {
		return nil, fmt.Errorf("encode OpenAI video response failed: %w", err)
	}
	return result, nil
}

// OverrideGeneratedVideoResponseMediaURLs preserves the upstream response while
// replacing both generated video and cover fields with persisted public URLs.
func OverrideGeneratedVideoResponseMediaURLs(body []byte, resultURL, coverURL string) ([]byte, error) {
	updated, err := OverrideGeneratedVideoResponseURL(body, resultURL)
	if err != nil {
		return nil, err
	}
	return OverrideGeneratedVideoCoverURL(updated, coverURL)
}

// OverrideGeneratedVideoResponseURL preserves the upstream JSON response while
// replacing generated video URL fields with the persisted public URL.
func OverrideGeneratedVideoResponseURL(body []byte, resultURL string) ([]byte, error) {
	if strings.TrimSpace(resultURL) == "" {
		return append([]byte(nil), body...), nil
	}
	updated, replaced, err := overrideGeneratedVideoURLValue(json.RawMessage(body), resultURL)
	if err != nil {
		return nil, fmt.Errorf("override generated video response URL failed: %w", err)
	}
	if replaced {
		return updated, nil
	}

	var response map[string]json.RawMessage
	if err := common.Unmarshal(updated, &response); err != nil {
		return nil, fmt.Errorf("decode generated video response failed: %w", err)
	}
	rawURL, err := common.Marshal(resultURL)
	if err != nil {
		return nil, fmt.Errorf("encode generated video URL failed: %w", err)
	}
	response["url"] = rawURL
	result, err := common.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("encode generated video response failed: %w", err)
	}
	return result, nil
}

func OverrideGeneratedVideoCoverURL(body []byte, coverURL string) ([]byte, error) {
	if strings.TrimSpace(coverURL) == "" {
		return append([]byte(nil), body...), nil
	}
	updated, replaced, err := overrideGeneratedVideoCoverURLValue(json.RawMessage(body), coverURL)
	if err != nil {
		return nil, fmt.Errorf("override generated video cover URL failed: %w", err)
	}
	if replaced {
		return updated, nil
	}

	var response map[string]json.RawMessage
	if err := common.Unmarshal(updated, &response); err != nil {
		return nil, fmt.Errorf("decode generated video response failed: %w", err)
	}
	rawURL, err := common.Marshal(coverURL)
	if err != nil {
		return nil, fmt.Errorf("encode generated video cover URL failed: %w", err)
	}
	response["cover_url"] = rawURL
	result, err := common.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("encode generated video response failed: %w", err)
	}
	return result, nil
}

func overrideGeneratedVideoURLValue(raw json.RawMessage, resultURL string) (json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return raw, false, nil
	}

	switch trimmed[0] {
	case '{':
		var object map[string]json.RawMessage
		if err := common.Unmarshal(trimmed, &object); err != nil {
			return nil, false, err
		}
		replaced := false
		for key, value := range object {
			if isGeneratedVideoURLKey(key) {
				updated, fieldReplaced, err := overrideGeneratedVideoURLField(value, resultURL)
				if err != nil {
					return nil, false, err
				}
				if fieldReplaced {
					object[key] = updated
					replaced = true
					continue
				}
			}
			updated, childReplaced, err := overrideGeneratedVideoURLValue(value, resultURL)
			if err != nil {
				return nil, false, err
			}
			if childReplaced {
				object[key] = updated
				replaced = true
			}
		}
		encoded, err := common.Marshal(object)
		return encoded, replaced, err
	case '[':
		var array []json.RawMessage
		if err := common.Unmarshal(trimmed, &array); err != nil {
			return nil, false, err
		}
		replaced := false
		for i := range array {
			updated, childReplaced, err := overrideGeneratedVideoURLValue(array[i], resultURL)
			if err != nil {
				return nil, false, err
			}
			if childReplaced {
				array[i] = updated
				replaced = true
			}
		}
		encoded, err := common.Marshal(array)
		return encoded, replaced, err
	default:
		return append(json.RawMessage(nil), raw...), false, nil
	}
}

func overrideGeneratedVideoURLField(raw json.RawMessage, resultURL string) (json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return raw, false, nil
	}
	if trimmed[0] == '"' {
		var current string
		if err := common.Unmarshal(trimmed, &current); err != nil || strings.TrimSpace(current) == "" {
			return raw, false, nil
		}
		replacement, err := common.Marshal(resultURL)
		return replacement, true, err
	}
	if trimmed[0] == '[' {
		var array []json.RawMessage
		if err := common.Unmarshal(trimmed, &array); err != nil {
			return nil, false, err
		}
		replaced := false
		for i := range array {
			updated, itemReplaced, err := overrideGeneratedVideoURLField(array[i], resultURL)
			if err != nil {
				return nil, false, err
			}
			if itemReplaced {
				array[i] = updated
				replaced = true
			}
		}
		if !replaced {
			return raw, false, nil
		}
		encoded, err := common.Marshal(array)
		return encoded, true, err
	}
	return raw, false, nil
}

func overrideGeneratedVideoCoverURLValue(raw json.RawMessage, coverURL string) (json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return raw, false, nil
	}

	switch trimmed[0] {
	case '{':
		var object map[string]json.RawMessage
		if err := common.Unmarshal(trimmed, &object); err != nil {
			return nil, false, err
		}
		replaced := false
		for key, value := range object {
			if isGeneratedVideoCoverURLKey(key) {
				var current string
				if err := common.Unmarshal(value, &current); err == nil && strings.TrimSpace(current) != "" {
					replacement, err := common.Marshal(coverURL)
					if err != nil {
						return nil, false, err
					}
					object[key] = replacement
					replaced = true
					continue
				}
			}
			updated, childReplaced, err := overrideGeneratedVideoCoverURLValue(value, coverURL)
			if err != nil {
				return nil, false, err
			}
			if childReplaced {
				object[key] = updated
				replaced = true
			}
		}
		encoded, err := common.Marshal(object)
		return encoded, replaced, err
	case '[':
		var array []json.RawMessage
		if err := common.Unmarshal(trimmed, &array); err != nil {
			return nil, false, err
		}
		replaced := false
		for i := range array {
			updated, childReplaced, err := overrideGeneratedVideoCoverURLValue(array[i], coverURL)
			if err != nil {
				return nil, false, err
			}
			if childReplaced {
				array[i] = updated
				replaced = true
			}
		}
		encoded, err := common.Marshal(array)
		return encoded, replaced, err
	default:
		return append(json.RawMessage(nil), raw...), false, nil
	}
}

func isGeneratedVideoCoverURLKey(key string) bool {
	switch key {
	case "cover_url", "coverUrl", "poster_url", "posterUrl", "thumbnail_url", "thumbnailUrl", "cover_base64", "coverBase64":
		return true
	default:
		return false
	}
}

func isGeneratedVideoURLKey(key string) bool {
	switch key {
	case "url", "video_url", "videoUrl", "output_url", "outputUrl", "download_url", "downloadUrl",
		"outputUrls", "outputs", "videoUrls", "resultUrls", "output_urls", "video_urls", "result_urls":
		return true
	default:
		return false
	}
}
