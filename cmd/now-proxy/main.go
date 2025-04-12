package main

import (
	"fmt"
	"github.com/morlay/now-proxy/cmd/now-proxy/internal/version"
	"log/slog"
	"net/http"
	"os"

	"github.com/morlay/now-proxy/pkg/proxy"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)

	slog.Default().
		With(
			slog.String("service.version", version.FullVersion()),
			slog.String("listen.addr", addr),
		).
		Info(fmt.Sprintf("serving"))

	err := http.ListenAndServe(addr, &proxy.Proxy{})
	if err != nil {
		panic(err)
	}
}
