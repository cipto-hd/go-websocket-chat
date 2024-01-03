package main

import (
	"log"
	"net/http"
)

func main() {
	mux := routes()

	log.Println("Starting web server in port 8080")

	http.Handle("/", mux)
	_ = http.ListenAndServe(":8080", nil)
}
