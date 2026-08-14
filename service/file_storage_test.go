package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configureLocalFileStorage(t *testing.T) string {
	t.Helper()
	basePath := t.TempDir()
	t.Setenv("FILE_OSS_ENABLE", "false")
	t.Setenv("FILE_LOCAL_ENABLE", "true")
	t.Setenv("FILE_LOCAL_BASE_PATH", basePath)
	oldServerAddress := system_setting.ServerAddress
	system_setting.ServerAddress = "https://files.example.com"
	t.Cleanup(func() {
		system_setting.ServerAddress = oldServerAddress
	})
	return basePath
}

func publicURLToLocalPath(t *testing.T, basePath, fileURL string) string {
	t.Helper()
	const serverURL = "https://files.example.com/"
	require.True(t, strings.HasPrefix(fileURL, serverURL), fileURL)
	key := strings.TrimPrefix(fileURL, serverURL)
	return filepath.Join(basePath, filepath.FromSlash(key))
}

func TestBuildStorageKey(t *testing.T) {
	key := buildStorageKey(FileCategoryImage, time.Date(2026, time.August, 12, 0, 0, 0, 0, time.UTC), "result.PNG", "image/png")
	assert.True(t, strings.HasPrefix(key, "gc/img/202608/"), key)
	assert.True(t, strings.HasSuffix(key, ".png"), key)
}

func TestSaveLocalFile(t *testing.T) {
	basePath := configureLocalFileStorage(t)
	fileURL, err := SaveBytes(context.Background(), FileCategoryUpload, []byte("hello"), "hello.txt", "text/plain")
	require.NoError(t, err)
	assert.Contains(t, fileURL, "/upload/"+time.Now().Format("200601")+"/")

	storedPath := publicURLToLocalPath(t, basePath, fileURL)
	data, err := os.ReadFile(storedPath)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestSaveRemoteRetriesTemporaryNotFound(t *testing.T) {
	basePath := configureLocalFileStorage(t)
	InitHttpClient()
	fetchSetting := system_setting.GetFetchSetting()
	oldSSRFProtection := fetchSetting.EnableSSRFProtection
	fetchSetting.EnableSSRFProtection = false
	t.Cleanup(func() { fetchSetting.EnableSSRFProtection = oldSSRFProtection })
	InitHttpClient()
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("video-bytes"))
	}))
	defer server.Close()

	fileURL, err := SaveRemoteWithHeaders(context.Background(), FileCategoryVideo, server.URL+"/result.mp4", http.Header{
		"Accept": []string{"video/*"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.FileExists(t, publicURLToLocalPath(t, basePath, fileURL))
}
func TestSaveDataURL(t *testing.T) {
	basePath := configureLocalFileStorage(t)
	fileURL, err := SaveDataURL(context.Background(), FileCategoryImage, "data:image/png;base64,iVBORw0KGgo=", "image")
	require.NoError(t, err)
	assert.Contains(t, fileURL, "/gc/img/"+time.Now().Format("200601")+"/")
	assert.FileExists(t, publicURLToLocalPath(t, basePath, fileURL))
}

func TestFileStorageConfigRequiresExplicitURLScheme(t *testing.T) {
	cfg := FileStorageConfig{
		BucketName: "test-bucket",
		OSSEnabled: true,
		Endpoint:   "ip:port",
		AccessKey:  "access-key",
		SecretKey:  "secret-key",
		Region:     "us-east-1",
		ServerURL:  "https://files.example.com",
	}

	require.ErrorContains(t, cfg.Validate(), "FILE_OSS_ENDPOINT")
}

func TestFileStorageConfigRejectsInvalidEndpoint(t *testing.T) {
	cfg := FileStorageConfig{
		BucketName: "test-bucket",
		OSSEnabled: true,
		Endpoint:   "ftp://files.example.com",
		AccessKey:  "access-key",
		SecretKey:  "secret-key",
		Region:     "us-east-1",
		ServerURL:  "https://files.example.com",
	}

	require.ErrorContains(t, cfg.Validate(), "FILE_OSS_ENDPOINT")
}

func TestSaveOSSFile(t *testing.T) {
	var requestPath string
	var requestBody []byte
	var payloadHash string
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.EscapedPath()
		requestBody, _ = io.ReadAll(r.Body)
		payloadHash = r.Header.Get("X-Amz-Content-Sha256")
		authorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("FILE_OSS_ENABLE", "true")
	t.Setenv("FILE_BUCKET_NAME", "test-bucket")
	t.Setenv("FILE_OSS_ENDPOINT", server.URL)
	t.Setenv("FILE_OSS_ACCESS_KEY", "test-access-key")
	t.Setenv("FILE_OSS_SECRET_KEY", "test-secret-key")
	t.Setenv("FILE_OSS_REGION", "us-east-1")
	t.Setenv("FILE_OSS_SERVER_URL", "https://files.example.com")
	t.Setenv("FILE_LOCAL_ENABLE", "false")

	fileURL, err := SaveBytes(context.Background(), FileCategoryUpload, []byte("hello"), "hello.txt", "text/plain")
	require.NoError(t, err)
	month := time.Now().Format("200601")
	assert.Contains(t, fileURL, "/test-bucket/upload/"+month+"/")
	assert.True(t, strings.HasPrefix(requestPath, "/test-bucket/upload/"+month+"/"), requestPath)
	assert.Equal(t, []byte("hello"), requestBody)
	assert.NotEmpty(t, payloadHash)
	assert.Contains(t, authorization, "AWS4-HMAC-SHA256")
}
