package handlers

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
)

type cacheHandler struct {
	common
	files map[string]*cacheFile
}

type cacheFile struct {
	// True if the file is a directory (generated or not)
	isDir bool

	// The mime type, used in Content-Type header
	contentType string

	// Last modified from the filesystem
	modTime time.Time
	// Last-Modified header
	modString string
	// Etag (sha256sum)
	etag string

	// Identity content without compression
	identityBytes []byte
	identityLen   string
	// Compressed version with "deflate"
	// May be nil if it bigger than no compression.
	deflateBytes []byte
	deflateLen   string
}

func (hand *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hand.Serve(w, r) {
		return
	}

	file := hand.files[strings.TrimPrefix(path.Clean(r.URL.Path), "/")]
	if file == nil {
		LogRequest(hand.Logger, http.StatusNotFound, r)
		servHTML(w, http.StatusNotFound, template.Error404(r.URL.Path))
		return
	} else if hand.endSlash(w, r, file.isDir) {
		return
	}

	w.Header().Add(headerContentType, file.contentType)
	w.Header().Add(headerLastModified, file.modString)
	w.Header().Add(headerETag, file.etag)
	if file.etag == r.Header.Get(headerIfNoneMatch) {
		LogRequest(hand.Logger, http.StatusNotModified, r)
		w.WriteHeader(http.StatusNotModified)
	} else if len(file.deflateBytes) > 0 && acceptDeflate(r.Header.Get(headerAcceptEncoding)) {
		LogRequest(hand.Logger, http.StatusOK, r)
		w.Header().Add(headerContentEncoding, deflateEncoding)
		w.Header().Add(headerContentLength, file.deflateLen)
		w.Write(file.deflateBytes)
	} else {
		LogRequest(hand.Logger, http.StatusOK, r)
		w.Header().Add(headerContentLength, file.identityLen)
		w.Write(file.identityBytes)
	}
}

func acceptDeflate(header string) bool {
	values := strings.FieldsFunc(header, func(r rune) bool { return !unicode.IsLetter(r) })
	for _, v := range values {
		if v == deflateEncoding {
			return true
		}
	}
	return false
}

// Create a file handler without memory copy of file.
// All 20 seconds, update the index.
func Cache(logger *slog.Logger, root, cacheControl string) http.Handler {
	hand := new(cacheHandler)
	hand.Logger = logger
	hand.CacheControl = cacheControl
	hand.files = make(map[string]*cacheFile)

	go func() {
		fsys := os.DirFS(root)
		hand.Update(fsys, time.Now())
		for now := range time.Tick(time.Second * 20) {
			hand.Update(fsys, now)
		}
	}()

	return hand
}

func (hand *cacheHandler) Update(fsys fs.FS, now time.Time) {
	now = now.UTC()
	newFiles := make(map[string]*cacheFile, len(hand.files))
	needIndex := make(map[string]struct{})
	dirs := make(map[string][]fs.FileInfo)

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		dir := path.Dir(p)
		dirs[dir] = append(dirs[dir], info)
		exactPath := p
		isIndex := false
		if info.IsDir() {
			needIndex[p] = struct{}{}
			return nil
		} else if info.Name() == "index.html" {
			p = dir
			delete(needIndex, p)
			isIndex = true
		}

		oldFile := hand.files[p]
		if oldFile != nil && oldFile.modTime.Equal(info.ModTime()) {
			newFiles[p] = oldFile
		} else {
			content, err := fs.ReadFile(fsys, exactPath)
			if err != nil {
				return err
			}
			newFiles[p] = newCacheFile(content, info.Name(), isIndex, info.ModTime())
		}

		return nil
	})
	if err != nil {
		hand.Logger.Warn("cache-update-fail", "err", err.Error())
		return
	}

	for p := range needIndex {
		newFiles[p] = newCacheFile(template.Index(p+"/", dirs[p]), htmlMIME, true, now)
	}

	newFiles[""] = newFiles["."]
	delete(newFiles, ".")

	hand.files = newFiles
}

func newCacheFile(content []byte, name string, isDir bool, lastModified time.Time) (file *cacheFile) {
	file = new(cacheFile)
	file.isDir = isDir
	file.contentType = mime.TypeByExtension(path.Ext(name))

	lastModified = lastModified.UTC()
	file.modTime = lastModified
	file.modString = lastModified.Format(time.RFC1123)

	hash := sha256.Sum256(content)
	file.etag = "\"" + base64.RawURLEncoding.EncodeToString(hash[:]) + "\""

	file.identityBytes = content
	file.identityLen = strconv.Itoa(len(content))

	buff := bytes.Buffer{}
	enc, _ := flate.NewWriter(&buff, flate.BestCompression)
	enc.Write(content)
	enc.Close()
	if buff.Len() < len(content) {
		file.deflateBytes = buff.Bytes()
		file.deflateLen = strconv.Itoa(buff.Len())
	}

	return
}
