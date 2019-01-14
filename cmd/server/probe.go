package main

import (
	"net/http"
)

func probe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newProbeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/probe", probe)
	mux.HandleFunc("/metrics", probe)

	return mux
}
