package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	myCongig, err := ReadFile("./example.toml")
	assert.NoError(t, err)
	assert.Equal(t, &Config{
		Log: "/var/log/servHTTP/",
		Mux: map[string]Mux{
			":80": {
				Handlers: map[string]Handler{
					"/":             {Type: "s"},
					"/.well-known/": {Type: "f", URL: "/var/letsencrypt/"},
				},
			},
			":443": {
				Cert: []Cert{
					{Root: "/etc/lego/certificates/", Key: "example.com.key", Crt: "example.com.crt"},
					{Root: "/etc/lego/certificates/", Key: "example.org.key", Crt: "example.org.crt"},
				},
				Handlers: map[string]Handler{
					"example.org/": {Type: "f", URL: "/var/www/example.org/", Cache: "max-age=60"},
					"example.com/": {Type: "f", URL: "/var/www/example.com/", Cache: "max-age=60"},
				},
			},
		},
	}, myCongig)
}
