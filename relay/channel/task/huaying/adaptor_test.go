package huaying

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestAdaptor(channelType int, baseURL string) *TaskAdaptor {
	adaptor := &TaskAdaptor{}
	adaptor.Init(&relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:    channelType,
			ChannelBaseUrl: baseURL,
			ApiKey:         "test-key",
		},
	})
	return adaptor
}

func loadBillingModesForTest(t *testing.T, modes string) {
	t.Helper()
	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})
	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"billing_setting.billing_mode": modes,
	}))
}

func TestParseDurationSeconds(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "integer", raw: `5`, want: 5},
		{name: "decimal", raw: `5.5`, want: 5.5},
		{name: "numeric string", raw: `"5"`, want: 5},
		{name: "seconds string", raw: `"5s"`, want: 5},
		{name: "uppercase seconds", raw: `"8S"`, want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDurationSeconds(json.RawMessage(tt.raw))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	for _, raw := range []string{"", `null`, `0`, `-1`, `"abc"`, `{}`} {
		t.Run("invalid_"+raw, func(t *testing.T) {
			_, err := parseDurationSeconds(json.RawMessage(raw))
			require.Error(t, err)
		})
	}
}

func TestBuildRequestURL(t *testing.T) {
	for _, baseURL := range []string{
		"http://ai-studio.aixyzz.com",
		"http://ai-studio.aixyzz.com/",
		"http://ai-studio.aixyzz.com/v1/",
	} {
		huaying := newTestAdaptor(constant.ChannelTypeHuaying, baseURL)
		got, err := huaying.BuildRequestURL(nil)
		require.NoError(t, err)
		require.Equal(t, "http://ai-studio.aixyzz.com/v1/videos/generations", got)
	}

	customPrefix := newTestAdaptor(constant.ChannelTypeHuaying, "http://example.com/custom-api/")
	got, err := customPrefix.BuildRequestURL(nil)
	require.NoError(t, err)
	require.Equal(t, "http://example.com/custom-api/videos/generations", got)

	xinghe := newTestAdaptor(constant.ChannelTypeXingHe, "https://api.xheai.cc/")
	got, err = xinghe.BuildRequestURL(nil)
	require.NoError(t, err)
	require.Equal(t, "https://api.xheai.cc/api/generate-video", got)
}

func TestVideoGenerationRequestExamples(t *testing.T) {
	examples := []string{
		`{"model":"seedance-2.0","prompt":"city","duration":5,"resolution":"720p","aspectRatio":"16:9"}`,
		`{"model":"seedance-2.0","requestId":"order_1","prompt":"product","images":["https://example.com/1.png"],"videos":["https://example.com/1.mp4"],"audios":["https://example.com/1.mp3"],"duration":"8s","resolution":"720p","aspectRatio":"16:9"}`,
		`{"model":"seedance-2.0","prompt":"frames","firstFrame":"https://example.com/start.png","lastFrame":"https://example.com/end.png","duration":5,"resolution":"720p","aspectRatio":"16:9"}`,
	}

	for _, example := range examples {
		var request VideoGenerationRequest
		require.NoError(t, common.Unmarshal([]byte(example), &request))
		require.Equal(t, "seedance-2.0", request.Model)
		require.NotEmpty(t, request.Prompt)
		require.NotEmpty(t, request.Resolution)
		require.NotEmpty(t, request.AspectRatio)
		seconds, err := parseDurationSeconds(request.Duration)
		require.NoError(t, err)
		require.Positive(t, seconds)
	}
}

func TestValidateXingHeCompatibleRequest(t *testing.T) {
	loadBillingModesForTest(t, `{"seedance-2.0-fast-yo":"per_second"}`)

	body := `{"model":"seedance-2.0-fast-yo","prompt":"cat on moon","seconds":5,"size":"1280x720","protect_stripe":false}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	defer common.CleanupBodyStorage(c)

	adaptor := newTestAdaptor(constant.ChannelTypeXingHe, "https://api.xheai.cc")
	info := &relaycommon.RelayInfo{
		OriginModelName: "seedance-2.0-fast-yo",
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{},
	}
	require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))
	require.Equal(t, map[string]float64{"seconds": 5}, adaptor.EstimateBilling(c, info))

	reader, err := adaptor.BuildRequestBody(c, &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "mapped-model"}})
	require.NoError(t, err)
	upstreamBody, err := io.ReadAll(reader)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, common.Unmarshal(upstreamBody, &payload))
	require.Equal(t, "mapped-model", payload["model"])
	require.Equal(t, float64(5), payload["seconds"])
	require.Equal(t, "1280x720", payload["size"])
	require.Equal(t, false, payload["protect_stripe"])
	require.NotContains(t, payload, "duration")
	require.NotContains(t, payload, "resolution")
	require.NotContains(t, payload, "aspectRatio")
}

func TestEstimateBillingSkipsDurationForPerCallModel(t *testing.T) {
	loadBillingModesForTest(t, `{"seedance-2.0-fast-y1":"ratio"}`)

	body := `{"model":"seedance-2.0-fast-y1","prompt":"cat on moon","duration":5,"resolution":"720p","aspectRatio":"16:9"}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	defer common.CleanupBodyStorage(c)

	adaptor := newTestAdaptor(constant.ChannelTypeHuaying, "http://example.com/v1")
	info := &relaycommon.RelayInfo{
		OriginModelName: "seedance-2.0-fast-y1",
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{},
	}
	require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))
	require.Nil(t, adaptor.EstimateBilling(c, info))
}

