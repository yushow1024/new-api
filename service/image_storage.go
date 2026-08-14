package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

var imageBase64Fields = []string{
	"b64_json",
	"base64",
	"b64",
	"image_base64",
	"imageBase64",
}

// PersistImageResponse uploads every generated image in an OpenAI-compatible
// image response and replaces url/base64 fields with the persisted public URL.
// Unknown response fields are retained as raw JSON. A failed item is kept
// unchanged so one unavailable image does not prevent other images from being
// persisted.
func PersistImageResponse(ctx context.Context, body []byte, proxy string) ([]byte, error) {
	var response map[string]json.RawMessage
	if err := common.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode image response failed: %w", err)
	}

	rawData, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("image response does not contain data")
	}
	var items []map[string]json.RawMessage
	if err := common.Unmarshal(rawData, &items); err != nil {
		return nil, fmt.Errorf("decode image response data failed: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("image response data is empty")
	}

	changed := false
	for i := range items {
		persistedURL, err := persistImageItem(ctx, items[i], proxy)
		if err != nil {
			common.SysError(fmt.Sprintf("persist generated image %d failed, keep this image unchanged: %v", i, err))
			continue
		}
		rawURL, err := common.Marshal(persistedURL)
		if err != nil {
			return nil, fmt.Errorf("encode persisted image URL failed: %w", err)
		}
		items[i]["url"] = rawURL
		for _, field := range imageBase64Fields {
			delete(items[i], field)
		}
		changed = true
	}

	if !changed {
		return append([]byte(nil), body...), nil
	}

	encodedData, err := common.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("encode image response data failed: %w", err)
	}
	response["data"] = encodedData
	result, err := common.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("encode image response failed: %w", err)
	}
	return result, nil
}

func persistImageItem(ctx context.Context, item map[string]json.RawMessage, proxy string) (string, error) {
	var sourceErrors []error
	foundSource := false

	if rawURL, ok := item["url"]; ok {
		foundSource = true
		var imageURL string
		if err := common.Unmarshal(rawURL, &imageURL); err != nil {
			sourceErrors = append(sourceErrors, fmt.Errorf("decode image URL failed: %w", err))
		} else if imageURL = strings.TrimSpace(imageURL); imageURL != "" {
			persistedURL, err := persistImageValue(ctx, imageURL, proxy, true)
			if err == nil {
				return persistedURL, nil
			}
			sourceErrors = append(sourceErrors, fmt.Errorf("persist image URL failed: %w", err))
		}
	}

	// Some adaptors return both url and b64_json. If downloading the temporary
	// URL fails, the Base64 payload is still a complete image and must be used
	// as the fallback source instead of abandoning persistence.
	for _, field := range imageBase64Fields {
		rawBase64, ok := item[field]
		if !ok {
			continue
		}
		foundSource = true

		var encoded string
		if err := common.Unmarshal(rawBase64, &encoded); err != nil {
			sourceErrors = append(sourceErrors, fmt.Errorf("decode image %s field failed: %w", field, err))
			continue
		}
		encoded = strings.TrimSpace(encoded)
		if encoded == "" {
			continue
		}

		persistedURL, err := persistImageValue(ctx, encoded, proxy, false)
		if err == nil {
			return persistedURL, nil
		}
		sourceErrors = append(sourceErrors, fmt.Errorf("persist image %s field failed: %w", field, err))
	}

	if len(sourceErrors) > 0 {
		return "", errors.Join(sourceErrors...)
	}
	if foundSource {
		return "", fmt.Errorf("image item contains only empty image sources")
	}
	return "", fmt.Errorf("image item contains neither url nor base64 data")
}

func persistImageValue(ctx context.Context, value, proxy string, fromURLField bool) (string, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "data:") {
		return SaveDataURL(ctx, FileCategoryImage, value, "image")
	}
	if fromURLField && (strings.HasPrefix(strings.ToLower(value), "http://") || strings.HasPrefix(strings.ToLower(value), "https://")) {
		return SaveRemoteWithOptions(ctx, FileCategoryImage, value, nil, proxy)
	}
	return SaveBase64(ctx, FileCategoryImage, value, "image", "")
}
