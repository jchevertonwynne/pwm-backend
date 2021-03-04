package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"pwm-backend/server"
)

func main() {
	go func() {
		http.ListenAndServe(":8081", nil)
	}()

	err := server.Run()
	if err != nil {
		fmt.Println(err)
	}
}
