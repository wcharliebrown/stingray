package main

import (
	"context"
	"log"
	"net/http"
	"stingray/database"
	"stingray/handlers"
)

type Server struct {
	db          *database.Database
	server      *http.Server
	pageHandler *handlers.PageHandler
	authHandler *handlers.AuthHandler
}

func NewServer(db *database.Database) *Server {
	mux := http.NewServeMux()
	server := &Server{
		db:          db,
		pageHandler: handlers.NewPageHandler(db),
		authHandler: handlers.NewAuthHandler(db),
	}

	// Page routes
	mux.HandleFunc("/", server.pageHandler.HandleHome)
	mux.HandleFunc("/page/", server.pageHandler.HandlePage)
	mux.HandleFunc("/pages", server.pageHandler.HandlePages)
	mux.HandleFunc("/templates", server.pageHandler.HandleTemplates)
	mux.HandleFunc("/template/", server.pageHandler.HandleTemplate)
	
	// Auth routes
	mux.HandleFunc("/user/login", server.authHandler.HandleLogin)
	mux.HandleFunc("/user/login_post", server.authHandler.HandleLoginPost)

	server.server = &http.Server{
		Addr:    ":6273",
		Handler: mux,
	}

	return server
}

func (s *Server) Start() error {
	log.Printf("Starting Sting Ray server on port 6273...")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server gracefully...")
	return s.server.Shutdown(ctx)
} 