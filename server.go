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
	roleMW      *handlers.RoleMiddleware
	apiHandler  *handlers.APIHandler
	metadataHandler *handlers.MetadataHandler
}

func NewServer(db *database.Database) *Server {
	mux := http.NewServeMux()
	sessionMW := handlers.NewSessionMiddleware(db)
	roleMW := handlers.NewRoleMiddleware(db)
	apiHandler := handlers.NewAPIHandler(db)
	
	server := &Server{
		db:          db,
		pageHandler: handlers.NewPageHandler(db),
		authHandler: handlers.NewAuthHandler(db),
		sessionMW:   sessionMW,
		roleMW:      roleMW,
		apiHandler:  apiHandler,
		metadataHandler: handlers.NewMetadataHandler(db),
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

	// Role-based page routes
	mux.HandleFunc("/page/orders", roleMW.RequireAdmin(server.pageHandler.HandlePage))
	mux.HandleFunc("/page/faq", roleMW.RequireCustomer(server.pageHandler.HandlePage))

	// API routes
	mux.HandleFunc("/api/users", apiHandler.HandleGetUsers)
	mux.HandleFunc("/api/groups", apiHandler.HandleGetGroups)
	mux.HandleFunc("/api/user-groups", apiHandler.HandleGetUserGroups)
	mux.HandleFunc("/api/current-user", apiHandler.HandleGetCurrentUser)

	// Metadata routes
	mux.HandleFunc("/metadata/tables", server.metadataHandler.HandleTableList)
	mux.HandleFunc("/metadata/table/", server.metadataHandler.HandleTableData)
	mux.HandleFunc("/metadata/edit/", sessionMW.RequireAuth(server.metadataHandler.HandleEditRow))
	mux.HandleFunc("/metadata/delete/", sessionMW.RequireAuth(server.metadataHandler.HandleDeleteRow))
	mux.HandleFunc("/metadata/edit-table/", sessionMW.RequireAuth(server.metadataHandler.HandleEditTableMetadata))
	mux.HandleFunc("/metadata/delete-table/", sessionMW.RequireAuth(server.metadataHandler.HandleDeleteTable))
	mux.HandleFunc("/metadata/create-table", sessionMW.RequireAuth(server.metadataHandler.HandleCreateTable))

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