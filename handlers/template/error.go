package template

import (
	"bytes"
	_ "embed"
	"regexp"
)

var (
	//go:embed error.html
	errorRaw []byte
	error404 []byte
	error405 []byte
	error500 []byte
	error502 []byte
)

func init() {
	errorRaw = minify(errorRaw)
	error404 = bytes.ReplaceAll(errorRaw, []byte("TITLE"), []byte("404 Not Found"))
	error405 = bytes.ReplaceAll(errorRaw, []byte("TITLE"), []byte("405 Method Not Allowed"))
	error500 = bytes.ReplaceAll(errorRaw, []byte("TITLE"), []byte("500 Internal Error"))
	error502 = bytes.ReplaceAll(errorRaw, []byte("TITLE"), []byte("502 Bad Gateway"))
}

func Error404(path string) []byte { return errorMake(path, error404) }
func Error405(path string) []byte { return errorMake(path, error405) }
func Error500(path string) []byte { return errorMake(path, error500) }
func Error502(path string) []byte { return errorMake(path, error502) }

func errorMake(path string, template []byte) []byte {
	buff := bytes.Buffer{}
	buff.Write(template)
	htmlPath(&buff, path)
	return buff.Bytes()
}

var minifyRegexp1 = regexp.MustCompile(`(\W)\s+`)
var minifyRegexp2 = regexp.MustCompile(`\s+(\W)`)

func minify(input []byte) []byte {
	input = minifyRegexp1.ReplaceAll(input, []byte("$1"))
	input = minifyRegexp2.ReplaceAll(input, []byte("$1"))
	return input
}
