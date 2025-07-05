package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Welcome to the Sting Ray API!")
    })

    http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"message": "Hello from the API!"})
    })

	// 0x6273 is 'bs' in hex
    fmt.Println("Server running on http://localhost:6273")
    http.ListenAndServe(":6273", nil)
} 