package config

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/HuguesGuilleus/go-logoutput"
	"github.com/HuguesGuilleus/servHTTP/handlers"
)

// Standard demon main.
// Usage: init your custom handlers, then just call this in your main func.
func Main() {
	flag.Usage = func() {
		os.Stderr.WriteString("Usage: $ serv [/etc/servHTTP.toml]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	configFile := "/etc/servHTTP.toml"
	if flag.Arg(0) != "" {
		configFile = flag.Arg(0)
	}
	Listen(configFile)
	os.Exit(1)
}

// Map of config handlers.
// The key is used in the config file.
// The value is used when initiate the server to create the handlers.
//
// You can add your cutom handler to this map at init.
var Handlers = map[string]func(logger *slog.Logger, u, cacheControle string) http.Handler{
	"f": handlers.File,
	"m": handlers.Cache,
	"r": handlers.Redirect,
	"s": handlers.Secure,
	"p": handlers.ReverseProxy,
}

type Config struct {
	// Log output directory
	Log string
	// Multiplexors, the key is the listened address `[IP]:port`
	Mux map[string]Mux
}

type Mux struct {
	// TLS certificates position.
	// Is any, do no use TLS.
	Cert []Cert
	// Handlers config, indexed by domain and path
	Handlers map[string]Handler `toml:"h"`
}

type Cert struct {
	Root string
	Key  string
	Crt  string
}

type Handler struct {
	// Type of the handler.
	// f: file
	// m: file with memory cache
	// r: redirect
	// s: redirect to HTTPS, ignore .U field
	// Future version
	// p: reverse proxy
	Type string `toml:"t"`
	// A URL, for file root, URL for redirect or reverse serv...
	URL string `toml:"u"`
	// Cache control instruction.
	Cache string `toml:"c"`
}

// Decode toml file into a Config structure.
func ReadFile(path string) (*Config, error) {
	config := new(Config)
	if _, err := toml.DecodeFile(path, config); err != nil {
		return nil, fmt.Errorf("decode file %q: %w", path, err)
	}
	return config, nil
}

// Liten on all multiplexer.
// Return if the load of config file fail or all multiplexer listen fail.
func Listen(configFile string) {
	config, err := ReadFile(configFile)
	if err != nil {
		slog.Error("init-fail", "err", err.Error())
		return
	}

	logger := slog.New(slog.NewJSONHandler(logoutput.New(config.Log), nil))

	wg := sync.WaitGroup{}
	wg.Add(len(config.Mux))
	defer wg.Wait()
	for address, mux := range config.Mux {
		go func(address string, mux Mux) {
			defer wg.Done()
			err := mux.Listen(logger, address).Error()
			logger.Error("init", "address", address, "err", err)
		}(address, mux)
	}
}

func (mux *Mux) Listen(logger *slog.Logger, address string) error {
	muxServer := http.NewServeMux()
	for pattern, config := range mux.Handlers {
		n := Handlers[config.Type]
		if n == nil {
			return fmt.Errorf("unknown handler type: %q", config.Type)
		}
		muxServer.Handle(pattern, n(logger, config.URL, config.Cache))
	}

	logger.Info("listen", "address", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	if config, err := mux.LoadTLS(); err != nil {
		return err
	} else if config != nil {
		listener = tls.NewListener(listener, config)
	}

	return (&http.Server{
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
		Handler:  muxServer,
	}).Serve(listener)
}

func (mux Mux) LoadTLS() (*tls.Config, error) {
	if len(mux.Cert) == 0 {
		return nil, nil
	}

	config := new(tls.Config)
	config.NextProtos = []string{"h2"}

	for _, c := range mux.Cert {
		certificate, err := tls.LoadX509KeyPair(filepath.Join(c.Root, c.Crt), filepath.Join(c.Root, c.Key))
		if err != nil {
			return nil, err
		}
		config.Certificates = append(config.Certificates, certificate)
	}

	return config, nil
}
