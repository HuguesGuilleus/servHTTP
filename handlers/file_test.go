package handlers

import (
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
	"github.com/stretchr/testify/assert"
)

var testFS = fstest.MapFS{
	"autoindex/file.txt": &fstest.MapFile{Data: []byte("file")},
	"dir":                &fstest.MapFile{Mode: fs.ModeDir},
	"hello.txt":          &fstest.MapFile{Data: []byte("Hello World")},
	"index/index.html":   &fstest.MapFile{Data: []byte("the index")},
}

func TestFile(t *testing.T) {
	logger, _ := testLoggerOne()
	assert.Equal(t, &fileHandler{
		common: common{logger, "cache"},
		fsys:   http.Dir("fs"),
	}, File(logger, "fs", "cache"))
}

func testFileHandler(r *http.Request) (*httptest.ResponseRecorder, *bytes.Buffer) {
	logger, logBuffer := testLoggerOne()
	hand := fileHandler{
		common: common{Logger: logger},
		fsys:   http.FS(testFS),
	}

	w := httptest.NewRecorder()
	hand.ServeHTTP(w, r)

	return w, logBuffer
}

func TestFileImplementCommon(t *testing.T) {
	logger, logBuffer := testLoggerOne()
	testCommonHandler(t, &fileHandler{
		common: common{Logger: logger, CacheControl: "max-age=900"},
	}, logBuffer)
}

func TestFile404(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/yolo.txt", nil)
	w, logBuffer := testFileHandler(r)
	assert.Equal(t, 404, w.Code)
	assert.Equal(t, template.Error404("/yolo.txt"), w.Body.Bytes())
	assert.Equal(t, "level=INFO msg=http s=404 ip=192.0.2.1:1234 h=example.com m=GET u=/yolo.txt\n", logBuffer.String())
}

func TestFileHandlerMissingSlash(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/dir", nil)
	w, logBuffer := testFileHandler(r)
	assert.Equal(t, 301, w.Code)
	assert.Equal(t, "/dir/", w.Header().Get("location"))
	assert.Equal(t, "level=INFO msg=http s=301 ip=192.0.2.1:1234 h=example.com m=GET u=/dir\n", logBuffer.String())
}
func TestFileHandlerUnneccessarySlash(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/hello.txt/?b=2", nil)
	w, logBuffer := testFileHandler(r)
	assert.Equal(t, 301, w.Code)
	assert.Equal(t, "/hello.txt?b=2", w.Header().Get("location"))
	assert.Equal(t, "level=INFO msg=http s=301 ip=192.0.2.1:1234 h=example.com m=GET u=/hello.txt/\n", logBuffer.String())
}

func TestFileIndex(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/index/", nil)
	w, logBuffer := testFileHandler(r)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "the index", w.Body.String())
	assert.Equal(t, "level=INFO msg=http s=200 ip=192.0.2.1:1234 h=example.com m=GET u=/index/\n", logBuffer.String())
}

func TestFileAutoIndex(t *testing.T) {
	info, err := (&testFS).Stat("autoindex/file.txt")
	assert.NoError(t, err)

	r := httptest.NewRequest("GET", "http://example.com/autoindex/", nil)
	w, logBuffer := testFileHandler(r)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(template.Index("/autoindex/", []fs.FileInfo{info})), w.Body.String())
	assert.Equal(t, "level=INFO msg=http s=200 ip=192.0.2.1:1234 h=example.com m=GET u=/autoindex/\n", logBuffer.String())
}

func TestFileHandlerOKFile(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/hello.txt", nil)
	w, logBuffer := testFileHandler(r)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())
	assert.Equal(t, "level=INFO msg=http s=200 ip=192.0.2.1:1234 h=example.com m=GET u=/hello.txt\n", logBuffer.String())
}

func TestFileFailerFS(t *testing.T) {
	tf := func(t *testing.T, err string, fsys http.FileSystem) {
		logger, logBuffer := testLoggerOne()
		hand := fileHandler{
			common: common{Logger: logger},
			fsys:   fsys,
		}

		r := httptest.NewRequest("GET", "http://example.com/", nil)
		w := httptest.NewRecorder()
		hand.ServeHTTP(w, r)
		assert.Equal(t, 500, w.Code)
		assert.Equal(t, template.Error500("/"), w.Body.Bytes())
		assert.Equal(t, "level=WARN msg=http err=\""+err+"\" s=500 ip=192.0.2.1:1234 h=example.com m=GET u=/\n", logBuffer.String())
	}
	t.Run("open", func(t *testing.T) {
		tf(t, "open fail", failWhenOpenFSH{})
	})
	t.Run("stat", func(t *testing.T) {
		tf(t, "stat fail", failWhenStatFSH{})
	})
	t.Run("readdir", func(t *testing.T) {
		tf(t, "readdir fail", failWhenReaddirFSH{})
	})
}

// A http.FileSystem that always fail with .Open()
type failWhenOpenFSH struct{}

func (failWhenOpenFSH) Open(string) (http.File, error) {
	return nil, errors.New("open fail")
}

type failWhenStatFSH struct{}

func (failWhenStatFSH) Open(string) (http.File, error) {
	return failWhenStatFile{}, nil
}

type failWhenStatFile struct{}

func (failWhenStatFile) Close() error                                 { return nil }
func (failWhenStatFile) Read(data []byte) (int, error)                { return len(data), nil }
func (failWhenStatFile) Seek(offset int64, whence int) (int64, error) { return 0, nil }
func (failWhenStatFile) Readdir(count int) ([]fs.FileInfo, error)     { return nil, nil }
func (failWhenStatFile) Stat() (fs.FileInfo, error) {
	return nil, errors.New("stat fail")
}

type failWhenReaddirFSH struct{}

func (failWhenReaddirFSH) Open(string) (http.File, error) {
	return failWhenReaddirDir{}, nil
}

type failWhenReaddirDir struct{}

func (failWhenReaddirDir) Close() error                                 { return nil }
func (failWhenReaddirDir) Read(data []byte) (int, error)                { return len(data), nil }
func (failWhenReaddirDir) Seek(offset int64, whence int) (int64, error) { return 0, nil }
func (failWhenReaddirDir) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, errors.New("readdir fail")
}
func (failWhenReaddirDir) Stat() (fs.FileInfo, error) {
	return failWhenReaddirInfo{}, nil
}

type failWhenReaddirInfo struct{}

func (failWhenReaddirInfo) Name() string       { return "" }          // base name of the file
func (failWhenReaddirInfo) Size() int64        { return 0 }           // length in bytes for regular files; system-dependent for others
func (failWhenReaddirInfo) Mode() fs.FileMode  { return fs.ModeDir }  // file mode bits
func (failWhenReaddirInfo) ModTime() time.Time { return time.Time{} } // modification time
func (failWhenReaddirInfo) IsDir() bool        { return true }        // abbreviation for Mode().IsDir()
func (failWhenReaddirInfo) Sys() any           { return nil }         // underlying data source (can return nil)
