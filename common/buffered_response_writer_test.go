package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferedResponseWriterCommit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	writer := NewBufferedResponseWriter(ctx.Writer)
	writer.Header().Set("X-Test", "buffered")
	writer.Header().Set("Content-Type", "application/custom")
	writer.WriteHeader(http.StatusCreated)
	_, err := writer.WriteString("upstream")
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, writer.Status())
	assert.Equal(t, "upstream", string(writer.BodyBytes()))
	assert.Equal(t, 0, recorder.Body.Len())

	writer.Header().Set("Content-Type", "application/json")
	require.NoError(t, writer.Commit([]byte(`{"url":"stored"}`)))

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "buffered", recorder.Header().Get("X-Test"))
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"url":"stored"}`, recorder.Body.String())
}
