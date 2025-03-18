package api

import (
    "database/sql"
    "github.com/gorilla/mux"
    "net/http"
    "devops-assign/internal/service"
)

type Server struct {
    router  *mux.Router
    db      *sql.DB
    scanner *service.ScannerService
}

func NewServer(db *sql.DB) *Server {
    s := &Server{
        router:  mux.NewRouter(),
        db:      db,
        scanner: service.NewScannerService(db),
    }
    s.setupRoutes()
    return s
}

func (s *Server) Start(addr string) error {
    return http.ListenAndServe(addr, s.router)
}

func (s *Server) setupRoutes() {
    s.router.HandleFunc("/api/health", s.healthCheck).Methods("GET")
    s.router.HandleFunc("/api/scan", s.uploadScan).Methods("POST")
    s.router.HandleFunc("/api/vulnerabilities", s.getVulnerabilities).Methods("GET")
    s.router.HandleFunc("/api/vulnerabilities/{id}", s.getVulnerability).Methods("GET")
    s.router.HandleFunc("/api/stats", s.getStats).Methods("GET")
    s.router.HandleFunc("/api/scans", s.getScans).Methods("GET")
}