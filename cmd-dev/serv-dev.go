package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/HuguesGuilleus/servHTTP/handlers"
)

func main() {
	l := flag.String("l", ":8000", "Listen address [ip]:port")
	flag.Usage = func() {
		fmt.Println("Usage of serv: $ serv [-l] ROOT")
		flag.PrintDefaults()
	}
	flag.Parse()
	root := flag.Arg(0)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("listen", "address", *l, "root", root)
	err := http.ListenAndServe(*l, handlers.File(logger, root, "no-store"))
	logger.Error("listen.fail", "err", err.Error())
	os.Exit(1)
}
