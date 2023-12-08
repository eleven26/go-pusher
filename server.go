package main

import (
	"net/http"
	"time"
)

func mux(hub *Hub) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		send(hub, w, r)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics(hub, w)
	})
	return mux
}

func server(addr string) error {
	hub := newHub()
	go hub.run()

	s := &http.Server{
		Addr:              addr,
		Handler:           mux(hub),
		ReadHeaderTimeout: 3 * time.Second,
	}
	return s.ListenAndServe()
}
