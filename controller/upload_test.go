package controller

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	basePath := t.TempDir()
	t.Setenv("FILE_OSS_ENABLE", "false")
	t.Setenv("FILE_LOCAL_ENABLE", "true")
	t.Setenv("FILE_LOCAL_BASE_PATH", basePath)
	oldServerAddress := system_setting.ServerAddress
	system_setting.ServerAddress = "https://files.example.com"
	t.Cleanup(func() {
		system_setting.ServerAddress = oldServerAddress
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "hello.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request

	Upload(ctx)

	require.Equal(t, http.StatusOK, response.Code)
	var responseBody struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
		Message string `json:"message"`
		Success bool   `json:"success"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &responseBody))
	assert.True(t, responseBody.Success)
	assert.Empty(t, responseBody.Message)
	fileURL := responseBody.Data.URL
	month := time.Now().Format("200601")
	assert.Contains(t, fileURL, "/upload/"+month+"/")
	key := strings.TrimPrefix(fileURL, "https://files.example.com/")
	stored, err := os.ReadFile(filepath.Join(basePath, filepath.FromSlash(key)))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(stored))
}
