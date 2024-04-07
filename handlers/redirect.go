package handlers

import (
	"log/slog"
	"net/http"
	"strings"
)

type redirectHandler struct {
	Logger *slog.Logger
	URL    string
}

func Redirect(logger *slog.Logger, url, _ string) http.Handler {
	return &redirectHandler{logger, url}
}

func (hand *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	LogRequest(hand.Logger, http.StatusPermanentRedirect, r)

	u := hand.URL + strings.TrimPrefix(r.URL.Path, "/")
	if r.URL.RawQuery != "" {
		u += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, u, http.StatusPermanentRedirect)
}
