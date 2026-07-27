package huaying

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const videoRequestContextKey = "huaying_video_request"

type TaskAdaptor struct {
	taskcommon.BaseBilling
	channelType int
	baseURL     string
	apiKey      string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.channelType = info.ChannelType
	if a.channelType == 0 && info.ChannelMeta != nil {
		a.channelType = info.ChannelMeta.ChannelType
	}
	a.baseURL = strings.TrimRight(info.ChannelBaseUrl, "/")
	if a.channelType == constant.ChannelTypeHuaying {
		a.baseURL = normalizeHuayingBaseURL(a.baseURL)
	}
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
		return service.TaskErrorWrapperLocal(fmt.Errorf("content-type must be application/json"), "invalid_request", http.StatusBadRequest)
	}

	var req VideoGenerationRequest
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
	}
	if err := normalizeVideoGenerationRequest(&req); err != nil {
		return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Model) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("model is required"), "missing_model", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("prompt is required"), "invalid_request", http.StatusBadRequest)
	}
	seconds, err := parseDurationSeconds(req.Duration)
	if err != nil || seconds <= 0 {
		if err == nil {
			err = fmt.Errorf("duration must be greater than 0")
		}
		return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Resolution) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("resolution is required"), "invalid_request", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.AspectRatio) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("aspectRatio is required"), "invalid_request", http.StatusBadRequest)
	}
	if req.LastFrame != nil && strings.TrimSpace(*req.LastFrame) != "" && (req.FirstFrame == nil || strings.TrimSpace(*req.FirstFrame) == "") {
		return service.TaskErrorWrapperLocal(fmt.Errorf("lastFrame requires firstFrame"), "invalid_request", http.StatusBadRequest)
	}

	info.Action = constant.TaskActionTextGenerate
	if len(req.Images) > 0 || len(req.Videos) > 0 || len(req.Audios) > 0 || (req.FirstFrame != nil && strings.TrimSpace(*req.FirstFrame) != "") {
		info.Action = constant.TaskActionGenerate
	}
	c.Set(videoRequestContextKey, req)
	return nil
}

func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	// 画影/星河同时支持按次和按秒计费。只有显式配置为按秒计费时，
	// duration 才是价格倍率；按次计费的 ModelPrice 不应再乘视频时长。
	if info == nil || billing_setting.GetBillingMode(info.OriginModelName) != billing_setting.BillingModePerSecond {
		return nil
	}
	value, ok := c.Get(videoRequestContextKey)
	if !ok {
		return nil
	}
	req, ok := value.(VideoGenerationRequest)
	if !ok {
		return nil
	}
	seconds, err := parseDurationSeconds(req.Duration)
	if err != nil || seconds <= 0 {
		return nil
	}
	return map[string]float64{"seconds": seconds}
}

func (a *TaskAdaptor) BuildRequestURL(_ *relaycommon.RelayInfo) (string, error) {
	switch a.channelType {
	case constant.ChannelTypeHuaying:
		return a.baseURL + huayingSubmitPath, nil
	case constant.ChannelTypeXingHe:
		return a.baseURL + xingHeSubmitPath, nil
	default:
		return "", fmt.Errorf("unsupported channel type: %d", a.channelType)
	}
}

