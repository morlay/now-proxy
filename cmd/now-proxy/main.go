package main

import (
	"fmt"
	"github.com/morlay/now-proxy/pkg/proxy"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), &proxy.Proxy{})
	if err != nil {
		panic(err)
	}
}
