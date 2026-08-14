package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedVideoDownloadHeaders(t *testing.T) {
	baseURL := "https://api.example.com/v1"
	key := "dedicated-video-key"
	ch := &model.Channel{Type: constant.ChannelTypeHongNiao, BaseURL: &baseURL, Key: key}
	task := &model.Task{PrivateData: model.TaskPrivateData{Key: key}}

	t.Run("same origin includes authorization", func(t *testing.T) {
		headers := generatedVideoDownloadHeaders(ch, task, "https://API.EXAMPLE.COM/v1/files/result.mp4")
		assert.Equal(t, "Bearer "+key, headers.Get("Authorization"))
	})

	t.Run("cdn origin does not include authorization", func(t *testing.T) {
		headers := generatedVideoDownloadHeaders(ch, task, "https://file.tripcdn.com/files/result.mp4")
		assert.Empty(t, headers.Get("Authorization"))
		assert.NotEmpty(t, headers.Get("User-Agent"))
		assert.Contains(t, headers.Get("Accept"), "video/mp4")
	})

	t.Run("different port does not include authorization", func(t *testing.T) {
		headers := generatedVideoDownloadHeaders(ch, task, "https://api.example.com:8443/v1/files/result.mp4")
		assert.Empty(t, headers.Get("Authorization"))
	})

	t.Run("non dedicated channel does not include authorization", func(t *testing.T) {
		other := &model.Channel{Type: constant.ChannelTypeOpenAI, BaseURL: &baseURL, Key: key}
		headers := generatedVideoDownloadHeaders(other, task, "https://api.example.com/v1/files/result.mp4")
		assert.Empty(t, headers)
	})
}
func TestPersistGeneratedVideoDataURL(t *testing.T) {
	basePath := configureLocalFileStorage(t)
	fileURL, err := PersistGeneratedVideo(
		context.Background(),
		&model.Channel{},
		&model.Task{},
		&relaycommon.TaskInfo{Url: "data:video/mp4;base64,AAAAIGZ0eXA="},
	)
	require.NoError(t, err)
	assert.Contains(t, fileURL, "/gc/video/"+time.Now().Format("200601")+"/")
	_, err = os.Stat(publicURLToLocalPath(t, basePath, fileURL))
	require.NoError(t, err)
}

func TestOverrideOpenAIVideoURL(t *testing.T) {
	body := []byte(`{
		"id": "task_1",
		"object": "video",
		"model": "video-model",
		"status": "completed",
		"progress": 100,
		"created_at": 1,
		"metadata": {"url": "https://upstream.example/video.mp4", "keep": true}
	}`)

	result, err := OverrideOpenAIVideoURL(body, "https://files.example.com/gc/video/202608/result.mp4")
	require.NoError(t, err)

	var video dto.OpenAIVideo
	require.NoError(t, common.Unmarshal(result, &video))
	assert.Equal(t, "https://files.example.com/gc/video/202608/result.mp4", video.Metadata["url"])
	assert.Equal(t, true, video.Metadata["keep"])
}

