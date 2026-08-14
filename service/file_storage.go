package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/google/uuid"
)

type FileCategory string

const (
	FileCategoryUpload FileCategory = "upload"
	FileCategoryImage  FileCategory = "gc/img"
	FileCategoryVideo  FileCategory = "gc/video"
)

type FileStorageConfig struct {
	BucketName    string
	OSSEnabled    bool
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Region        string
	ServerURL     string
	LocalEnabled  bool
	LocalBasePath string
}

func LoadFileStorageConfig() FileStorageConfig {
	return FileStorageConfig{
		BucketName:    strings.TrimSpace(common.GetEnvOrDefaultString("FILE_BUCKET_NAME", "")),
		OSSEnabled:    common.GetEnvOrDefaultBool("FILE_OSS_ENABLE", false),
		Endpoint:      strings.TrimRight(strings.TrimSpace(common.GetEnvOrDefaultString("FILE_OSS_ENDPOINT", "")), "/"),
		AccessKey:     strings.TrimSpace(common.GetEnvOrDefaultString("FILE_OSS_ACCESS_KEY", "")),
		SecretKey:     strings.TrimSpace(common.GetEnvOrDefaultString("FILE_OSS_SECRET_KEY", "")),
		Region:        strings.TrimSpace(common.GetEnvOrDefaultString("FILE_OSS_REGION", "us-east-1")),
		ServerURL:     strings.TrimRight(strings.TrimSpace(common.GetEnvOrDefaultString("FILE_OSS_SERVER_URL", "")), "/"),
		LocalEnabled:  common.GetEnvOrDefaultBool("FILE_LOCAL_ENABLE", false),
		LocalBasePath: strings.TrimSpace(common.GetEnvOrDefaultString("FILE_LOCAL_BASE_PATH", ".")),
	}
}

func validateStorageURL(name, value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", name, err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid HTTP(S) URL", name)
	}
	return nil
}

func (c FileStorageConfig) Validate() error {
	if c.OSSEnabled {
		if c.BucketName == "" || c.Endpoint == "" || c.AccessKey == "" || c.SecretKey == "" || c.ServerURL == "" {
			return fmt.Errorf("OSS storage is enabled but required configuration is incomplete")
		}
		if c.Region == "" {
			return fmt.Errorf("OSS storage region is empty")
		}
		if err := validateStorageURL("FILE_OSS_ENDPOINT", c.Endpoint); err != nil {
			return err
		}
		if err := validateStorageURL("FILE_OSS_SERVER_URL", c.ServerURL); err != nil {
			return err
		}
		return nil
	}
	if c.LocalEnabled {
		if c.LocalBasePath == "" {
			return fmt.Errorf("local storage is enabled but FILE_LOCAL_BASE_PATH is empty")
		}
		return nil
	}
	return fmt.Errorf("file storage is not enabled")
}

func (c FileStorageConfig) UsesLocalStorage() bool {
	return !c.OSSEnabled && c.LocalEnabled
}

func Save(ctx context.Context, category FileCategory, reader io.Reader, originalName, contentType string) (string, error) {
	cfg := LoadFileStorageConfig()
	if err := cfg.Validate(); err != nil {
		return "", err
	}
	key := buildStorageKey(category, time.Now(), originalName, contentType)
	if cfg.OSSEnabled {
		if err := putOSS(ctx, cfg, key, reader, contentType); err != nil {
			return "", err
		}
		return publicURL(cfg.ServerURL, path.Join(cfg.BucketName, key)), nil
	}
	if err := putLocal(cfg.LocalBasePath, key, reader); err != nil {
		return "", err
	}
	return publicURL(system_setting.ServerAddress, key), nil
}

func SaveBytes(ctx context.Context, category FileCategory, data []byte, originalName, contentType string) (string, error) {
	return Save(ctx, category, bytes.NewReader(data), originalName, contentType)
}

func SaveDataURL(ctx context.Context, category FileCategory, dataURL, originalName string) (string, error) {
	mediaType, data, err := decodeDataURL(dataURL)
	if err != nil {
		return "", err
	}
	return SaveBytes(ctx, category, data, originalName, mediaType)
}

