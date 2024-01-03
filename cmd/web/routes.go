package main

import (
	"net/http"

	"github.com/bmizerany/pat"

	"go-websocket-chat/internal/handlers"
)

func routes() http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	mux.Get("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return mux
}
