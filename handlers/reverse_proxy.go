package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
)

func ReverseProxy(logger *slog.Logger, rawURL, _ string) http.Handler {
	targetURL, err := url.Parse(rawURL)
	if err != nil {
		logger.Error("reverseParseURL", "rawURL", rawURL, "err", err.Error())
		return http.NotFoundHandler()
	}

	return &httputil.ReverseProxy{
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.Header.Del("X-Forwarded-For")
			r.Out.Host = targetURL.Host
			r.SetURL(targetURL)
			r.SetXForwarded()
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			const status = http.StatusBadGateway
			LogRequest(logger.With("err", err.Error()), status, r)
			servHTML(w, status, template.Error502(r.URL.Path))
		},
		ModifyResponse: func(w *http.Response) error {
			LogRequest(logger, w.StatusCode, w.Request)
			return nil
		},
	}
}
