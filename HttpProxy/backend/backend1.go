package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend 1 (API): You requested %s\n", r.URL.Path)
		log.Printf("backend1: Handle request: %s %s", r.Method, r.URL.Path)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Healthy\n")
		log.Printf("Handle request: %s %s", r.Method, r.URL.Path)
	})

	fmt.Println("Backend 1 listening on :8081")
	http.ListenAndServe(":8081", nil)
}
