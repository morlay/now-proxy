package main

import (
	"fmt"
	"github.com/morlay/now-proxy/pkg/proxy"
	"github.com/morlay/now-proxy/pkg/version"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	fmt.Println(fmt.Sprintf("serve on 0.0.0.0:%s (%v)", port, version.FullVersion()))

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), &proxy.Proxy{})
	if err != nil {
		panic(err)
	}
}
