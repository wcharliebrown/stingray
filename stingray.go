package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)



func main() {
    // Initialize database
    if err := InitDatabase(); err != nil {
        fmt.Printf("Failed to initialize database: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()
    
    // Create a channel to receive shutdown signals
    shutdownChan := make(chan os.Signal, 1)
    signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

    // Create a server instance for graceful shutdown
    server := &http.Server{
        Addr: ":6273",
    }

    // Shutdown endpoint
    http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        // Trigger graceful shutdown
        go func() {
            time.Sleep(100 * time.Millisecond) // Give response time to be sent
            shutdownChan <- syscall.SIGTERM
        }()
        
        // Use HandlePageRequest for the response
        HandlePageRequest(w, r, "shutdown")
    })

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        // Serve home page from database
        HandlePageRequest(w, r, "home")
    })

    // GET /user/login
    http.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        // Use HandlePageRequest for the response
        HandlePageRequest(w, r, "login")
    })

    // POST /user/login
    http.HandleFunc("/user/login_post", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]string{"info": "POST user login"})
        } else {
            // Create a login result page using template
            resultPage := &Page{
                Title:          "Login Result",
                MetaDescription: "Login submission result",
                Header:         "<h1>Login Submitted</h1>",
                Navigation:     `<nav><a href="/">Home</a></nav>`,
                MainContent:    "<p>Login form submitted successfully.</p><p><a href='/'>Back to Home</a></p>",
                Sidebar:        "<div class='sidebar'><h3>Next Steps</h3><p>You will be redirected shortly.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "login-result-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", resultPage)
        }
    })

    // GET /page/{slug} - Dynamic page serving using database
    http.HandleFunc("/page/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        // Extract slug from URL path
        slug := r.URL.Path[len("/page/"):]
        if slug == "" {
            http.Error(w, "Page slug required", http.StatusBadRequest)
            return
        }
        
        HandlePageRequest(w, r, slug)
    })

    // GET /pages - Get all available pages
    http.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(GetAllPages())
        }
    })

    // GET /templates - Get all available templates
    http.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            templates := getAvailableTemplates()
            json.NewEncoder(w).Encode(map[string]interface{}{
                "templates": templates,
                "count": len(templates),
            })
        }
    })

    // GET /template/{name} - Get template by name
    http.HandleFunc("/template/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        // Extract template name from URL path
        templateName := r.URL.Path[len("/template/"):]
        if templateName == "" {
            http.Error(w, "Template name required", http.StatusBadRequest)
            return
        }
        
        // Check if template exists
        if !templateExists(templateName) {
            http.Error(w, "Template not found", http.StatusNotFound)
            return
        }
        
        // Load template content
        html, err := loadTemplateFromFile(templateName)
        if err != nil {
            http.Error(w, "Template not found", http.StatusNotFound)
            return
        }
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "name": templateName,
                "html": html,
            })
        }
    })

    // Start server in a goroutine
    go func() {
        // 0x6273 is 'bs' in hex
        fmt.Println("Server running on http://localhost:6273")
        fmt.Println("To stop the server, send POST request to http://localhost:6273/shutdown")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("Server error: %v\n", err)
        }
    }()

    // Wait for shutdown signal
    <-shutdownChan
    fmt.Println("\nShutting down server gracefully...")
    
    // Create context with timeout for graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        fmt.Printf("Server shutdown error: %v\n", err)
    } else {
        fmt.Println("Server stopped gracefully")
    }
} 