func SaveBase64(ctx context.Context, category FileCategory, encoded, originalName, contentType string) (string, error) {
	data, err := decodeBase64(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64 file failed: %w", err)
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return SaveBytes(ctx, category, data, originalName, contentType)
}

func SaveRemote(ctx context.Context, category FileCategory, rawURL string) (string, error) {
	return SaveRemoteWithHeaders(ctx, category, rawURL, nil)
}

func SaveRemoteWithHeaders(ctx context.Context, category FileCategory, rawURL string, headers http.Header) (string, error) {
	return SaveRemoteWithOptions(ctx, category, rawURL, headers, "")
}

// MediaDownloadError describes a failed origin download without exposing the
// complete source URL. Retryable statuses are treated as temporary by video
// task polling, because a provider can publish the task before its CDN object.
type MediaDownloadError struct {
	StatusCode    int
	OriginHost    string
	PathExtension string
	ViaProxy      bool
	Attempts      int
	ContentType   string
	ContentLength int64
}

func (e *MediaDownloadError) Error() string {
	if e == nil {
		return "generated media download failed"
	}
	return fmt.Sprintf("download generated file failed: HTTP %d, origin_host=%s, path_ext=%s, proxy=%t, content_type=%q, content_length=%d, attempts=%d",
		e.StatusCode, e.OriginHost, e.PathExtension, e.ViaProxy, e.ContentType, e.ContentLength, e.Attempts)
}

func (e *MediaDownloadError) Temporary() bool {
	return e != nil && isRetryableMediaDownloadStatus(e.StatusCode)
}

func SaveRemoteWithOptions(ctx context.Context, category FileCategory, rawURL string, headers http.Header, proxy string) (string, error) {
	resp, err := downloadRemoteWithRetry(ctx, rawURL, headers, proxy)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", formatMediaDownloadError(rawURL, resp, false, 0)
	}
	maxBytes := int64(constant.MaxFileDownloadMB) << 20
	if maxBytes <= 0 {
		maxBytes = 64 << 20
	}
	if resp.ContentLength > maxBytes {
		return "", fmt.Errorf("generated file size %d exceeds limit %d", resp.ContentLength, maxBytes)
	}
	limited := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("generated file exceeds %d byte limit", maxBytes)
	}
	contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}
	originalName := ""
	if parsed, parseErr := url.Parse(rawURL); parseErr == nil {
		originalName = path.Base(parsed.Path)
	}
	return SaveBytes(ctx, category, data, originalName, contentType)
}

// downloadRemoteWithRetry tolerates short-lived CDN propagation failures. Some
// video providers report a completed task before the returned object is visible
// on every CDN node. When a channel-specific proxy is configured, the first
// retry also bypasses that proxy because generated URLs may be hosted elsewhere.
func downloadRemoteWithRetry(ctx context.Context, rawURL string, headers http.Header, proxy string) (*http.Response, error) {
	// Keep both worker/proxy access and local direct access. A CDN object can
	// become visible a few seconds after the provider reports task success.
	const maxAttempts = 6
	currentProxy := strings.TrimSpace(proxy)
	var lastErr error
	lastStatus := 0
	lastViaProxy := currentProxy != "" || systemProxyConfigured()

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var resp *http.Response
		var err error
		viaProxy := currentProxy != "" || (systemProxyConfigured() && currentProxy == "")
		if attempt%2 == 0 {
			// Even attempts use the original path, including a configured worker.
			resp, err = downloadRemote(ctx, rawURL, headers, currentProxy)
		} else {
			// Odd attempts bypass the worker, channel proxy, and process proxy.
			viaProxy = false
			resp, err = downloadRemoteWithoutProxy(ctx, rawURL, headers)
		}
		if err == nil && resp != nil {
			lastStatus = resp.StatusCode
			lastViaProxy = viaProxy
			if !isRetryableMediaDownloadStatus(resp.StatusCode) {
				return resp, nil
			}
			if attempt == maxAttempts-1 {
				err := formatMediaDownloadError(rawURL, resp, viaProxy, attempt+1)
				_ = resp.Body.Close()
				return nil, err
			}
			_ = resp.Body.Close()
		} else {
			lastErr = err
		}

		if attempt == 0 {
			currentProxy = ""
		}
		// Back off for 0.5, 1, 2, 4, and 8 seconds to allow CDN propagation.
		delay := time.Duration(1<<attempt) * 500 * time.Millisecond
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("download generated file failed: %w, %s", lastErr, mediaDownloadSourceInfo(rawURL, lastViaProxy))
	}
	return nil, &MediaDownloadError{
		StatusCode:    lastStatus,
		OriginHost:    mediaDownloadHost(rawURL),
		PathExtension: mediaDownloadExtension(rawURL),
		ViaProxy:      lastViaProxy,
		Attempts:      maxAttempts,
		ContentLength: -1,
	}
}

