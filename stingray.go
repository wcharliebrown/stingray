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
    htmltemplate "html/template"
)



func main() {
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
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]string{
                "message": "Shutdown initiated",
                "status": "shutting_down",
            })
        } else {
            // Create a simple page for shutdown
            shutdownPage := &Page{
                Title:          "Shutdown",
                MetaDescription: "Server shutdown page",
                Header:         "<h1>Shutdown Initiated</h1>",
                Navigation:     `<nav><a href="/">Home</a></nav>`,
                MainContent:    "<p>The server is shutting down gracefully.</p><p>Status: <strong>shutting_down</strong></p>",
                Sidebar:        "<div class='sidebar'><h3>Info</h3><p>This page will close shortly.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "shutdown-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", shutdownPage)
        }
        
        // Trigger graceful shutdown
        go func() {
            time.Sleep(100 * time.Millisecond) // Give response time to be sent
            shutdownChan <- syscall.SIGTERM
        }()
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
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
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
        } else {
            // Create a login page using template
            loginPage := &Page{
                Title:          "User Login",
                MetaDescription: "Login to your account",
                Header:         "<h1>User Login</h1>",
                Navigation:     `<nav><a href="/">Home</a></nav>`,
                MainContent:    `<form method="post" action="/user/login_post">
    <div class="form-group">
        <label for="username">Username:</label>
        <input type="text" id="username" name="username" required>
    </div>
    <div class="form-group">
        <label for="password">Password:</label>
        <input type="password" id="password" name="password" required>
    </div>
    <button type="submit" class="btn">Login</button>
</form>`,
                Sidebar:        "<div class='sidebar'><h3>Need Help?</h3><p>Contact support if you're having trouble logging in.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "login-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", loginPage)
        }
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

    // GET /table/test123
    http.HandleFunc("/table/test123", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "table": "test123",
                "rows": []map[string]interface{}{
                    {"id": 1, "value": "row1"},
                    {"id": 2, "value": "row2"},
                },
            })
        } else {
            // Create a table page using template
            tablePage := &Page{
                Title:          "Table Test123",
                MetaDescription: "Test table data",
                Header:         "<h1>Table: test123</h1>",
                Navigation:     `<nav><a href="/">Home</a></nav>`,
                MainContent:    `<table>
    <thead>
        <tr><th>ID</th><th>Value</th></tr>
    </thead>
    <tbody>
        <tr><td>1</td><td>row1</td></tr>
        <tr><td>2</td><td>row2</td></tr>
    </tbody>
</table>
<p><a href="/">Back to Home</a></p>`,
                Sidebar:        "<div class='sidebar'><h3>Table Info</h3><p>This is a test table with sample data.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "table-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", tablePage)
        }
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
        } else {
            // Create a pages list using template
            pages := GetAllPages()
            pagesList := "<ul>"
            for slug, page := range pages {
                pagesList += fmt.Sprintf(`<li><a href="/page/%s">%s</a> - %s</li>`, slug, page.Title, page.MetaDescription)
            }
            pagesList += "</ul>"
            
            pagesPage := &Page{
                Title:          "All Pages",
                MetaDescription: "List of all available pages",
                Header:         "<h1>All Available Pages</h1>",
                Navigation:     `<nav><a href="/">Home</a></nav>`,
                MainContent:    htmltemplate.HTML(pagesList + "<p><a href='/'>Back to Home</a></p>"),
                Sidebar:        "<div class='sidebar'><h3>Navigation</h3><p>Browse all available pages in the application.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "pages-list-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", pagesPage)
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
            json.NewEncoder(w).Encode(GetAllTemplates())
        } else {
            // Create a templates list using template
            templates := GetAllTemplates()
            templatesList := "<ul>"
            for id, template := range templates {
                templatesList += fmt.Sprintf(`<li><a href="/template/%d">%s</a> (ID: %d)</li>`, id, template.Name, template.ID)
            }
            templatesList += "</ul>"
            
            templatesPage := &Page{
                Title:          "All Templates",
                MetaDescription: "List of all available templates",
                Header:         "<h1>All Available Templates</h1>",
                Navigation:     `<nav><a href="/">Home</a></nav>`,
                MainContent:    htmltemplate.HTML(templatesList + "<p><a href='/'>Back to Home</a></p>"),
                Sidebar:        "<div class='sidebar'><h3>Templates</h3><p>Browse all available templates in the application.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "templates-list-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", templatesPage)
        }
    })

    // GET /template/{id} - Get template by ID
    http.HandleFunc("/template/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        // Extract ID from URL path
        idStr := r.URL.Path[len("/template/"):]
        if idStr == "" {
            http.Error(w, "Template ID required", http.StatusBadRequest)
            return
        }
        
        // Parse ID as integer
        var id int
        _, err := fmt.Sscanf(idStr, "%d", &id)
        if err != nil {
            http.Error(w, "Invalid template ID", http.StatusBadRequest)
            return
        }
        
        template, exists := GetTemplateByID(id)
        if !exists {
            http.Error(w, "Template not found", http.StatusNotFound)
            return
        }
        
        responseFormat := getResponseFormat(r)
        if responseFormat == "json" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(template)
        } else {
            // Create a template detail page using template
            templatePage := &Page{
                Title:          "Template: " + template.Name,
                MetaDescription: "Template details and HTML content",
                Header:         htmltemplate.HTML("<h1>Template: " + template.Name + "</h1>"),
                Navigation:     `<nav><a href="/">Home</a> | <a href="/templates">Templates</a></nav>`,
                MainContent:    htmltemplate.HTML(fmt.Sprintf(`<p><strong>ID:</strong> %d</p>
<p><strong>Name:</strong> %s</p>
<h2>HTML Content:</h2>
<pre style="background: #f5f5f5; padding: 15px; overflow-x: auto;">%s</pre>
<p><a href="/templates">Back to Templates</a> | <a href="/">Back to Home</a></p>`, template.ID, template.Name, template.HTML)),
                Sidebar:        "<div class='sidebar'><h3>Template Info</h3><p>View template details and HTML content.</p></div>",
                Footer:         "<footer>&copy; 2024 Sting Ray</footer>",
                CSSClass:       "template-detail-page",
                Scripts:        "",
            }
            renderHTMLWithTemplate(w, "simple", templatePage)
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