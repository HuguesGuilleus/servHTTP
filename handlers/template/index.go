package template

import (
	"bytes"
	_ "embed"
	"html"
	"io/fs"
	"slices"
	"sort"
	"strings"
	"time"
)

var (
	//go:embed index-0.html
	indexRaw0 []byte
	//go:embed index-1.html
	indexRaw1 []byte
	//go:embed index-2.html
	indexRaw2 []byte
)

func init() {
	indexRaw0 = minify(indexRaw0)
	indexRaw1 = minify(indexRaw1)
	indexRaw2 = minify(indexRaw2)
}

func Index(path string, entrys []fs.FileInfo) []byte {
	// Prepare files for index
	entrys = slices.DeleteFunc(entrys, func(entry fs.FileInfo) bool {
		return strings.HasPrefix(entry.Name(), ".")
	})
	sort.Slice(entrys, func(i, j int) bool {
		if entrys[i].IsDir() != entrys[j].IsDir() {
			return entrys[i].IsDir()
		}
		return entrys[i].Name() < entrys[j].Name()
	})

	// Headers
	buff := bytes.Buffer{}
	buff.Write(indexRaw0)
	htmlPath(&buff, path)
	buff.Write(indexRaw1)

	// File list
	readme := ""
	for _, entry := range entrys {
		if entry.IsDir() {
			buff.WriteString(`<br><div>-</div>`)
		} else {
			t := entry.ModTime().UTC().Format(time.RFC3339)
			buff.WriteString(`<time datetime="`)
			buff.WriteString(t)
			buff.WriteString(`">`)
			buff.WriteString(t)
			buff.WriteString(`</time>`)

			buff.WriteString(`<div>`)
			htmlSize(&buff, entry.Size())
			buff.WriteString(`</div>`)
		}

		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		nameEscaped := html.EscapeString(name)
		switch strings.ToLower(name) {
		case "readme", "readme.md", "readme.txt":
			readme = nameEscaped
		}
		buff.WriteString(`<a href="`)
		buff.WriteString(nameEscaped)
		buff.WriteString(`">`)
		buff.WriteString(nameEscaped)
		buff.WriteString(`</a>`)
		buff.WriteString(``)
	}
	buff.WriteString(`</div>`)

	// Readme
	if readme != "" {
		buff.WriteString(`<pre id=r>`)
		buff.WriteString(readme)
		buff.WriteString(`</pre>`)
	}

	// Add script
	buff.Write(indexRaw2)

	return buff.Bytes()
}
