package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("router service ok"))
	})
	log.Println("Router service running on :8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
