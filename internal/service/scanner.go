// internal/service/scanner.go
package service

import (
    "database/sql"
    "errors"
	"encoding/json"
    "strings"
    "fmt"
    "time"
    "devops-assign/internal/models"
)

type ScannerService struct {
    db *sql.DB
}

func NewScannerService(db *sql.DB) *ScannerService {
    return &ScannerService{db: db}
}

// ProcessScanResults handles the processing and storage of scan results
func (s *ScannerService) ProcessScanResults(scans []models.ScanResult) error {
    if len(scans) == 0 {
        return errors.New("no scan results provided")
    }

    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer tx.Rollback()

    for _, scan := range scans {
        // Insert or update scan result
        _, err = tx.Exec(`
            INSERT INTO scan_results 
                (scan_id, timestamp, scan_status, resource_type, resource_name)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (scan_id) DO UPDATE 
            SET 
                timestamp = EXCLUDED.timestamp,
                scan_status = EXCLUDED.scan_status,
                resource_type = EXCLUDED.resource_type,
                resource_name = EXCLUDED.resource_name
        `, scan.ScanID, scan.Timestamp, scan.ScanStatus, scan.ResourceType, scan.ResourceName)

        if err != nil {
            return fmt.Errorf("failed to insert scan result: %v", err)
        }

        // Process vulnerabilities
        for _, vuln := range scan.Vulnerabilities {
            err = s.processVulnerability(tx, scan.ScanID, vuln)
            if err != nil {
                return fmt.Errorf("failed to process vulnerability: %v", err)
            }
        }
    }

    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    return nil
}

// processVulnerability handles individual vulnerability processing
func (s *ScannerService) processVulnerability(tx *sql.Tx, scanID string, vuln models.Vulnerability) error {
    // Insert or update vulnerability
    _, err := tx.Exec(`
        INSERT INTO vulnerabilities (
            vuln_id, scan_id, severity, cvss, status, 
            package_name, current_version, fixed_version, 
            description, published_date, link
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (vuln_id) DO UPDATE 
        SET 
            scan_id = EXCLUDED.scan_id,
            severity = EXCLUDED.severity,
            cvss = EXCLUDED.cvss,
            status = EXCLUDED.status,
            package_name = EXCLUDED.package_name,
            current_version = EXCLUDED.current_version,
            fixed_version = EXCLUDED.fixed_version,
            description = EXCLUDED.description,
            published_date = EXCLUDED.published_date,
            link = EXCLUDED.link
    `, vuln.ID, scanID, vuln.Severity, vuln.CVSS, vuln.Status,
       vuln.PackageName, vuln.CurrVersion, vuln.FixedVersion,
       vuln.Description, vuln.PublishedDate, vuln.Link)

    if err != nil {
        return err
    }

    // Delete existing risk factors for this vulnerability
    _, err = tx.Exec("DELETE FROM risk_factors WHERE vuln_id = $1", vuln.ID)
    if err != nil {
        return err
    }

    // Insert new risk factors
    for _, factor := range vuln.RiskFactors {
        _, err = tx.Exec(`
            INSERT INTO risk_factors (vuln_id, factor)
            VALUES ($1, $2)
        `, vuln.ID, factor)

        if err != nil {
            return err
        }
    }

    return nil
}

