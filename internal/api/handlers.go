package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "github.com/gorilla/mux"
    "devops-assign/internal/models"
)

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
    err := s.db.Ping()
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Database connection failed")
        return
    }
    respondWithJSON(w, http.StatusOK, map[string]string{
        "status":   "healthy",
        "database": "connected",
    })
}

func (s *Server) uploadScan(w http.ResponseWriter, r *http.Request) {
    var scanResults []models.ScanResult
    if err := json.NewDecoder(r.Body).Decode(&scanResults); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    err := s.scanner.ProcessScanResults(scanResults)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusCreated, map[string]string{
        "message": "Scan results processed successfully",
    })
}

func (s *Server) getVulnerabilities(w http.ResponseWriter, r *http.Request) {
    filters := models.VulnerabilityFilters{
        Severity:    r.URL.Query().Get("severity"),
        Status:      r.URL.Query().Get("status"),
        PackageName: r.URL.Query().Get("package"),
    }

    vulns, err := s.scanner.GetVulnerabilities(filters)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, vulns)
}

func (s *Server) getVulnerability(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    vuln, err := s.scanner.GetVulnerabilityByID(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Vulnerability not found")
        return
    }

    respondWithJSON(w, http.StatusOK, vuln)
}

func (s *Server) getStats(w http.ResponseWriter, r *http.Request) {
    stats, err := s.scanner.GetStatistics()
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, stats)
}

func (s *Server) getScans(w http.ResponseWriter, r *http.Request) {
    limit := 10
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil {
            limit = l
        }
    }

    scans, err := s.scanner.GetRecentScans(limit)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, scans)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}