func systemProxyConfigured() bool {
	return strings.TrimSpace(os.Getenv("HTTP_PROXY")) != "" || strings.TrimSpace(os.Getenv("HTTPS_PROXY")) != "" ||
		strings.TrimSpace(os.Getenv("http_proxy")) != "" || strings.TrimSpace(os.Getenv("https_proxy")) != ""
}

func mediaDownloadHost(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil || parsed.Host == "" {
		return "unknown"
	}
	return parsed.Host
}

func mediaDownloadExtension(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return "none"
	}
	if ext := path.Ext(parsed.Path); ext != "" {
		return ext
	}
	return "none"
}

func mediaDownloadSourceInfo(rawURL string, viaProxy bool) string {
	return fmt.Sprintf("origin_host=%s, path_ext=%s, proxy=%t", mediaDownloadHost(rawURL), mediaDownloadExtension(rawURL), viaProxy)
}

func formatMediaDownloadError(rawURL string, resp *http.Response, viaProxy bool, attempt int) error {
	status := 0
	contentType := ""
	contentLength := int64(-1)
	if resp != nil {
		status = resp.StatusCode
		contentType = strings.TrimSpace(resp.Header.Get("Content-Type"))
		contentLength = resp.ContentLength
	}
	originHost := mediaDownloadHost(rawURL)
	pathExtension := mediaDownloadExtension(rawURL)
	return &MediaDownloadError{
		StatusCode:    status,
		OriginHost:    originHost,
		PathExtension: pathExtension,
		ViaProxy:      viaProxy,
		Attempts:      attempt,
		ContentType:   contentType,
		ContentLength: contentLength,
	}
}

