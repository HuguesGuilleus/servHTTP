package handlers

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
	"github.com/stretchr/testify/assert"
)

func TestCommonOK(t *testing.T) {
	for _, method := range []string{"GET", "HEAD"} {
		logger, logBuffer := testLoggerOne()
		r := httptest.NewRequest(method, "http://example.com/dir/", nil)
		w := httptest.NewRecorder()
		hand := common{Logger: logger, CacheControl: "max-age=900"}
		assert.False(t, hand.Serve(w, r))
		assert.Equal(t, "", logBuffer.String())
	}
}

func TestCommonFail(t *testing.T) {
	logger, logBuffer := testLoggerOne()
	c := &common{Logger: logger, CacheControl: "max-age=900"}
	testCommonHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, c.Serve(w, r))
	}), logBuffer)
}

// Test handler or inherit
func testCommonHandler(t *testing.T, hand http.Handler, logBuffer *bytes.Buffer) {
	t.Run("method", func(t *testing.T) {
		r := httptest.NewRequest("POST", "http://example.com/", nil)
		w := httptest.NewRecorder()
		hand.ServeHTTP(w, r)
		assert.Equal(t, w.Code, 405)
		assert.Equal(t, "text/html", w.Header().Get("content-type"))
		assert.Equal(t, "max-age=900", w.Header().Get("cache-control"))
		assert.Equal(t, template.Error405("/"), w.Body.Bytes())
		assert.Equal(t, "level=INFO msg=http s=405 ip=192.0.2.1:1234 h=example.com m=POST u=/\n", logBuffer.String())
		logBuffer.Reset()
	})
	t.Run("index", func(t *testing.T) {
		r := httptest.NewRequest("GET", "http://example.com/dir/index.html?a=2", nil)
		w := httptest.NewRecorder()
		hand.ServeHTTP(w, r)
		assert.Equal(t, w.Code, 308)
		assert.Equal(t, "max-age=900", w.Header().Get("cache-control"))
		assert.Equal(t, "/dir/?a=2", w.Header().Get("location"))
		assert.Equal(t, "level=INFO msg=http s=308 ip=192.0.2.1:1234 h=example.com m=GET u=/dir/index.html\n", logBuffer.String())
	})
}

func testLoggerOne() (*slog.Logger, *bytes.Buffer) {
	option := slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return a
		},
	}
	buff := bytes.NewBuffer(nil)
	return slog.New(slog.NewTextHandler(buff, &option)), buff
}

func testLoggerLine() (*slog.Logger, func() []string) {
	logger, buff := testLoggerOne()
	return logger, func() []string { return strings.Split(buff.String(), "\n") }
}
