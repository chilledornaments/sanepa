package main

import "net/http"

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive"))
	return
}

func scaleEndpoint(w http.ResponseWriter, r *http.Request) {

}
