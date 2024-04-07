package handlers

import (
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
	"github.com/stretchr/testify/assert"
)

func TestCacheImplementCommon(t *testing.T) {
	logger, logBuffer := testLoggerOne()
	testCommonHandler(t, &cacheHandler{
		common: common{Logger: logger, CacheControl: "max-age=900"},
	}, logBuffer)
}

func newTestCache(r *http.Request) (*httptest.ResponseRecorder, *bytes.Buffer) {
	logger, logBuffer := testLoggerOne()

	w := httptest.NewRecorder()
	(&cacheHandler{
		common: common{Logger: logger},
		files: map[string]*cacheFile{
			"file": {
				contentType:   "text/plain; charset=utf-8",
				etag:          `"etagT"`,
				modString:     "Wed, 21 Oct 2015 07:28:00 GMT",
				identityBytes: []byte("text"),
				identityLen:   "4",
				deflateBytes:  []byte("co"),
				deflateLen:    "2",
			},
			"compress": {
				contentType:  "text/plain; charset=utf-8",
				etag:         `"etagC"`,
				modString:    "Wed, 22 Oct 2015 07:28:00 GMT",
				deflateBytes: []byte("co"),
				deflateLen:   "2",
			},
		},
	}).ServeHTTP(w, r)

	return w, logBuffer
}

func TestCacheNotFound(t *testing.T) {
	r := httptest.NewRequest("GET", "http://host/x", nil)
	w, logBuffer := newTestCache(r)

	assert.Equal(t, 404, w.Code)
	assert.Equal(t, htmlMIME, w.Header().Get(headerContentType))
	assert.Equal(t, template.Error404("/x"), w.Body.Bytes())

	assert.Equal(t, "level=INFO msg=http s=404 ip=192.0.2.1:1234 h=host m=GET u=/x\n", logBuffer.String())
}

func TestCacheEndSlash(t *testing.T) {
	r := httptest.NewRequest("GET", "http://host/file/", nil)
	w, logBuffer := newTestCache(r)

	assert.Equal(t, 301, w.Code)
	assert.Equal(t, "/file", w.Header().Get("location"))

	assert.Equal(t, "level=INFO msg=http s=301 ip=192.0.2.1:1234 h=host m=GET u=/file/\n", logBuffer.String())
}

func TestCacheWithEtag(t *testing.T) {
	r := httptest.NewRequest("GET", "http://host/file", nil)
	r.Header.Add("if-none-match", "\"etagT\"")
	w, logBuffer := newTestCache(r)

	assert.Equal(t, 304, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get(headerContentType))
	assert.Equal(t, `"etagT"`, w.Header().Get(headerETag))
	assert.Equal(t, "Wed, 21 Oct 2015 07:28:00 GMT", w.Header().Get(headerLastModified))
	assert.Equal(t, "", w.Header().Get(headerContentLength))
	assert.Equal(t, "", w.Body.String())

	assert.Equal(t, "level=INFO msg=http s=304 ip=192.0.2.1:1234 h=host m=GET u=/file\n", logBuffer.String())
}

func TestCacheCompress(t *testing.T) {
	r := httptest.NewRequest("GET", "http://host/compress", nil)
	r.Header.Add("accept-encoding", "deflate, gzip;q=1.0, *;q=0.5")
	w, logBuffer := newTestCache(r)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get(headerContentType))
	assert.Equal(t, `"etagC"`, w.Header().Get(headerETag))
	assert.Equal(t, "Wed, 22 Oct 2015 07:28:00 GMT", w.Header().Get(headerLastModified))
	assert.Equal(t, "2", w.Header().Get(headerContentLength))
	assert.Equal(t, "co", w.Body.String())

	assert.Equal(t, "level=INFO msg=http s=200 ip=192.0.2.1:1234 h=host m=GET u=/compress\n", logBuffer.String())
}

func TestCacheNormal(t *testing.T) {
	r := httptest.NewRequest("GET", "http://host/file", nil)
	w, logBuffer := newTestCache(r)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get(headerContentType))
	assert.Equal(t, `"etagT"`, w.Header().Get(headerETag))
	assert.Equal(t, "Wed, 21 Oct 2015 07:28:00 GMT", w.Header().Get(headerLastModified))
	assert.Equal(t, "4", w.Header().Get(headerContentLength))
	assert.Equal(t, "text", w.Body.String())

	assert.Equal(t, "level=INFO msg=http s=200 ip=192.0.2.1:1234 h=host m=GET u=/file\n", logBuffer.String())
}

func TestCacheUpdateOK(t *testing.T) {
	autoindex_file := &cacheFile{
		isDir:         false,
		contentType:   "text/plain; charset=utf-8",
		modString:     "Mon, 01 Jan 0001 00:00:00 UTC",
		etag:          `"O5w1jzbwoxtq0-FPMJx88ZiskkboMW-c5UPVsZrAK4A"`,
		identityBytes: []byte("file"),
		identityLen:   "4",
	}

	now := time.Now()
	logger, logBuffer := testLoggerOne()
	hand := cacheHandler{common: common{logger, ""}, files: map[string]*cacheFile{
		"autoindex/file.txt": autoindex_file,
	}}

	hand.Update(testFS, now)

	// files
	assert.Equal(t, &cacheFile{
		isDir:         false,
		contentType:   "text/plain; charset=utf-8",
		modString:     "Mon, 01 Jan 0001 00:00:00 UTC",
		etag:          `"pZGm1Av0IEBKARczz7exkNYsZb8LzaMrV7J32a2fFG4"`,
		identityBytes: []byte("Hello World"),
		identityLen:   "11",
	}, hand.files["hello.txt"])
	assert.Equal(t, &cacheFile{
		isDir:         true,
		contentType:   "text/html; charset=utf-8",
		modString:     "Mon, 01 Jan 0001 00:00:00 UTC",
		etag:          `"KBS6GuR-Ifb0nlVoOVLTLuy6QjM_EsZDSUqhq8wXOgo"`,
		identityBytes: []byte("the index"),
		identityLen:   "9",
	}, hand.files["index"])
	assert.Same(t, autoindex_file, hand.files["autoindex/file.txt"])

	// auto index
	assert.NotNil(t, hand.files["autoindex"])
	assert.NotNil(t, hand.files["dir"])
	assert.NotNil(t, hand.files[""])

	assert.Len(t, hand.files, 6)

	assert.Equal(t, "", logBuffer.String())
}

func TestCacheUpdateFail(t *testing.T) {
	logger, logBuffer := testLoggerOne()
	hand := cacheHandler{
		common: common{Logger: logger},
		files:  make(map[string]*cacheFile),
	}
	hand.Update(failWhenOpenFS{}, time.Now())
	assert.Equal(t, "level=WARN msg=cache-update-fail err=\"open fail\"\n", logBuffer.String())
}

// A http.FileSystem that always fail with .Open()
type failWhenOpenFS struct{}

func (failWhenOpenFS) Open(string) (fs.File, error) {
	return nil, errors.New("open fail")
}
