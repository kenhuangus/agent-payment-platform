package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("consent service ok"))
	})
	log.Println("Consent service running on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
