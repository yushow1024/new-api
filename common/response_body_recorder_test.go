package common

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponseBodyRecorder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)

	recorder := StartResponseBodyCapture(ctx)
	if _, err := ctx.Writer.Write([]byte(`{"data":`)); err != nil {
		t.Fatalf("write response bytes: %v", err)
	}
	if _, err := ctx.Writer.WriteString(`[]}`); err != nil {
		t.Fatalf("write response string: %v", err)
	}

	const want = `{"data":[]}`
	if got := recorder.BodyString(); got != want {
		t.Fatalf("recorded response = %q, want %q", got, want)
	}
	if got := response.Body.String(); got != want {
		t.Fatalf("forwarded response = %q, want %q", got, want)
	}
	if got := GetCapturedResponseBody(ctx); got != want {
		t.Fatalf("context response = %q, want %q", got, want)
	}
	if got := StartResponseBodyCapture(ctx); got != recorder {
		t.Fatal("response writer was wrapped more than once")
	}
}
