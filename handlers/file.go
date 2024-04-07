package handlers

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"

	"github.com/HuguesGuilleus/servHTTP/handlers/template"
)

type fileHandler struct {
	common
	fsys http.FileSystem
}

func File(logger *slog.Logger, root, cacheControl string) http.Handler {
	return &fileHandler{
		common: common{
			Logger:       logger,
			CacheControl: cacheControl,
		},
		fsys: http.Dir(root),
	}
}

func (hand *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hand.common.Serve(w, r) {
		return
	}

	file, stat, err := open(hand.fsys, path.Clean(r.URL.Path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			LogRequest(hand.Logger, http.StatusNotFound, r)
			servHTML(w, http.StatusNotFound, template.Error404(r.URL.Path))
		} else {
			LogRequest(hand.Logger.With("err", err.Error()), http.StatusInternalServerError, r)
			servHTML(w, http.StatusInternalServerError, template.Error500(r.URL.Path))
		}
		return
	}
	defer file.Close()

	if hand.endSlash(w, r, stat.IsDir()) {
		return
	}

	if stat.IsDir() {
		index, info, err := open(hand.fsys, r.URL.Path+"index.html")
		if err == nil {
			defer index.Close()
			if !info.IsDir() {
				LogRequest(hand.Logger, http.StatusOK, r)
				http.ServeContent(w, r, "index.html", info.ModTime(), index)
				return
			}
		}

		entries, err := file.Readdir(0)
		if err != nil {
			LogRequest(hand.Logger.With("err", err.Error()), http.StatusInternalServerError, r)
			servHTML(w, http.StatusInternalServerError, template.Error500(r.URL.Path))
			return
		}
		LogRequest(hand.Logger, http.StatusOK, r)
		servHTML(w, http.StatusOK, template.Index(r.URL.Path, entries))
	} else {
		LogRequest(hand.Logger, http.StatusOK, r)
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	}
}

func open(fsys http.FileSystem, name string) (file http.File, info fs.FileInfo, err error) {
	file, err = fsys.Open(name)
	if err != nil {
		return
	}
	info, err = file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return
}
