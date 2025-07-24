package main

import (
	"context"
	"log"
	"net/http"
	"stingray/config"
	"stingray/database"
	"stingray/handlers"
	"stingray/logging"
)

type Server struct {
	db          *database.Database
	cfg         *config.Config
	logger      *logging.Logger
	server      *http.Server
	pageHandler *handlers.PageHandler
	authHandler *handlers.AuthHandler
	sessionMW   *handlers.SessionMiddleware
	roleMW      *handlers.RoleMiddleware
	loggingMW   *handlers.LoggingMiddleware
	apiHandler  *handlers.APIHandler
	metadataHandler *handlers.MetadataHandler
	passwordResetHandler *handlers.PasswordResetHandler
}

func NewServer(db *database.Database, cfg *config.Config) *Server {
	mux := http.NewServeMux()
	logger := logging.NewLogger(cfg.LoggingLevel)
	sessionMW := handlers.NewSessionMiddleware(db)
	roleMW := handlers.NewRoleMiddleware(db)
	loggingMW := handlers.NewLoggingMiddleware(logger)
	apiHandler := handlers.NewAPIHandler(db, cfg)
	
	server := &Server{
		db:          db,
		cfg:         cfg,
		logger:      logger,
		pageHandler: handlers.NewPageHandler(db, cfg), // Pass cfg
		authHandler: handlers.NewAuthHandler(db, logger),
		sessionMW:   sessionMW,
		roleMW:      roleMW,
		loggingMW:   loggingMW,
		apiHandler:  apiHandler,
		metadataHandler: handlers.NewMetadataHandler(db),
		passwordResetHandler: handlers.NewPasswordResetHandler(db, cfg, logger),
	}

	// Page routes with optional auth middleware
	mux.HandleFunc("/", loggingMW.Wrap(sessionMW.OptionalAuth(server.pageHandler.HandleHome)))
	mux.HandleFunc("/page/", loggingMW.Wrap(sessionMW.OptionalAuth(server.pageHandler.HandlePage)))
	mux.HandleFunc("/pages", loggingMW.Wrap(sessionMW.OptionalAuth(server.pageHandler.HandlePages)))
	mux.HandleFunc("/templates", loggingMW.Wrap(sessionMW.OptionalAuth(server.pageHandler.HandleTemplate)))
	mux.HandleFunc("/template/", loggingMW.Wrap(sessionMW.OptionalAuth(server.pageHandler.HandleTemplate)))

	// Config settings page (admin or engineer only)
	mux.HandleFunc("/config", loggingMW.Wrap(func(w http.ResponseWriter, r *http.Request) {
		if !sessionMW.IsAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		session, err := sessionMW.GetSessionFromRequest(r)
		if err != nil {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		isAdmin, _ := db.IsUserInGroup(session.UserID, "admin")
		isEngineer, _ := db.IsUserInGroup(session.UserID, "engineer")
		if !isAdmin && !isEngineer {
			handlers.RenderMessage(w, "Access Denied", "Access Denied", "error", "You do not have permission to access this page.", "/", "Go Home", http.StatusForbidden)
			return
		}
		if r.Method == "POST" {
			log.Printf("[DEBUG] HandleConfigPage POST called")
			server.pageHandler.HandleConfigPage(w, r)
		} else {
			server.pageHandler.HandleConfigPage(w, r)
		}
	}))
	
	// Auth routes
	mux.HandleFunc("/user/login", loggingMW.Wrap(server.authHandler.HandleLogin))
	mux.HandleFunc("/user/login_post", loggingMW.Wrap(server.authHandler.HandleLoginPost))
	mux.HandleFunc("/user/logout", loggingMW.Wrap(server.authHandler.HandleLogout))
	mux.HandleFunc("/user/profile", loggingMW.Wrap(sessionMW.RequireAuth(server.authHandler.HandleProfile)))

	// Password reset routes
	mux.HandleFunc("/user/password-reset-request", loggingMW.Wrap(server.passwordResetHandler.HandlePasswordResetRequest))
	mux.HandleFunc("/user/password-reset-confirm", loggingMW.Wrap(server.passwordResetHandler.HandlePasswordResetConfirm))

	// Role-based page routes
	mux.HandleFunc("/page/orders", loggingMW.Wrap(roleMW.RequireAdmin(server.pageHandler.HandlePage)))
	mux.HandleFunc("/page/faq", loggingMW.Wrap(roleMW.RequireCustomer(server.pageHandler.HandlePage)))

	// API routes
	mux.HandleFunc("/api/users", loggingMW.Wrap(apiHandler.HandleGetUsers))
	mux.HandleFunc("/api/groups", loggingMW.Wrap(apiHandler.HandleGetGroups))
	mux.HandleFunc("/api/user-groups", loggingMW.Wrap(apiHandler.HandleGetUserGroups))
	mux.HandleFunc("/api/current-user", loggingMW.Wrap(apiHandler.HandleGetCurrentUser))

	// Metadata routes
	mux.HandleFunc("/metadata/tables", loggingMW.Wrap(server.metadataHandler.HandleTableList))
	mux.HandleFunc("/metadata/table/", loggingMW.Wrap(server.metadataHandler.HandleTableData))
	mux.HandleFunc("/metadata/edit/", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleEditRow)))
	mux.HandleFunc("/metadata/delete/", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleDeleteRow)))
	mux.HandleFunc("/metadata/edit-table/", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleEditTableMetadata)))
	mux.HandleFunc("/metadata/delete-table/", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleDeleteTable)))
	mux.HandleFunc("/metadata/create-table", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleCreateTable)))

	// Field metadata API routes
	mux.HandleFunc("/api/metadata/field", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleFieldMetadata)))
	mux.HandleFunc("/api/metadata/field/", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleFieldMetadata)))

	// Register the new /api/reload route
	mux.HandleFunc("/api/reload", loggingMW.Wrap(sessionMW.RequireAuth(apiHandler.HandleReloadEnv)))

	server.server = &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	return server
}

func (s *Server) Start() error {
	s.logger.LogVerbose("Starting Sting Ray server on port " + s.cfg.ServerPort + "...")
	log.Printf("Starting Sting Ray server on port %s...", s.cfg.ServerPort)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.LogVerbose("Shutting down server gracefully...")
	log.Println("Shutting down server gracefully...")
	return s.server.Shutdown(ctx)
} 