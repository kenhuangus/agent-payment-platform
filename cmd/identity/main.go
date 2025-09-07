package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("identity service ok"))
	})
	log.Println("Identity service running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
