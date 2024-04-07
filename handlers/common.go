package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
)

var (
	headerAcceptEncoding  = http.CanonicalHeaderKey("Accept-Encoding")
	headerCacheControl    = http.CanonicalHeaderKey("cache-control")
	headerContentEncoding = http.CanonicalHeaderKey("Content-Encoding")
	headerContentLength   = http.CanonicalHeaderKey("Content-Length")
	headerContentType     = http.CanonicalHeaderKey("Content-Type")
	headerETag            = http.CanonicalHeaderKey("ETag")
	headerIfNoneMatch     = http.CanonicalHeaderKey("If-None-Match")
	headerLastModified    = http.CanonicalHeaderKey("Last-Modified")

	htmlMIME        = "text/html"
	deflateEncoding = "deflate"
)

// common operation for file serving.
type common struct {
	// Log each request
	Logger *slog.Logger
	// CacheControl control header.
	CacheControl string
}

// Add Cache control if any.
// Then Manage request for a static file system:
// - Reject method is different of HEAD and GET
// - Redirect request with url "/index.html" end
// else return false
func (hand *common) Serve(w http.ResponseWriter, r *http.Request) bool {
	if hand.CacheControl != "" {
		w.Header().Add(headerCacheControl, hand.CacheControl)
	}

	switch r.Method {
	case "GET", "HEAD":
	default:
		LogRequest(hand.Logger, http.StatusMethodNotAllowed, r)
		servHTML(w, http.StatusMethodNotAllowed, template.Error405(r.URL.Path))
		return true
	}

	if strings.HasSuffix(r.URL.Path, "/index.html") {
		LogRequest(hand.Logger, http.StatusPermanentRedirect, r)
		p := strings.TrimSuffix(r.URL.Path, "index.html")
		if r.URL.RawQuery != "" {
			p += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, p, http.StatusPermanentRedirect)
		return true
	}

	return false
}

func (hand *common) endSlash(w http.ResponseWriter, r *http.Request, isDir bool) bool {
	endSlash := strings.HasSuffix(r.URL.Path, "/")
	if isDir != endSlash {
		p := r.URL.Path
		if endSlash {
			p = strings.TrimSuffix(p, "/")
		} else {
			p += "/"
		}
		if r.URL.RawQuery != "" {
			p += "?" + r.URL.RawQuery
		}
		LogRequest(hand.Logger, http.StatusMovedPermanently, r)
		http.Redirect(w, r, p, http.StatusMovedPermanently)
		return true
	}

	return false
}

func LogRequest(logger *slog.Logger, status int, r *http.Request) {
	l := logger.With("s", status, "ip", r.RemoteAddr, "h", r.Host, "m", r.Method, "u", r.URL.Path)

	if status < http.StatusInternalServerError {
		l.Info("http")
	} else {
		l.Warn("http")
	}
}

func servHTML(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set(headerContentType, htmlMIME)
	w.Header().Set(headerContentLength, strconv.Itoa(len(body)))
	w.WriteHeader(status)
	w.Write(body)
}
