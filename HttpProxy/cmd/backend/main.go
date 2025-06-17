package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {

	numOfBackends := 4
	wg := sync.WaitGroup{}
	wg.Add(numOfBackends)

	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second

	for i := 0; i < numOfBackends; i++ {
		go func(i int) {
			defer wg.Done()

			port := fmt.Sprintf(":%d", 8081+i)

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				proxyHandler(w, r, port)
			})

			server := &http.Server{
				Addr:         port,
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
				Handler:      mux,
			}

			fmt.Printf("Backend %d listening on port %s\n", i, port)
			server.ListenAndServe()
		}(i)
	}
	wg.Wait()
}

func proxyHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Fprintf(w, "Backend (API) on port %s: You requested %s\n", port, r.URL.Path)
	log.Printf("backend on port %s: Handle request: %s %s", port, r.Method, r.URL.Path)
}
