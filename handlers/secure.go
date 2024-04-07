package handlers

import (
	"log/slog"
	"net/http"
)

type secureHandler struct {
	Logger *slog.Logger
}

func Secure(logger *slog.Logger, _, _ string) http.Handler {
	return &secureHandler{logger}
}

func (hand *secureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	LogRequest(hand.Logger, http.StatusPermanentRedirect, r)
	r.URL.Scheme = "https"
	r.URL.Host = r.Host
	http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
}
