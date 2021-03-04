package server

import (
	"fmt"
	"net/http"
)

func Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)

	return http.ListenAndServe(":8080", mux)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong"))
	if err != nil {
		fmt.Println(err)
	}
}