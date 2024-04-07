package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureServeHTTP(t *testing.T) {
	r := httptest.NewRequest("X", "http://sub.example.com/dir/yolo?a=1", nil)
	w := httptest.NewRecorder()
	logger, logBuff := testLoggerOne()

	s := Secure(logger, "", "")
	s.ServeHTTP(w, r)

	assert.Equal(t, 308, w.Code)
	assert.Equal(t, "https://sub.example.com/dir/yolo?a=1", w.Header().Get("location"))
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, "level=INFO msg=http s=308 ip=192.0.2.1:1234 h=sub.example.com m=X u=/dir/yolo\n", logBuff.String())
}
