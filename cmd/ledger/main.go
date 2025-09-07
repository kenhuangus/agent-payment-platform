package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ledger service ok"))
	})
	log.Println("Ledger service running on :8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}
