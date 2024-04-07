package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirectServeHTTP(t *testing.T) {
	logger, logLines := testLoggerLine()
	s := Redirect(logger, "https://www.example.com/root/", "")

	r := httptest.NewRequest("X", "http://sub.example.com/dir/yolo?a=1", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	assert.Equal(t, 308, w.Code)
	assert.Equal(t, "https://www.example.com/root/dir/yolo?a=1", w.Header().Get("location"))
	assert.Equal(t, 0, w.Body.Len())

	r = httptest.NewRequest("X", "http://sub.example.com/dir/yolo?a=1", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	assert.Equal(t, 308, w.Code)
	assert.Equal(t, "https://www.example.com/root/dir/yolo?a=1", w.Header().Get("location"))
	assert.Equal(t, 0, w.Body.Len())

	assert.Equal(t, []string{
		"level=INFO msg=http s=308 ip=192.0.2.1:1234 h=sub.example.com m=X u=/dir/yolo",
		"level=INFO msg=http s=308 ip=192.0.2.1:1234 h=sub.example.com m=X u=/dir/yolo",
		"",
	}, logLines())
}
