package common

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BufferedResponseWriter buffers a non-streaming response so it can be
// transformed before anything is sent to the client.
type BufferedResponseWriter struct {
	gin.ResponseWriter
	header      http.Header
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func NewBufferedResponseWriter(writer gin.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		ResponseWriter: writer,
		header:         writer.Header().Clone(),
		status:         http.StatusOK,
	}
}

func (w *BufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *BufferedResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

func (w *BufferedResponseWriter) WriteHeaderNow() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
}

func (w *BufferedResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	return w.body.Write(data)
}

func (w *BufferedResponseWriter) WriteString(data string) (int, error) {
	w.WriteHeaderNow()
	return w.body.WriteString(data)
}

func (w *BufferedResponseWriter) Status() int { return w.status }
func (w *BufferedResponseWriter) Size() int   { return w.body.Len() }
func (w *BufferedResponseWriter) Written() bool {
	return w.wroteHeader
}

func (w *BufferedResponseWriter) BodyBytes() []byte {
	return append([]byte(nil), w.body.Bytes()...)
}

func (w *BufferedResponseWriter) Reset() {
	w.header = w.ResponseWriter.Header().Clone()
	w.status = http.StatusOK
	w.wroteHeader = false
	w.body.Reset()
}

func (w *BufferedResponseWriter) Commit(body []byte) error {
	targetHeader := w.ResponseWriter.Header()
	for key := range targetHeader {
		targetHeader.Del(key)
	}
	for key, values := range w.header {
		targetHeader[key] = append([]string(nil), values...)
	}
	w.ResponseWriter.WriteHeader(w.status)
	_, err := w.ResponseWriter.Write(body)
	return err
}
