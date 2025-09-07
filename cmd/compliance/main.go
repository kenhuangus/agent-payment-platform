package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("compliance service ok"))
	})
	log.Println("Compliance service running on :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