func downloadRemoteWithoutProxy(ctx context.Context, rawURL string, headers http.Header) (*http.Response, error) {
	fetchSetting := system_setting.GetFetchSetting()
	if err := common.ValidateURLWithFetchSetting(rawURL, fetchSetting.EnableSSRFProtection, fetchSetting.AllowPrivateIp, fetchSetting.DomainFilterMode, fetchSetting.IpFilterMode, fetchSetting.DomainList, fetchSetting.IpList, fetchSetting.AllowedPorts, fetchSetting.ApplyIPFilterForDomain); err != nil {
		return nil, fmt.Errorf("request reject: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		req.Header = headers.Clone()
	}

	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok || baseTransport == nil {
		baseTransport = &http.Transport{}
	}
	transport := baseTransport.Clone()
	transport.Proxy = nil
	if common.TLSInsecureSkipVerify {
		transport.TLSClientConfig = common.InsecureTLSConfig
	}
	client := &http.Client{Transport: transport, CheckRedirect: checkRedirect}
	if common.RelayTimeout > 0 {
		client.Timeout = time.Duration(common.RelayTimeout) * time.Second
	}
	return client.Do(req)
}

func isRetryableMediaDownloadStatus(status int) bool {
	return status == http.StatusNotFound || status == http.StatusConflict ||
		status == http.StatusTooEarly || status == http.StatusTooManyRequests ||
		status >= http.StatusInternalServerError
}
func downloadRemote(ctx context.Context, rawURL string, headers http.Header, proxy string) (*http.Response, error) {
	if strings.TrimSpace(proxy) == "" && system_setting.EnableWorker() {
		common.SysLog(fmt.Sprintf("downloading file from worker: %s, reason: persist generated media", common.MaskSensitiveInfo(rawURL)))
		workerHeaders := make(map[string]string, len(headers))
		for key, values := range headers {
			if len(values) > 0 {
				workerHeaders[key] = values[0]
			}
		}
		return DoWorkerRequest(&WorkerRequest{
			URL:     rawURL,
			Key:     system_setting.WorkerValidKey,
			Method:  http.MethodGet,
			Headers: workerHeaders,
		})
	}
	if len(headers) == 0 && strings.TrimSpace(proxy) == "" {
		return DoDownloadRequest(rawURL, "persist generated media")
	}
	fetchSetting := system_setting.GetFetchSetting()
	if err := common.ValidateURLWithFetchSetting(rawURL, fetchSetting.EnableSSRFProtection, fetchSetting.AllowPrivateIp, fetchSetting.DomainFilterMode, fetchSetting.IpFilterMode, fetchSetting.DomainList, fetchSetting.IpList, fetchSetting.AllowedPorts, fetchSetting.ApplyIPFilterForDomain); err != nil {
		return nil, fmt.Errorf("request reject: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers.Clone()
	client, err := GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	return client.Do(req)
}

func buildStorageKey(category FileCategory, now time.Time, originalName, contentType string) string {
	ext := resolveExtension(originalName, contentType)
	return path.Join(string(category), now.Format("200601"), uuid.NewString()+ext)
}

func putLocal(basePath, key string, reader io.Reader) error {
	baseAbs, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(filepath.Join(baseAbs, filepath.FromSlash(key)))
	if err != nil {
		return err
	}
	prefix := baseAbs + string(os.PathSeparator)
	if targetAbs != baseAbs && !strings.HasPrefix(targetAbs, prefix) {
		return fmt.Errorf("invalid file path")
	}
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(targetAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

func putOSS(ctx context.Context, cfg FileStorageConfig, key string, reader io.Reader, contentType string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	objectURL := strings.TrimRight(cfg.Endpoint, "/") + "/" + url.PathEscape(cfg.BucketName) + "/" + escapeObjectKey(key)
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	hash := sha256.Sum256(data)
	payloadHash := hex.EncodeToString(hash[:])
	request.Header.Set("X-Amz-Content-Sha256", payloadHash)
	creds, err := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "").Retrieve(ctx)
	if err != nil {
		return err
	}
	if err := v4.NewSigner().SignHTTP(ctx, creds, request, payloadHash, "s3", cfg.Region, time.Now()); err != nil {
		return err
	}
	client := GetHttpClient()
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("upload to OSS failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func escapeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func decodeDataURL(value string) (string, []byte, error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:") {
		return "", nil, fmt.Errorf("invalid data URL")
	}
	header := strings.TrimPrefix(parts[0], "data:")
	if !strings.Contains(header, ";base64") {
		return "", nil, fmt.Errorf("only base64 data URL is supported")
	}
	mediaType := strings.TrimSuffix(header, ";base64")
	data, err := decodeBase64(parts[1])
	if err != nil {
		return "", nil, err
	}
	if mediaType == "" {
		mediaType = http.DetectContentType(data)
	}
	return mediaType, data, nil
}

func decodeBase64(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if data, err := base64.StdEncoding.DecodeString(value); err == nil {
		return data, nil
	}
	return base64.RawStdEncoding.DecodeString(value)
}

func resolveExtension(originalName, contentType string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if validExtension(ext) {
		return ext
	}
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	if extensions, err := mime.ExtensionsByType(contentType); err == nil && len(extensions) > 0 && validExtension(extensions[0]) {
		return strings.ToLower(extensions[0])
	}
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	}
	return ""
}

func validExtension(ext string) bool {
	if ext == "" || len(ext) > 16 || !strings.HasPrefix(ext, ".") {
		return false
	}
	for _, r := range ext[1:] {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func publicURL(baseURL, key string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(key, "/")
}
