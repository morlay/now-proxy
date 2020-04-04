package main

import (
	"fmt"
	"net/http"
	"os"

	handler "github.com/morlay/now-proxy/api"
)

type Handler func(writer http.ResponseWriter, request *http.Request)

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h(writer, request)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), Handler(handler.Handler))
	if err != nil {
		panic(err)
	}
}
