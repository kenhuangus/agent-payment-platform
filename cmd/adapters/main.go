package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("adapters service ok"))
	})
	log.Println("Adapters service running on :8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
}
