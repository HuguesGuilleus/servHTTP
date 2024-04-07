package template

import (
	"testing"
)

// Tests to prevent regression.

func TestError404(t *testing.T) {
	expected := `<!DOCTYPE html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1.0"><title>404 Not Found</title><style>body{max-width:60ex;margin:20vh auto 0;font-family:monospace;font-size:xx-large;background:#eae5dc;border:dodgerblue solid 0.3ex;border-style:solid none;padding:2ex 0}h1,#p{display:table;padding:0.2em 0.5em;background:#FFF}a{color:#06C;text-decoration:none}a:hover{color:#00B;text-decoration:underline}</style></head><body><h1>404 Not Found</h1><div id=p><a href="/">/</a><a href="/file/">file/</a></div>`
	assertString(t, expected, string(Error404("/file/")))
}
func TestError405(t *testing.T) {
	expected := `<!DOCTYPE html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1.0"><title>405 Method Not Allowed</title><style>body{max-width:60ex;margin:20vh auto 0;font-family:monospace;font-size:xx-large;background:#eae5dc;border:dodgerblue solid 0.3ex;border-style:solid none;padding:2ex 0}h1,#p{display:table;padding:0.2em 0.5em;background:#FFF}a{color:#06C;text-decoration:none}a:hover{color:#00B;text-decoration:underline}</style></head><body><h1>405 Method Not Allowed</h1><div id=p><a href="/">/</a><a href="/file/">file/</a></div>`
	assertString(t, expected, string(Error405("/file/")))
}
func TestError500(t *testing.T) {
	expected := `<!DOCTYPE html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1.0"><title>500 Internal Error</title><style>body{max-width:60ex;margin:20vh auto 0;font-family:monospace;font-size:xx-large;background:#eae5dc;border:dodgerblue solid 0.3ex;border-style:solid none;padding:2ex 0}h1,#p{display:table;padding:0.2em 0.5em;background:#FFF}a{color:#06C;text-decoration:none}a:hover{color:#00B;text-decoration:underline}</style></head><body><h1>500 Internal Error</h1><div id=p><a href="/">/</a><a href="/file/">file/</a></div>`
	assertString(t, expected, string(Error500("/file/")))
}

func TestError502(t *testing.T) {
	expected := `<!DOCTYPE html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1.0"><title>502 Bad Gateway</title><style>body{max-width:60ex;margin:20vh auto 0;font-family:monospace;font-size:xx-large;background:#eae5dc;border:dodgerblue solid 0.3ex;border-style:solid none;padding:2ex 0}h1,#p{display:table;padding:0.2em 0.5em;background:#FFF}a{color:#06C;text-decoration:none}a:hover{color:#00B;text-decoration:underline}</style></head><body><h1>502 Bad Gateway</h1><div id=p><a href="/">/</a><a href="/file/">file/</a></div>`
	assertString(t, expected, string(Error502("/file/")))
}
