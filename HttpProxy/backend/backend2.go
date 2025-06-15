package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend 2 (API): You requested %s\n", r.URL.Path)
		log.Printf("Handle request: %s %s", r.Method, r.URL.Path)
		log.Println("Oh yeah I am backend 2")
	})
	fmt.Println("Backend 2 listening on :8082")
	http.ListenAndServe(":8082", nil)
}