func TestOverrideGeneratedVideoResponseURL(t *testing.T) {
	body := []byte(`{
		"id": "task_1",
		"status": "completed",
		"data": {
			"video_url": "https://upstream.example/video.mp4",
			"keep": 123
		}
	}`)

	result, err := OverrideGeneratedVideoResponseURL(body, "https://files.example.com/gc/video/202608/result.mp4")
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id": "task_1",
		"status": "completed",
		"data": {
			"video_url": "https://files.example.com/gc/video/202608/result.mp4",
			"keep": 123
		}
	}`, string(result))
}

func TestPersistGeneratedVideoCoverDataURL(t *testing.T) {
	basePath := configureLocalFileStorage(t)
	fileURL, err := PersistGeneratedVideoCover(
		context.Background(),
		&model.Channel{},
		&model.Task{},
		&relaycommon.TaskInfo{CoverUrl: "data:image/png;base64,iVBORw0KGgo="},
	)
	require.NoError(t, err)
	assert.Contains(t, fileURL, "/gc/img/"+time.Now().Format("200601")+"/")
	assert.FileExists(t, publicURLToLocalPath(t, basePath, fileURL))
}

func TestOverrideGeneratedVideoResponseMediaURLs(t *testing.T) {
	body := []byte(`{
		"id": "job_1",
		"status": "succeeded",
		"data": [{
			"url": "https://upstream.example/video.mp4",
			"cover_url": "https://upstream.example/poster.jpg",
			"keep": true
		}]
	}`)

	result, err := OverrideGeneratedVideoResponseMediaURLs(
		body,
		"https://files.example.com/gc/video/202608/result.mp4",
		"https://files.example.com/gc/img/202608/poster.jpg",
	)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id": "job_1",
		"status": "succeeded",
		"data": [{
			"url": "https://files.example.com/gc/video/202608/result.mp4",
			"cover_url": "https://files.example.com/gc/img/202608/poster.jpg",
			"keep": true
		}]
	}`, string(result))

	arrayBody := []byte(`{"result":{"output":{"outputUrls":["https://upstream.example/video.mp4"]},"outputs":["https://upstream.example/video.mp4"],"videoUrls":["https://upstream.example/video.mp4"],"resultUrls":["https://upstream.example/video.mp4"]}}`)
	arrayResult, err := OverrideGeneratedVideoResponseURL(arrayBody, "https://files.example.com/gc/video/202608/result.mp4")
	require.NoError(t, err)
	assert.JSONEq(t, `{"result":{"output":{"outputUrls":["https://files.example.com/gc/video/202608/result.mp4"]},"outputs":["https://files.example.com/gc/video/202608/result.mp4"],"videoUrls":["https://files.example.com/gc/video/202608/result.mp4"],"resultUrls":["https://files.example.com/gc/video/202608/result.mp4"]}}`, string(arrayResult))
}

func TestOverrideOpenAIVideoMediaURLs(t *testing.T) {
	body := []byte(`{
		"id": "task_1",
		"object": "video",
		"model": "video-model",
		"status": "completed",
		"progress": 100,
		"created_at": 1,
		"metadata": {"keep": true}
	}`)

	result, err := OverrideOpenAIVideoMediaURLs(
		body,
		"https://files.example.com/gc/video/202608/result.mp4",
		"https://files.example.com/gc/img/202608/poster.jpg",
	)
	require.NoError(t, err)

	var video dto.OpenAIVideo
	require.NoError(t, common.Unmarshal(result, &video))
	assert.Equal(t, "https://files.example.com/gc/video/202608/result.mp4", video.Metadata["url"])
	assert.Equal(t, "https://files.example.com/gc/img/202608/poster.jpg", video.Metadata["cover_url"])
	assert.Equal(t, true, video.Metadata["keep"])
}

func TestPersistGeneratedVideoMediaBestEffortKeepsOriginalOnFailure(t *testing.T) {
	configureLocalFileStorage(t)
	task := &model.Task{TaskID: "task_keep_original"}
	result := &relaycommon.TaskInfo{
		Url:      "data:video/mp4;base64,%%%",
		CoverUrl: "https://upstream.example/poster.jpg",
	}

	videoURL, coverURL, persisted, err := PersistGeneratedVideoMediaBestEffort(context.Background(), &model.Channel{}, task, result)
	require.Error(t, err)
	assert.False(t, persisted)
	assert.Equal(t, result.Url, videoURL)
	assert.Equal(t, result.CoverUrl, coverURL)
}

func TestOriginalGeneratedVideoMediaURLsUsesRemoteURLFallback(t *testing.T) {
	task := &model.Task{TaskID: "task_remote_fallback"}
	result := &relaycommon.TaskInfo{
		RemoteUrl: "https://upstream.example/video-content",
		CoverUrl:  "https://upstream.example/poster.jpg",
	}

	videoURL, coverURL := originalGeneratedVideoMediaURLs(task, result)
	assert.Equal(t, result.RemoteUrl, videoURL)
	assert.Equal(t, result.CoverUrl, coverURL)
}
