package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend 2 (API): You requested %s\n", r.URL.Path)
		log.Printf("backend2: Handle request: %s %s", r.Method, r.URL.Path)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Healthy\n")
		log.Printf("Handle request: %s %s", r.Method, r.URL.Path)
	})

	fmt.Println("Backend 2 listening on :8082")
	http.ListenAndServe(":8082", nil)
}
