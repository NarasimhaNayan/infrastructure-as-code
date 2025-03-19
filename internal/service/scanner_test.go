package service

import (
    "testing"
    "time"
    "devops-assign/internal/models"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
)

func TestProcessScanResults(t *testing.T) {
    // Create mock db
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock: %v", err)
    }
    defer db.Close()

    service := NewScannerService(db)

    // Test case 1: Successful scan processing
    t.Run("Successful scan processing", func(t *testing.T) {
        scanResult := []models.ScanResult{
            {
                ScanID:      "test_scan_001",
                Timestamp:   time.Now(),
                ScanStatus:  "completed",
                ResourceType: "container",
                ResourceName: "test-container:latest",
                Vulnerabilities: []models.Vulnerability{
                    {
                        ID:           "CVE-2025-1234",
                        Severity:     "HIGH",
                        CVSS:        8.5,
                        Status:      "active",
                        PackageName: "openssl",
                        CurrVersion: "1.1.1t-r0",
                        RiskFactors: []string{"RCE"},
                    },
                },
            },
        }

        // Expect transaction begin
        mock.ExpectBegin()

        // Expect scan result insert
        mock.ExpectExec("INSERT INTO scan_results").
            WithArgs(
                scanResult[0].ScanID,
                scanResult[0].Timestamp,
                scanResult[0].ScanStatus,
                scanResult[0].ResourceType,
                scanResult[0].ResourceName,
            ).
            WillReturnResult(sqlmock.NewResult(1, 1))

        // Expect vulnerability insert
        mock.ExpectExec("INSERT INTO vulnerabilities").
            WithArgs(
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
                sqlmock.AnyArg(),
            ).
            WillReturnResult(sqlmock.NewResult(1, 1))

        // Expect risk factors deletion
        mock.ExpectExec("DELETE FROM risk_factors").
            WithArgs(scanResult[0].Vulnerabilities[0].ID).
            WillReturnResult(sqlmock.NewResult(1, 1))

        // Expect risk factors insert
        mock.ExpectExec("INSERT INTO risk_factors").
            WithArgs(
                scanResult[0].Vulnerabilities[0].ID,
                "RCE",
            ).
            WillReturnResult(sqlmock.NewResult(1, 1))

        // Expect transaction commit
        mock.ExpectCommit()

        err := service.ProcessScanResults(scanResult)
        assert.NoError(t, err)
    })

    // Test case 2: Empty scan results
    t.Run("Empty scan results", func(t *testing.T) {
        err := service.ProcessScanResults([]models.ScanResult{})
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "no scan results provided")
    })
}

func TestGetVulnerabilities(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock: %v", err)
    }
    defer db.Close()

    service := NewScannerService(db)

    t.Run("Get vulnerabilities with filters", func(t *testing.T) {
        filters := models.VulnerabilityFilters{
            Severity: "HIGH",
            Status:  "active",
        }

        rows := sqlmock.NewRows([]string{
            "vuln_id", "scan_id", "severity", "cvss", "status",
            "package_name", "current_version", "fixed_version",
            "description", "published_date", "link", "risk_factors",
        }).AddRow(
            "CVE-2025-1234", "scan_001", "HIGH", 8.5, "active",
            "openssl", "1.1.1t-r0", "1.1.1u-r0",
            "Test vulnerability", time.Now(), "http://example.com",
            `["RCE", "High CVSS"]`,
        )

        mock.ExpectQuery("SELECT (.+) FROM vulnerabilities").
            WithArgs("HIGH", "active").
            WillReturnRows(rows)

        vulns, err := service.GetVulnerabilities(filters)
        assert.NoError(t, err)
        assert.Len(t, vulns, 1)
        assert.Equal(t, "CVE-2025-1234", vulns[0].ID)
        assert.Equal(t, "HIGH", vulns[0].Severity)
    })
}