// GetVulnerabilities retrieves vulnerabilities based on filters
func (s *ScannerService) GetVulnerabilities(filters models.VulnerabilityFilters) ([]models.Vulnerability, error) {
    query := `
        SELECT 
            v.vuln_id,
            v.scan_id,
            v.severity,
            v.cvss,
            v.status,
            v.package_name,
            v.current_version,
            v.fixed_version,
            v.description,
            v.published_date,
            v.link,
            COALESCE(array_to_json(array_agg(rf.factor))::text, '[]') as risk_factors
        FROM vulnerabilities v
        LEFT JOIN risk_factors rf ON v.vuln_id = rf.vuln_id
        WHERE 1=1
    `
    params := []interface{}{}
    paramCount := 1

    if filters.Severity != "" {
        query += fmt.Sprintf(" AND v.severity = $%d", paramCount)
        params = append(params, filters.Severity)
        paramCount++
    }

    if filters.Status != "" {
        query += fmt.Sprintf(" AND v.status = $%d", paramCount)
        params = append(params, filters.Status)
        paramCount++
    }

    if filters.PackageName != "" {
        query += fmt.Sprintf(" AND v.package_name = $%d", paramCount)
        params = append(params, filters.PackageName)
        paramCount++
    }

    query += ` GROUP BY v.vuln_id, v.scan_id, v.severity, v.cvss, v.status, 
               v.package_name, v.current_version, v.fixed_version, v.description, 
               v.published_date, v.link ORDER BY v.published_date DESC`

    rows, err := s.db.Query(query, params...)
    if err != nil {
        return nil, fmt.Errorf("failed to query vulnerabilities: %v", err)
    }
    defer rows.Close()

    var vulnerabilities []models.Vulnerability
    for rows.Next() {
        var v models.Vulnerability
        var scanID string
        var riskFactorsJSON string
        
        err := rows.Scan(
            &v.ID,
            &scanID,
            &v.Severity,
            &v.CVSS,
            &v.Status,
            &v.PackageName,
            &v.CurrVersion,
            &v.FixedVersion,
            &v.Description,
            &v.PublishedDate,
            &v.Link,
            &riskFactorsJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan vulnerability row: %v", err)
        }

        // Parse risk factors JSON
        err = json.Unmarshal([]byte(riskFactorsJSON), &v.RiskFactors)
        if err != nil {
            return nil, fmt.Errorf("failed to parse risk factors: %v", err)
        }

        vulnerabilities = append(vulnerabilities, v)
    }

    return vulnerabilities, nil
}

func (s *ScannerService) GetVulnerabilityByID(id string) (*models.Vulnerability, error) {
    query := `SELECT v.vuln_id, v.scan_id, v.severity, v.cvss, v.status, v.package_name, v.current_version, v.fixed_version, v.description, v.published_date, v.link, ARRAY_AGG(rf.factor) as risk_factors FROM vulnerabilities v LEFT JOIN risk_factors rf ON v.vuln_id = rf.vuln_id WHERE v.vuln_id = $1 GROUP BY v.vuln_id, v.scan_id, v.severity, v.cvss, v.status, v.package_name, v.current_version, v.fixed_version, v.description, v.published_date, v.link`

    var v models.Vulnerability
    var scanID string
    var riskFactors string // Change this to string

    err := s.db.QueryRow(query, id).Scan(
        &v.ID,
        &scanID,
        &v.Severity,
        &v.CVSS,
        &v.Status,
        &v.PackageName,
        &v.CurrVersion,
        &v.FixedVersion,
        &v.Description,
        &v.PublishedDate,
        &v.Link,
        &riskFactors, // Scan into string first
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get vulnerability: %v", err)
    }

    // Parse the PostgreSQL array string format into string slice
    riskFactors = strings.Trim(riskFactors, "{}")
    if riskFactors != "" {
        v.RiskFactors = strings.Split(riskFactors, ",")
    } else {
        v.RiskFactors = []string{}
    }

    return &v, nil
}

