-- high_severity_vulnerabilities
SELECT vuln_id,
       scan_id,
       severity,
       cvss,
       status,
       package_name,
       current_version,
       fixed_version,
       description,
       published_date,
       link
FROM vulnerabilities
WHERE severity IN ('CRITICAL', 'HIGH')
ORDER BY published_date DESC;

--- Aggregated Count by Package
SELECT package_name,
       COUNT(*) AS vuln_count,
       AVG(cvss) AS avg_cvss,
       MAX(cvss) AS max_cvss
FROM vulnerabilities
WHERE severity IN ('CRITICAL', 'HIGH')
GROUP BY package_name
ORDER BY max_cvss DESC;

--- Recent High-Severity Vulnerabilities
SELECT vuln_id,
       scan_id,
       severity,
       cvss,
       status,
       package_name,
       current_version,
       fixed_version,
       description,
       published_date,
       link
FROM vulnerabilities
WHERE severity IN ('CRITICAL', 'HIGH')
  AND published_date >= NOW() - INTERVAL '7 days'
ORDER BY published_date DESC;

--- Vulnerabilities with Associated Risk Factors
SELECT v.vuln_id,
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
       rf.factor
FROM vulnerabilities v
LEFT JOIN risk_factors rf ON v.vuln_id = rf.vuln_id
WHERE v.severity IN ('CRITICAL', 'HIGH')
ORDER BY v.published_date DESC;