func TestBuildHuayingBodyFromXingHeAliases(t *testing.T) {
	body := `{"model":"seedance-2.0","prompt":"cat on moon","seconds":5,"size":"1280x720","protect_stripe":false}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	defer common.CleanupBodyStorage(c)

	adaptor := newTestAdaptor(constant.ChannelTypeHuaying, "http://example.com/v1")
	info := &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
	require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))

	reader, err := adaptor.BuildRequestBody(c, &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "seedance-2.0"}})
	require.NoError(t, err)
	upstreamBody, err := io.ReadAll(reader)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, common.Unmarshal(upstreamBody, &payload))
	require.Equal(t, float64(5), payload["duration"])
	require.Equal(t, "720p", payload["resolution"])
	require.Equal(t, "16:9", payload["aspectRatio"])
	require.NotContains(t, payload, "seconds")
	require.NotContains(t, payload, "size")
	require.NotContains(t, payload, "protect_stripe")
}

func TestBuildXingHeBodyFromHuayingContract(t *testing.T) {
	body := `{"model":"seedance-2.0","prompt":"cat on moon","duration":"5s","resolution":"720p","aspectRatio":"9:16"}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	defer common.CleanupBodyStorage(c)

	adaptor := newTestAdaptor(constant.ChannelTypeXingHe, "https://api.xheai.cc")
	info := &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
	require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))

	reader, err := adaptor.BuildRequestBody(c, &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "mapped-model"}})
	require.NoError(t, err)
	upstreamBody, err := io.ReadAll(reader)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, common.Unmarshal(upstreamBody, &payload))
	require.Equal(t, float64(5), payload["seconds"])
	require.Equal(t, "720x1280", payload["size"])
}

func TestParseTaskResult(t *testing.T) {
	adaptor := newTestAdaptor(constant.ChannelTypeHuaying, "http://example.com/v1")
	tests := []struct {
		name       string
		body       string
		wantStatus model.TaskStatus
		wantURL    string
	}{
		{name: "processing", body: `{"id":"job_1","status":"processing","data":[]}`, wantStatus: model.TaskStatusInProgress},
		{name: "success array", body: `{"id":"job_1","status":"succeeded","data":[{"url":"https://example.com/video.mp4"}]}`, wantStatus: model.TaskStatusSuccess, wantURL: "https://example.com/video.mp4"},
		{name: "success direct string", body: `{"taskId":"job_2","status":"completed","data":"https://example.com/direct.mp4"}`, wantStatus: model.TaskStatusSuccess, wantURL: "https://example.com/direct.mp4"},
		{name: "canceled", body: `{"id":"job_3","status":"canceled","error":{"message":"canceled by upstream"}}`, wantStatus: model.TaskStatusFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adaptor.ParseTaskResult([]byte(tt.body))
			require.NoError(t, err)
			require.Equal(t, string(tt.wantStatus), result.Status)
			require.Equal(t, tt.wantURL, result.Url)
		})
	}
}

func TestConvertToOpenAIVideoUsesPublicTaskID(t *testing.T) {
	adaptor := newTestAdaptor(constant.ChannelTypeHuaying, "http://example.com/v1")
	task := &model.Task{
		TaskID: "task_public",
		Status: model.TaskStatusSuccess,
		Data:   json.RawMessage(`{"id":"job_secret","status":"succeeded","data":[{"url":"https://example.com/video.mp4"}]}`),
	}

	body, err := adaptor.ConvertToOpenAIVideo(task)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, common.Unmarshal(body, &payload))
	require.Equal(t, "task_public", payload["id"])
	require.Equal(t, "succeeded", payload["status"])
	require.NotContains(t, string(body), "job_secret")
}

func TestFetchTask(t *testing.T) {
	service.InitHttpClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/v1/videos/job_123", r.URL.EscapedPath())
		require.Equal(t, "Bearer polling-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"job_123","status":"processing","data":[]}`)
	}))
	defer server.Close()

	adaptor := newTestAdaptor(constant.ChannelTypeHuaying, server.URL+"/v1")
	resp, err := adaptor.FetchTask(server.URL, "polling-key", map[string]any{"task_id": "job_123"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