// GetStatistics retrieves overall vulnerability statistics
func (s *ScannerService) GetStatistics() (*models.Statistics, error) {
    stats := &models.Statistics{
        VulnerabilitiesBySeverity: make(map[string]int),
    }

    // Get total scans
    err := s.db.QueryRow(`
        SELECT COUNT(*) 
        FROM scan_results
    `).Scan(&stats.TotalScans)
    if err != nil {
        return nil, fmt.Errorf("failed to get total scans: %v", err)
    }

    // Get vulnerabilities by severity
    rows, err := s.db.Query(`
        SELECT severity, COUNT(*) 
        FROM vulnerabilities 
        GROUP BY severity
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to get severity counts: %v", err)
    }
    defer rows.Close()

    for rows.Next() {
        var severity string
        var count int
        if err := rows.Scan(&severity, &count); err != nil {
            return nil, fmt.Errorf("failed to scan severity row: %v", err)
        }
        stats.VulnerabilitiesBySeverity[severity] = count
    }

    // Get active and fixed vulnerability counts
    err = s.db.QueryRow(`
        SELECT 
            COUNT(*) FILTER (WHERE status = 'active') as active_count,
            COUNT(*) FILTER (WHERE status = 'fixed') as fixed_count
        FROM vulnerabilities
    `).Scan(&stats.ActiveVulnerabilities, &stats.FixedVulnerabilities)
    if err != nil {
        return nil, fmt.Errorf("failed to get vulnerability counts: %v", err)
    }

    return stats, nil
}

// GetRecentScans retrieves the most recent scan results
func (s *ScannerService) GetRecentScans(limit int) ([]models.ScanResult, error) {
    if limit <= 0 {
        limit = 10
    }

    query := `
        SELECT 
            sr.scan_id,
            sr.timestamp,
            sr.scan_status,
            sr.resource_type,
            sr.resource_name,
            COUNT(v.vuln_id) as vuln_count
        FROM scan_results sr
        LEFT JOIN vulnerabilities v ON sr.scan_id = v.scan_id
        GROUP BY sr.scan_id, sr.timestamp, sr.scan_status, sr.resource_type, sr.resource_name
        ORDER BY sr.timestamp DESC
        LIMIT $1
    `

    rows, err := s.db.Query(query, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to get recent scans: %v", err)
    }
    defer rows.Close()

    var scans []models.ScanResult
    for rows.Next() {
        var scan models.ScanResult
        var vulnCount int
        err := rows.Scan(
            &scan.ScanID,
            &scan.Timestamp,
            &scan.ScanStatus,
            &scan.ResourceType,
            &scan.ResourceName,
            &vulnCount,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan row: %v", err)
        }
        scan.Summary = models.Summary{
            TotalVulnerabilities: vulnCount,
        }
        scans = append(scans, scan)
    }

    return scans, nil
}

// GetVulnerabilityTrends gets vulnerability trends over time
func (s *ScannerService) GetVulnerabilityTrends(days int) (map[string][]int, error) {
    if days <= 0 {
        days = 30
    }

    query := `
        WITH RECURSIVE dates AS (
            SELECT date_trunc('day', now()) as day
            UNION ALL
            SELECT day - interval '1 day'
            FROM dates
            WHERE day > date_trunc('day', now()) - interval '1 day' * $1
        ),
        daily_counts AS (
            SELECT 
                date_trunc('day', v.published_date) as day,
                v.severity,
                count(*) as count
            FROM vulnerabilities v
            WHERE v.published_date >= now() - interval '1 day' * $1
            GROUP BY date_trunc('day', v.published_date), v.severity
        )
        SELECT 
            d.day,
            dc.severity,
            COALESCE(dc.count, 0) as count
        FROM dates d
        LEFT JOIN daily_counts dc ON d.day = dc.day
        ORDER BY d.day, dc.severity
    `

    rows, err := s.db.Query(query, days)
    if err != nil {
        return nil, fmt.Errorf("failed to get vulnerability trends: %v", err)
    }
    defer rows.Close()

    trends := make(map[string][]int)
    for rows.Next() {
        var day time.Time
        var severity string
        var count int
        err := rows.Scan(&day, &severity, &count)
        if err != nil {
            return nil, fmt.Errorf("failed to scan trend row: %v", err)
        }
        if trends[severity] == nil {
            trends[severity] = make([]int, days)
        }
        dayIndex := days - int(time.Since(day).Hours()/24) - 1
        if dayIndex >= 0 && dayIndex < days {
            trends[severity][dayIndex] = count
        }
    }

    return trends, nil
}