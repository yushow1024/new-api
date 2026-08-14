package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// PersistImageResponse uploads every generated image in an OpenAI-compatible
// image response and replaces url/b64_json with the persisted public URL.
// Unknown response fields are retained as raw JSON.
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

	for i := range items {
		persistedURL, err := persistImageItem(ctx, items[i], proxy)
		if err != nil {
			common.SysError(fmt.Sprintf("persist generated image %d failed, keep upstream response unchanged: %v", i, err))
			return append([]byte(nil), body...), nil
		}
		rawURL, err := common.Marshal(persistedURL)
		if err != nil {
			return nil, fmt.Errorf("encode persisted image URL failed: %w", err)
		}
		items[i]["url"] = rawURL
		delete(items[i], "b64_json")
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
	var imageURL string
	if rawURL, ok := item["url"]; ok {
		if err := common.Unmarshal(rawURL, &imageURL); err != nil {
			return "", fmt.Errorf("decode image URL failed: %w", err)
		}
	}
	if strings.TrimSpace(imageURL) != "" {
		if strings.HasPrefix(imageURL, "data:") {
			return SaveDataURL(ctx, FileCategoryImage, imageURL, "image")
		}
		return SaveRemoteWithOptions(ctx, FileCategoryImage, imageURL, nil, proxy)
	}

	var encoded string
	if rawBase64, ok := item["b64_json"]; ok {
		if err := common.Unmarshal(rawBase64, &encoded); err != nil {
			return "", fmt.Errorf("decode image base64 field failed: %w", err)
		}
	}
	if strings.TrimSpace(encoded) == "" {
		return "", fmt.Errorf("image item contains neither url nor b64_json")
	}
	if strings.HasPrefix(encoded, "data:") {
		return SaveDataURL(ctx, FileCategoryImage, encoded, "image")
	}
	return SaveBase64(ctx, FileCategoryImage, encoded, "image", "")
}