func (a *TaskAdaptor) BuildRequestHeader(_ *gin.Context, req *http.Request, _ *relaycommon.RelayInfo) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	storage, err := common.GetBodyStorage(c)
	if err != nil {
		return nil, errors.Wrap(err, "get request body failed")
	}
	body, err := storage.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "read request body failed")
	}
	var payload map[string]any
	if err := common.Unmarshal(body, &payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal request body failed")
	}
	value, ok := c.Get(videoRequestContextKey)
	if !ok {
		return nil, errors.New("normalized video request is missing")
	}
	normalized, ok := value.(VideoGenerationRequest)
	if !ok {
		return nil, errors.New("normalized video request has an invalid type")
	}
	payload["model"] = info.UpstreamModelName
	switch a.channelType {
	case constant.ChannelTypeHuaying:
		duration, err := rawMessageToValue(normalized.Duration)
		if err != nil {
			return nil, errors.Wrap(err, "convert duration failed")
		}
		payload["duration"] = duration
		payload["resolution"] = normalized.Resolution
		payload["aspectRatio"] = normalized.AspectRatio
		delete(payload, "seconds")
		delete(payload, "size")
		delete(payload, "protect_stripe")
	case constant.ChannelTypeXingHe:
		seconds, err := parseDurationSeconds(normalized.Duration)
		if err != nil {
			return nil, errors.Wrap(err, "convert duration to seconds failed")
		}
		payload["seconds"] = compactNumber(seconds)
		size, err := huayingVideoSize(normalized.Resolution, normalized.AspectRatio)
		if err != nil {
			return nil, errors.Wrap(err, "convert resolution and aspectRatio to size failed")
		}
		payload["size"] = size
		delete(payload, "duration")
		delete(payload, "resolution")
		delete(payload, "aspectRatio")
	default:
		return nil, fmt.Errorf("unsupported channel type: %d", a.channelType)
	}
	body, err = common.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body failed")
	}
	return bytes.NewReader(body), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (string, []byte, *dto.TaskError) {
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}

	var payload map[string]any
	if err := common.Unmarshal(responseBody, &payload); err != nil {
		return "", nil, service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	upstreamTaskID := findString(payload, "id", "task_id", "taskId", "job_id", "jobId")
	if upstreamTaskID == "" {
		if nested, ok := payload["data"].(map[string]any); ok {
			upstreamTaskID = findString(nested, "id", "task_id", "taskId", "job_id", "jobId")
		}
	}
	if upstreamTaskID == "" && findVideoURL(payload) == "" {
		return "", nil, service.TaskErrorWrapper(fmt.Errorf("upstream task id is empty"), "invalid_response", http.StatusInternalServerError)
	}
	if upstreamTaskID == "" {
		// 星核部分实现会同步返回视频 URL，此时无需再访问上游任务 ID。
		upstreamTaskID = info.PublicTaskID
	}

	setPublicTaskID(payload, info.PublicTaskID)
	publicBody, err := common.Marshal(payload)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
	}
	c.Data(resp.StatusCode, "application/json", publicBody)
	return upstreamTaskID, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseURL, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("invalid task_id")
	}
	if a.channelType != constant.ChannelTypeHuaying {
		return nil, fmt.Errorf("channel %d does not expose an asynchronous task query endpoint", a.channelType)
	}
	uri := normalizeHuayingBaseURL(baseURL) + "/videos/" + url.PathEscape(taskID)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
		responseBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("temporary upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return resp, nil
}

// normalizeHuayingBaseURL accepts both the documented API base URL ending in
// /v1 and a bare origin URL. This is intentionally limited to an empty/root
// path so custom reverse-proxy prefixes remain untouched.
func normalizeHuayingBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return baseURL
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/v1"
		return strings.TrimRight(parsed.String(), "/")
	}
	return baseURL
}

