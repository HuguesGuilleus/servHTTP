package template

import (
	"bytes"
	"testing"
)

func TestHtmlSize(t *testing.T) {
	wrap := func(size int64, expected string) {
		buff := bytes.Buffer{}
		htmlSize(&buff, size)
		assertString(t, expected, buff.String())
	}

	wrap(0, "0 B")
	wrap(1, "1 B")
	wrap(12, "12 B")
	wrap(123, "123 B")
	wrap(6_789_123, "6 789 123 B")
	wrap(56_789_123, "56 789 123 B")
	wrap(456_789_123, "456 789 123 B")
}

func TestHtmlPath(t *testing.T) {
	wrap := func(expected, path string) {
		buff := bytes.Buffer{}
		htmlPath(&buff, path)
		assertString(t, expected, buff.String())
	}

	wrap(`<div id=p><a href="/">/</a></div>`, "/")

	wrap(`<div id=p>`+
		`<a href="/">/</a>`+
		`<a href="">file.txt</a>`+
		`</div>`,
		"/file.txt")

	wrap(`<div id=p>`+
		`<a href="/">/</a>`+
		`<a href="/d&amp;r/">d&amp;r/</a>`+
		`</div>`,
		"/d&r/")

	wrap(`<div id=p>`+
		`<a href="/">/</a>`+
		`<a href="/d&amp;r/">d&amp;r/</a>`+
		`<a href="/d&amp;r/subdir/">subdir/</a>`+
		`</div>`,
		"/d&r/subdir/")

	wrap(`<div id=p>`+
		`<a href="/">/</a>`+
		`<a href="/d&amp;r/">d&amp;r/</a>`+
		`<a href="/d&amp;r/subdir/">subdir/</a>`+
		`<a href="">file-&gt;.txt</a>`+
		`</div>`,
		"/d&r/subdir/file->.txt")
}
