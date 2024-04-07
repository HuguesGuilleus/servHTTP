# Server HTTP

A basic server for develop or as demon.

# Configuration

File `/etc/servHTTP.toml` or the first args.

```toml
# A log directory.
log = "/var/log/servHTTP/"

# Each handlers have a type:
# - f => file
# - m => cache
# - r => redirect
# - s => secure (redirect to https)
# - p => reverse proxy
[mux.":443".h]
"example.org/" = { t = "r", u = "https://www.example.org/" }
"www.example.org/" = { t = "f", u = "www root...", c = "max-age=60" }
"www.example.org/assets/" = { t = "m", u = "www assets...", c = "max-age=3600, immutable" }
"www.example.org/api/" = { t = "p", u = "http://localhost:8000" },

# Define certificate directory and file.
[[mux.":443".cert]]
root = "/etc/lego/certificates/"
# The crt and key are added to the root.
crt = "example.org.crt"
key = "example.org.key"

# Define each handlers for each port
# Because no [[mux.":80".cert]], do no activate TLS.
[mux.":80".h]
"/" = { t = "s" }
"/.well-known/" = { t = "f", u = "/var/letsencrypt/" }
```