func TestGetVulnerabilityByID(t *testing.T) {
    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
    if err != nil {
        t.Fatalf("Failed to create mock: %v", err)
    }
    defer db.Close()

    service := NewScannerService(db)

    t.Run("Get existing vulnerability", func(t *testing.T) {
        vuln_id := "CVE-2025-1234"
        publishedDate := time.Now()

        expectedQuery := `SELECT v.vuln_id, v.scan_id, v.severity, v.cvss, v.status, v.package_name, v.current_version, v.fixed_version, v.description, v.published_date, v.link, ARRAY_AGG(rf.factor) as risk_factors FROM vulnerabilities v LEFT JOIN risk_factors rf ON v.vuln_id = rf.vuln_id WHERE v.vuln_id = $1 GROUP BY v.vuln_id, v.scan_id, v.severity, v.cvss, v.status, v.package_name, v.current_version, v.fixed_version, v.description, v.published_date, v.link`

        rows := sqlmock.NewRows([]string{
            "vuln_id", "scan_id", "severity", "cvss", "status",
            "package_name", "current_version", "fixed_version",
            "description", "published_date", "link", "risk_factors",
        }).AddRow(
            vuln_id, "scan_001", "HIGH", 8.5, "active",
            "openssl", "1.1.1t-r0", "1.1.1u-r0",
            "Test vulnerability", publishedDate, "http://example.com",
            "{RCE,High CVSS}", // PostgreSQL array format as string
        )

        mock.ExpectQuery(expectedQuery).
            WithArgs(vuln_id).
            WillReturnRows(rows)

        vuln, err := service.GetVulnerabilityByID(vuln_id)
        assert.NoError(t, err)
        assert.NotNil(t, vuln)
        if vuln != nil {
            assert.Equal(t, vuln_id, vuln.ID)
            assert.Equal(t, "HIGH", string(vuln.Severity))
            assert.Equal(t, float64(8.5), vuln.CVSS)
            assert.Equal(t, "active", vuln.Status)
            assert.Equal(t, "openssl", vuln.PackageName)
            assert.ElementsMatch(t, []string{"RCE", "High CVSS"}, vuln.RiskFactors)
        }
    })
}

func TestGetStatistics(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock: %v", err)
    }
    defer db.Close()

    service := NewScannerService(db)

    t.Run("Get statistics", func(t *testing.T) {
        // Mock total scans
        totalScansRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
        mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM scan_results").
            WillReturnRows(totalScansRows)

        // Mock severity counts
        severityRows := sqlmock.NewRows([]string{"severity", "count"}).
            AddRow("HIGH", 5).
            AddRow("MEDIUM", 3).
            AddRow("LOW", 2)
        mock.ExpectQuery("SELECT severity, COUNT\\(\\*\\) FROM vulnerabilities").
            WillReturnRows(severityRows)

        // Mock active/fixed counts
        statusRows := sqlmock.NewRows([]string{"active_count", "fixed_count"}).
            AddRow(7, 3)
        mock.ExpectQuery("SELECT COUNT\\(\\*\\) FILTER").
            WillReturnRows(statusRows)

        stats, err := service.GetStatistics()
        assert.NoError(t, err)
        assert.NotNil(t, stats)
        assert.Equal(t, 10, stats.TotalScans)
        assert.Equal(t, 7, stats.ActiveVulnerabilities)
        assert.Equal(t, 3, stats.FixedVulnerabilities)
        assert.Equal(t, 5, stats.VulnerabilitiesBySeverity["HIGH"])
    })
}

func TestGetRecentScans(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock: %v", err)
    }
    defer db.Close()

    service := NewScannerService(db)

    t.Run("Get recent scans", func(t *testing.T) {
        timestamp := time.Now()
        rows := sqlmock.NewRows([]string{
            "scan_id", "timestamp", "scan_status",
            "resource_type", "resource_name", "vuln_count",
        }).AddRow(
            "scan_001", timestamp, "completed",
            "container", "test:latest", 5,
        )

        mock.ExpectQuery("SELECT (.+) FROM scan_results").
            WithArgs(10).
            WillReturnRows(rows)

        scans, err := service.GetRecentScans(10)
        assert.NoError(t, err)
        assert.Len(t, scans, 1)
        assert.Equal(t, "scan_001", scans[0].ScanID)
        assert.Equal(t, 5, scans[0].Summary.TotalVulnerabilities)
    })
}