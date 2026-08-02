package common

import (
	"bytes"
	"sync"

	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
)

// ResponseBodyRecorder records the response body while forwarding it to the
// original Gin response writer.
type ResponseBodyRecorder struct {
	gin.ResponseWriter

	mu   sync.RWMutex
	body bytes.Buffer
}

func NewResponseBodyRecorder(writer gin.ResponseWriter) *ResponseBodyRecorder {
	return &ResponseBodyRecorder{ResponseWriter: writer}
}

func (w *ResponseBodyRecorder) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.ResponseWriter.Write(data)
	if n > 0 {
		_, _ = w.body.Write(data[:n])
	}
	return n, err
}

func (w *ResponseBodyRecorder) WriteString(data string) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.ResponseWriter.WriteString(data)
	if n > 0 {
		_, _ = w.body.WriteString(data[:n])
	}
	return n, err
}

func (w *ResponseBodyRecorder) BodyString() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.body.String()
}

// StartResponseBodyCapture wraps the current Gin writer once and stores the
// recorder in the request context for later log persistence.
func StartResponseBodyCapture(c *gin.Context) *ResponseBodyRecorder {
	if recorder, ok := GetContextKeyType[*ResponseBodyRecorder](c, constant.ContextKeyResponseBodyRecorder); ok {
		return recorder
	}

	recorder := NewResponseBodyRecorder(c.Writer)
	c.Writer = recorder
	SetContextKey(c, constant.ContextKeyResponseBodyRecorder, recorder)
	return recorder
}

func GetCapturedResponseBody(c *gin.Context) string {
	recorder, ok := GetContextKeyType[*ResponseBodyRecorder](c, constant.ContextKeyResponseBodyRecorder)
	if !ok {
		return ""
	}
	return recorder.BodyString()
}
