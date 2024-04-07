package template

import (
	"io/fs"
	"testing"
	"time"
)

var (
	indexBegin = `<!DOCTYPE html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=0.5"><title>Index</title><style>body{margin:1rem;font-family:monospace;font-size:xx-large;background:#eae5dc}nav{display:grid;grid-template-columns:auto 1fr;grid-gap:1rem;margin:1rem 0}h1,nav>*{margin:0;display:table;padding:0.2em 0.5em;background:#FFF}input{width:100%;border:none;font:inherit}a,input{color:#06C;text-decoration:none}a:hover{color:#00B;text-decoration:underline}#f{display:grid;grid-template-columns:auto auto 1fr;grid-gap:0 1rem}time,#f>div{color:#0009}#f>div{text-align:right}time::before{content:"["}time::after{content:"]"}pre{white-space:pre-wrap}</style></head><body><h1>Index</h1><nav><div id=p><a href="/">/</a><a href="/file/">file/</a></div><input type=search id=s placeholder=Search autofocus></nav><div id=f>`
	indexEnd   = `<script>(async(q=(t,f)=>document.querySelectorAll(t).forEach(f),P="previousElementSibling",H="hidden",T="innerText")=>{q("time",t=>t[T]=new Date(t.dateTime).toLocaleString());s.oninput=_=>q("#f>a",a=>a[P][P][H]=a[P][H]=a[H]=!a[T].includes(s.value));if(r&&r[T])r[T]=await(await fetch(r[T])).text()})();</script>`
)

func TestIndexEmpty(t *testing.T) {
	expected := indexBegin + `</div>` + indexEnd
	assertString(t, expected, string(Index("/file/", nil)))
}

func TestIndexSome(t *testing.T) {
	expected := indexBegin +
		`<br><div>-</div><a href="a/">a/</a>` +
		`<br><div>-</div><a href="b/">b/</a>` +
		`<time datetime="2023-12-02T12:23:25Z">2023-12-02T12:23:25Z</time><div>123 456 B</div><a href="A&amp;a">A&amp;a</a>` +
		`<time datetime="2023-12-02T12:23:25Z">2023-12-02T12:23:25Z</time><div>123 456 B</div><a href="README.md">README.md</a>` +
		`</div><pre id=r>README.md</pre>` + indexEnd

	modTime := time.Date(2023, 12, 2, 13, 23, 25, 123, time.FixedZone("Paris", +1*3600))
	entrys := []fs.FileInfo{
		&Info{modTime: modTime, name: "README.md", size: 123_456},
		&Info{modTime: modTime, name: "A&a", size: 123_456},
		&Info{modTime: modTime, name: "b", size: 0},
		&Info{modTime: modTime, name: "a", size: 0},
		&Info{modTime: modTime, name: ".hide", size: 0},
	}

	assertString(t, expected, string(Index("/file/", entrys)))
}

// Implement fs.Info and fs.FileInfo
type Info struct {
	name    string
	size    int64 // 0 if is dir
	modTime time.Time
}

func (info *Info) Name() string       { return info.name }
func (info *Info) Size() int64        { return info.size }
func (*Info) Mode() fs.FileMode       { panic("do no use this method!") }
func (info *Info) ModTime() time.Time { return info.modTime }
func (info *Info) IsDir() bool        { return info.size == 0 }
func (*Info) Sys() any                { return nil }

// Assert to string are identical
func assertString(t *testing.T, a, b string) {
	if a == b {
		return
	}

	brunes := []rune(b)
	for i, ra := range a[:min(len(a), len(b))] {
		if ra != brunes[i] {
			t.Error("no equal string")
			t.Logf("a: %q", a[:i]+"█████"+a[i:])
			t.Logf("b: %q", b[:i]+"█████"+b[i:])
			t.FailNow()
			return
		}
	}

	if len(a) != len(b) {
		t.Logf("a: %q", a)
		t.Logf("b: %q", b)
		t.FailNow()
	}
}
