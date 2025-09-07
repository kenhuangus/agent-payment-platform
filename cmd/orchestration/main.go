package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("orchestration service ok"))
	})
	log.Println("Orchestration service running on :8087")
	log.Fatal(http.ListenAndServe(":8087", nil))
}
