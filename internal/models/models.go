package models

import (
    "time"
)

type ScanResult struct {
    ScanID        string          `json:"scan_id"`
    Timestamp     time.Time       `json:"timestamp"`
    ScanStatus    string          `json:"scan_status"`
    ResourceType  string          `json:"resource_type"`
    ResourceName  string          `json:"resource_name"`
    Vulnerabilities []Vulnerability `json:"vulnerabilities"`
    Summary       Summary         `json:"summary"`
}

type Vulnerability struct {
    ID            string    `json:"id"`
    Severity      string    `json:"severity"`
    CVSS         float64   `json:"cvss"`
    Status       string    `json:"status"`
    PackageName  string    `json:"package_name"`
    CurrVersion  string    `json:"current_version"`
    FixedVersion string    `json:"fixed_version"`
    Description  string    `json:"description"`
    PublishedDate time.Time `json:"published_date"`
    Link         string    `json:"link"`
    RiskFactors  []string  `json:"risk_factors"`
}

type Summary struct {
    TotalVulnerabilities int            `json:"total_vulnerabilities"`
    SeverityCounts      map[string]int `json:"severity_counts"`
    FixableCount        int            `json:"fixable_count"`
    Compliant          bool           `json:"compliant"`
}

type VulnerabilityFilters struct {
    Severity    string
    Status      string
    PackageName string
}

type Statistics struct {
    TotalScans              int            `json:"total_scans"`
    VulnerabilitiesBySeverity map[string]int `json:"vulnerabilities_by_severity"`
    ActiveVulnerabilities    int            `json:"active_vulnerabilities"`
    FixedVulnerabilities     int            `json:"fixed_vulnerabilities"`
}