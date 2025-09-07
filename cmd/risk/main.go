package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("risk service ok"))
	})
	log.Println("Risk service running on :8086")
	log.Fatal(http.ListenAndServe(":8086", nil))
}
