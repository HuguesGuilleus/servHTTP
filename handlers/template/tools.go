package template

import (
	"bytes"
	"html"
	"strings"
)

func htmlSize(buff *bytes.Buffer, size int64) {
	thousand := int64(1)
	for j := size; j > 1000; j /= 1000 {
		thousand *= 1000
	}

	if true {
		i := (size / thousand) % 1000
		thousand /= 1000
		if 100 < i {
			buff.WriteByte(byte(i/100) + '0')
		}
		if 10 < i {
			buff.WriteByte(byte(i/10%10) + '0')
		}
		buff.WriteByte(byte(i%10) + '0')
	}

	for ; thousand > 0; thousand /= 1000 {
		i := (size / thousand) % 1000
		buff.WriteByte(' ')
		buff.WriteByte(byte(i/100) + '0')
		buff.WriteByte(byte(i/10%10) + '0')
		buff.WriteByte(byte(i%10) + '0')
	}

	buff.WriteString(" B")
}

func htmlPath(buff *bytes.Buffer, path string) {
	buff.WriteString(`<div id=p><a href="/">/</a>`)
	defer buff.WriteString(`</div>`)

	splits := strings.Split(strings.TrimPrefix(path, "/"), "/")
	end := splits[len(splits)-1]
	splits = splits[:len(splits)-1]
	all := bytes.NewBufferString("/")
	for _, p := range splits {
		h := html.EscapeString(p)
		all.WriteString(h)
		all.WriteByte('/')
		buff.WriteString(`<a href="`)
		buff.WriteString(all.String())
		buff.WriteString(`">`)
		buff.WriteString(h)
		buff.WriteString(`/</a>`)
	}

	if end != "" {
		buff.WriteString(`<a href="">`)
		buff.WriteString(html.EscapeString(end))
		buff.WriteString(`</a>`)
	}
}
