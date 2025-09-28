package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from target server", "timestamp": "%s", "path": "%s"}`,
			time.Now().Format(time.RFC3339), r.URL.Path)
	})

	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Slow endpoint", "delay": "100ms"}`)
	})

	http.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": "Something went wrong"}`, http.StatusInternalServerError)
	})

	log.Println("Test server starting on :3000")
	log.Println("Endpoints:")
	log.Println("  GET /        - Returns basic info")
	log.Println("  GET /slow    - Slow endpoint (100ms delay)")
	log.Println("  GET /error   - Returns 500 error")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Test server failed: %v", err)
	}
}
