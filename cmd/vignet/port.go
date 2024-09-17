package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httputils"
)

func newPort(clients clients, services services) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("HEAD /", services.vignet.HandleHead)
	mux.HandleFunc("GET /", services.vignet.HandleGet)
	mux.HandleFunc("POST /", services.vignet.HandlePost)
	mux.HandleFunc("PUT /", services.vignet.HandlePut)
	mux.HandleFunc("PATCH /", services.vignet.HandlePatch)
	mux.HandleFunc("DELETE /", services.vignet.HandleDelete)

	return httputils.Handler(mux, clients.health,
		clients.telemetry.Middleware("http"),
	)
}
