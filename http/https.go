package main

import (
	"net/http"
	"crypto/tls"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandler)
	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{

		},
	}
	server.ListenAndServe()
}
