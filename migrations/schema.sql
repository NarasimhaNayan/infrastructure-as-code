-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types for fixed values
CREATE TYPE severity_level AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'UNKNOWN');
CREATE TYPE scan_status_type AS ENUM ('completed', 'failed', 'in_progress');

-- Scan Results Table
CREATE TABLE IF NOT EXISTS scan_results (
    scan_id VARCHAR(100) PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    scan_status scan_status_type NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vulnerabilities Table
CREATE TABLE IF NOT EXISTS vulnerabilities (
    vuln_id VARCHAR(50) PRIMARY KEY,
    scan_id VARCHAR(100) REFERENCES scan_results(scan_id),
    severity severity_level NOT NULL,
    cvss DECIMAL(4,1) NOT NULL CHECK (cvss >= 0 AND cvss <= 10),
    status VARCHAR(20) NOT NULL,
    package_name VARCHAR(100) NOT NULL,
    current_version VARCHAR(50) NOT NULL,
    fixed_version VARCHAR(50),
    description TEXT,
    published_date TIMESTAMP NOT NULL,
    link VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Risk Factors Table
CREATE TABLE IF NOT EXISTS risk_factors (
    vuln_id VARCHAR(50) REFERENCES vulnerabilities(vuln_id) ON DELETE CASCADE,
    factor VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (vuln_id, factor)
);

-- Scan Summary Table
CREATE TABLE IF NOT EXISTS scan_summaries (
    scan_id VARCHAR(100) PRIMARY KEY REFERENCES scan_results(scan_id) ON DELETE CASCADE,
    total_vulnerabilities INT NOT NULL DEFAULT 0,
    critical_count INT NOT NULL DEFAULT 0,
    high_count INT NOT NULL DEFAULT 0,
    medium_count INT NOT NULL DEFAULT 0,
    low_count INT NOT NULL DEFAULT 0,
    fixable_count INT NOT NULL DEFAULT 0,
    compliant BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vulnerability History Table for tracking changes
CREATE TABLE IF NOT EXISTS vulnerability_history (
    id SERIAL PRIMARY KEY,
    vuln_id VARCHAR(50) REFERENCES vulnerabilities(vuln_id),
    scan_id VARCHAR(100) REFERENCES scan_results(scan_id),
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    old_version VARCHAR(50),
    new_version VARCHAR(50),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_scan_results_timestamp ON scan_results(timestamp);
CREATE INDEX idx_scan_results_status ON scan_results(scan_status);
CREATE INDEX idx_scan_results_resource ON scan_results(resource_type, resource_name);

CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX idx_vulnerabilities_status ON vulnerabilities(status);
CREATE INDEX idx_vulnerabilities_package ON vulnerabilities(package_name);
CREATE INDEX idx_vulnerabilities_cvss ON vulnerabilities(cvss);
CREATE INDEX idx_vulnerabilities_scan_id ON vulnerabilities(scan_id);
CREATE INDEX idx_vulnerabilities_published ON vulnerabilities(published_date);

CREATE INDEX idx_risk_factors_factor ON risk_factors(factor);
CREATE INDEX idx_vulnerability_history_vuln_id ON vulnerability_history(vuln_id);
CREATE INDEX idx_vulnerability_history_changed_at ON vulnerability_history(changed_at);

-- Create a function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update timestamps
CREATE TRIGGER update_scan_results_updated_at
    BEFORE UPDATE ON scan_results
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vulnerabilities_updated_at
    BEFORE UPDATE ON vulnerabilities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scan_summaries_updated_at
    BEFORE UPDATE ON scan_summaries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create a function to track vulnerability changes
CREATE OR REPLACE FUNCTION track_vulnerability_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'UPDATE') THEN
        IF (NEW.status != OLD.status OR NEW.current_version != OLD.current_version) THEN
            INSERT INTO vulnerability_history (
                vuln_id,
                scan_id,
                old_status,
                new_status,
                old_version,
                new_version
            ) VALUES (
                NEW.vuln_id,
                NEW.scan_id,
                OLD.status,
                NEW.status,
                OLD.current_version,
                NEW.current_version
            );
        END IF;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for tracking vulnerability changes
CREATE TRIGGER track_vulnerability_changes_trigger
    AFTER UPDATE ON vulnerabilities
    FOR EACH ROW
    EXECUTE FUNCTION track_vulnerability_changes();

-- Create a function to update scan summary
CREATE OR REPLACE FUNCTION update_scan_summary()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO scan_summaries (
        scan_id,
        total_vulnerabilities,
        critical_count,
        high_count,
        medium_count,
        low_count,
        fixable_count,
        compliant
    )
    SELECT 
        NEW.scan_id,
        COUNT(*),
        COUNT(*) FILTER (WHERE severity = 'CRITICAL'),
        COUNT(*) FILTER (WHERE severity = 'HIGH'),
        COUNT(*) FILTER (WHERE severity = 'MEDIUM'),
        COUNT(*) FILTER (WHERE severity = 'LOW'),
        COUNT(*) FILTER (WHERE fixed_version IS NOT NULL),
        CASE 
            WHEN COUNT(*) FILTER (WHERE severity IN ('CRITICAL', 'HIGH')) = 0 THEN true 
            ELSE false 
        END
    FROM vulnerabilities
    WHERE scan_id = NEW.scan_id
    ON CONFLICT (scan_id) DO UPDATE
    SET 
        total_vulnerabilities = EXCLUDED.total_vulnerabilities,
        critical_count = EXCLUDED.critical_count,
        high_count = EXCLUDED.high_count,
        medium_count = EXCLUDED.medium_count,
        low_count = EXCLUDED.low_count,
        fixable_count = EXCLUDED.fixable_count,
        compliant = EXCLUDED.compliant,
        updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for updating scan summary
CREATE TRIGGER update_scan_summary_trigger
    AFTER INSERT OR UPDATE ON vulnerabilities
    FOR EACH ROW
    EXECUTE FUNCTION update_scan_summary();

-- Create views for common queries
CREATE OR REPLACE VIEW vulnerability_stats AS
SELECT
    DATE_TRUNC('day', published_date) as date,
    severity,
    COUNT(*) as count,
    AVG(cvss) as avg_cvss
FROM vulnerabilities
GROUP BY DATE_TRUNC('day', published_date), severity
ORDER BY date DESC;

CREATE OR REPLACE VIEW high_risk_packages AS
SELECT
    package_name,
    COUNT(*) as vuln_count,
    MAX(cvss) as max_cvss,
    COUNT(*) FILTER (WHERE severity = 'CRITICAL') as critical_count,
    COUNT(*) FILTER (WHERE severity = 'HIGH') as high_count
FROM vulnerabilities
WHERE status = 'active'
GROUP BY package_name
HAVING COUNT(*) > 1
ORDER BY max_cvss DESC;