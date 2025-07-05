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

    // GET /user/login
    http.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "info": "GET user login page",
            "form": map[string]interface{}{
                "elements": []map[string]interface{}{
                    {
                        "name": "username",
                        "type": "text",
                        "required": true,
                        "title": "Enter your username",
                        "placeholder": "Username",
                    },
                    {
                        "name": "password", 
                        "type": "password",
                        "required": true,
                        "title": "Enter your password",
                        "placeholder": "Password",
                    },
                },
            },
        })
    })

    // POST /user/login
    http.HandleFunc("/user/login_post", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        // Example: parse JSON body (not implemented, just echoing for now)
        json.NewEncoder(w).Encode(map[string]string{"info": "POST user login"})
    })

    // GET /page/about
    http.HandleFunc("/page/about", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"about": "This is the about page."})
    })

    // GET /table/test123
    http.HandleFunc("/table/test123", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "table": "test123",
            "rows": []map[string]interface{}{
                {"id": 1, "value": "row1"},
                {"id": 2, "value": "row2"},
            },
        })
    })

	// 0x6273 is 'bs' in hex
    fmt.Println("Server running on http://localhost:6273")
    http.ListenAndServe(":6273", nil)
} 