func (a *TaskAdaptor) ParseInitialTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	return a.ParseTaskResult(respBody)
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	var payload map[string]any
	if err := common.Unmarshal(respBody, &payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}
	status := strings.ToLower(strings.TrimSpace(findString(payload, "status", "state")))
	result := &relaycommon.TaskInfo{
		TaskID: findString(payload, "id", "task_id", "taskId", "job_id", "jobId"),
		Url:    findVideoURL(payload),
	}
	switch status {
	case "queued", "pending", "submitted", "created", "waiting":
		result.Status = string(model.TaskStatusQueued)
		result.Progress = taskcommon.ProgressQueued
	case "processing", "running", "in_progress", "generating":
		result.Status = string(model.TaskStatusInProgress)
		result.Progress = taskcommon.ProgressInProgress
	case "succeeded", "success", "completed", "done", "finished":
		result.Status = string(model.TaskStatusSuccess)
		result.Progress = taskcommon.ProgressComplete
	case "failed", "failure", "canceled", "cancelled", "error":
		result.Status = string(model.TaskStatusFailure)
		result.Progress = taskcommon.ProgressComplete
		result.Reason = findErrorMessage(payload)
	default:
		if result.Url != "" {
			result.Status = string(model.TaskStatusSuccess)
			result.Progress = taskcommon.ProgressComplete
		} else if reason := findErrorMessage(payload); reason != "" {
			result.Status = string(model.TaskStatusFailure)
			result.Progress = taskcommon.ProgressComplete
			result.Reason = reason
		} else {
			result.Status = string(model.TaskStatusInProgress)
			result.Progress = taskcommon.ProgressInProgress
		}
	}
	if result.Status == string(model.TaskStatusFailure) && result.Reason == "" {
		result.Reason = "task failed"
	}
	return result, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	payload := map[string]any{}
	if len(task.Data) > 0 {
		_ = common.Unmarshal(task.Data, &payload)
	}
	setPublicTaskID(payload, task.TaskID)
	switch task.Status {
	case model.TaskStatusSuccess:
		payload["status"] = "succeeded"
	case model.TaskStatusFailure:
		payload["status"] = "failed"
		if _, exists := payload["error"]; !exists || payload["error"] == nil {
			payload["error"] = map[string]any{
				"message": task.FailReason,
				"type":    "task_error",
				"code":    "task_failed",
			}
		}
	case model.TaskStatusQueued, model.TaskStatusSubmitted:
		payload["status"] = "processing"
	case model.TaskStatusInProgress, model.TaskStatusNotStart:
		if strings.TrimSpace(findString(payload, "status")) == "" {
			payload["status"] = "processing"
		}
	}
	if task.Status == model.TaskStatusSuccess && task.GetResultURL() != "" && findVideoURL(payload) == "" {
		payload["data"] = []any{map[string]any{"url": task.GetResultURL()}}
	}
	return common.Marshal(payload)
}

func (a *TaskAdaptor) GetModelList() []string { return nil }

func (a *TaskAdaptor) GetChannelName() string {
	if a.channelType == constant.ChannelTypeXingHe {
		return "xinghe-video"
	}
	return ChannelName
}

func parseDurationSeconds(raw json.RawMessage) (float64, error) {
	if rawMessageMissing(raw) {
		return 0, fmt.Errorf("duration is required")
	}
	var number float64
	if err := common.Unmarshal(raw, &number); err == nil {
		if number <= 0 {
			return 0, fmt.Errorf("duration must be greater than 0")
		}
		return number, nil
	}
	var text string
	if err := common.Unmarshal(raw, &text); err != nil {
		return 0, fmt.Errorf("duration must be a number or a string ending in s")
	}
	text = strings.TrimSpace(strings.TrimSuffix(strings.ToLower(text), "s"))
	number, err := strconv.ParseFloat(text, 64)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("invalid duration")
	}
	return number, nil
}

func normalizeVideoGenerationRequest(req *VideoGenerationRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	if rawMessageMissing(req.Duration) && !rawMessageMissing(req.Seconds) {
		req.Duration = append(json.RawMessage(nil), req.Seconds...)
	}

	if strings.TrimSpace(req.Resolution) == "" || strings.TrimSpace(req.AspectRatio) == "" {
		size := strings.TrimSpace(req.Size)
		if size != "" {
			resolution, aspectRatio, err := huayingParametersFromSize(size)
			if err != nil {
				return err
			}
			if strings.TrimSpace(req.Resolution) == "" {
				req.Resolution = resolution
			}
			if strings.TrimSpace(req.AspectRatio) == "" {
				req.AspectRatio = aspectRatio
			}
		}
	}
	return nil
}

func rawMessageMissing(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed == "" || trimmed == "null"
}

