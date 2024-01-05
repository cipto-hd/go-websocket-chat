package main

import (
	"log"
	"net/http"

	"go-websocket-chat/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting web server in port 8080")

	http.Handle("/", mux)
	_ = http.ListenAndServe(":8080", nil)
}
