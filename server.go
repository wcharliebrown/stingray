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
	sessionMW   *handlers.SessionMiddleware
}

func NewServer(db *database.Database) *Server {
	mux := http.NewServeMux()
	sessionMW := handlers.NewSessionMiddleware(db)
	
	server := &Server{
		db:          db,
		pageHandler: handlers.NewPageHandler(db),
		authHandler: handlers.NewAuthHandler(db),
		sessionMW:   sessionMW,
	}

	// Page routes with optional auth middleware
	mux.HandleFunc("/", sessionMW.OptionalAuth(server.pageHandler.HandleHome))
	mux.HandleFunc("/page/", sessionMW.OptionalAuth(server.pageHandler.HandlePage))
	mux.HandleFunc("/pages", sessionMW.OptionalAuth(server.pageHandler.HandlePages))
	mux.HandleFunc("/templates", sessionMW.OptionalAuth(server.pageHandler.HandleTemplate))
	mux.HandleFunc("/template/", sessionMW.OptionalAuth(server.pageHandler.HandleTemplate))
	
	// Auth routes
	mux.HandleFunc("/user/login", server.authHandler.HandleLogin)
	mux.HandleFunc("/user/login_post", server.authHandler.HandleLoginPost)
	mux.HandleFunc("/user/logout", server.authHandler.HandleLogout)
	mux.HandleFunc("/user/profile", sessionMW.RequireAuth(server.authHandler.HandleProfile))

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