func rawMessageToValue(raw json.RawMessage) (any, error) {
	if rawMessageMissing(raw) {
		return nil, fmt.Errorf("duration is required")
	}
	var value any
	if err := common.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func compactNumber(value float64) any {
	if value == math.Trunc(value) && value <= math.MaxInt64 {
		return int64(value)
	}
	return value
}

func huayingParametersFromSize(size string) (string, string, error) {
	width, height, err := parseVideoDimensions(size)
	if err != nil {
		return "", "", fmt.Errorf("invalid size: %w", err)
	}
	shortEdge := width
	if height < shortEdge {
		shortEdge = height
	}
	divisor := greatestCommonDivisor(width, height)
	return fmt.Sprintf("%dp", shortEdge), fmt.Sprintf("%d:%d", width/divisor, height/divisor), nil
}

func huayingVideoSize(resolution, aspectRatio string) (string, error) {
	resolution = strings.ToLower(strings.TrimSpace(resolution))
	if strings.Contains(resolution, "x") {
		width, height, err := parseVideoDimensions(resolution)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%dx%d", width, height), nil
	}

	shortEdgeText := strings.TrimSuffix(resolution, "p")
	shortEdge, err := strconv.Atoi(shortEdgeText)
	if err != nil || shortEdge <= 0 {
		return "", fmt.Errorf("invalid resolution %q", resolution)
	}
	ratioParts := strings.Split(strings.TrimSpace(aspectRatio), ":")
	if len(ratioParts) != 2 {
		return "", fmt.Errorf("invalid aspectRatio %q", aspectRatio)
	}
	ratioWidth, err := strconv.Atoi(strings.TrimSpace(ratioParts[0]))
	if err != nil || ratioWidth <= 0 {
		return "", fmt.Errorf("invalid aspectRatio %q", aspectRatio)
	}
	ratioHeight, err := strconv.Atoi(strings.TrimSpace(ratioParts[1]))
	if err != nil || ratioHeight <= 0 {
		return "", fmt.Errorf("invalid aspectRatio %q", aspectRatio)
	}

	width, height := shortEdge, shortEdge
	if ratioWidth >= ratioHeight {
		width = int(math.Round(float64(shortEdge) * float64(ratioWidth) / float64(ratioHeight)))
	} else {
		height = int(math.Round(float64(shortEdge) * float64(ratioHeight) / float64(ratioWidth)))
	}
	return fmt.Sprintf("%dx%d", width, height), nil
}

func parseVideoDimensions(size string) (int, int, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected WIDTHxHEIGHT")
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || width <= 0 {
		return 0, 0, fmt.Errorf("invalid width")
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || height <= 0 {
		return 0, 0, fmt.Errorf("invalid height")
	}
	return width, height, nil
}

func greatestCommonDivisor(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a <= 0 {
		return 1
	}
	return a
}

func findString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok && value != nil {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return text
			}
		}
	}
	return ""
}

func findVideoURL(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"video_url", "videoUrl", "url", "output_url", "outputUrl"} {
			if text, ok := typed[key].(string); ok && strings.TrimSpace(text) != "" {
				return text
			}
		}
		for _, key := range []string{"data", "output", "result", "video"} {
			if nested, exists := typed[key]; exists {
				if url := findVideoURL(nested); url != "" {
					return url
				}
			}
		}
	case []any:
		for _, item := range typed {
			if url := findVideoURL(item); url != "" {
				return url
			}
		}
	case string:
		text := strings.TrimSpace(typed)
		if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
			return text
		}
	}
	return ""
}

func findErrorMessage(payload map[string]any) string {
	if text := findString(payload, "message", "error_message", "errorMessage", "reason"); text != "" {
		return text
	}
	if errObj, ok := payload["error"].(map[string]any); ok {
		return findString(errObj, "message", "detail", "reason", "code")
	}
	if text, ok := payload["error"].(string); ok {
		return text
	}
	return ""
}

func setPublicTaskID(payload map[string]any, publicTaskID string) {
	if payload == nil {
		return
	}
	if _, exists := payload["task_id"]; exists {
		payload["task_id"] = publicTaskID
	}
	if _, exists := payload["taskId"]; exists {
		payload["taskId"] = publicTaskID
	}
	payload["id"] = publicTaskID

	for _, key := range []string{"data", "result"} {
		if nested, ok := payload[key].(map[string]any); ok {
			if _, exists := nested["id"]; exists {
				nested["id"] = publicTaskID
			}
			if _, exists := nested["task_id"]; exists {
				nested["task_id"] = publicTaskID
			}
			if _, exists := nested["taskId"]; exists {
				nested["taskId"] = publicTaskID
			}
		}
	}
}
