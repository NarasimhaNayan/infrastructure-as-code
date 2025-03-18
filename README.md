# Vulnerability Scanner Service

A Go-based service for processing and analyzing vulnerability scan results. This service provides APIs to upload scan results, query vulnerabilities, and get statistical information about your security posture.

## Features

- Process vulnerability scan results in JSON format
- Store vulnerability data in PostgreSQL database
- Query vulnerabilities with filtering options
- Get vulnerability statistics
- Track scan history
- Prevent duplicate entries
- Handle risk factors and severity levels

## Prerequisites

- Docker and Docker Compose
- Go 1.20 or higher (for local development)
- PostgreSQL 13 or higher (for local development)

## Project Structure

```
/vuln-scanner
  ├── cmd/                  # Application entrypoints
  ├── internal/            # Private application code
  │   ├── api/            # HTTP handlers and routes
  │   ├── models/         # Data models
  │   ├── db/             # Database operations
  │   └── service/        # Business logic
  ├── migrations/         # Database migrations
  ├── Dockerfile         # Docker build instructions
  ├── docker-compose.yml # Docker compose configuration
  └── README.md         # This file
```

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/chanduusc/Devops-task.git
cd Devops-task
```

2. Start the service using Docker Compose:
```bash
docker-compose up -d
```

3. The service will be available at `http://localhost:8000`

## API Endpoints

### Upload Scan Results
```bash
POST /api/scan
Content-Type: application/json

# Request body should contain the scan results in JSON format
```

### Query Vulnerabilities
```bash
GET /api/vulnerabilities?severity=CRITICAL&status=active&package=nginx
```

### Get Statistics
```bash
GET /api/stats
```

### Get Scan History
```bash
GET /api/scans
```

### Health Check
```bash
GET /api/health
```

## Development

1. Install dependencies:
```bash
go mod download
```

2. Set up the database:
```bash
psql -U postgres -f migrations/schema.sql
```

3. Run the service:
```bash
go run cmd/server/main.go
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (default: "postgresql://user:password@localhost:5432/vulndb?sslmode=disable")
- `PORT`: Server port (default: 8000)
- `APP_ENV`: Application environment (development/production)

## Testing

Run the tests:
```bash
go test ./...
```

## Docker Build

Build the Docker image:
```bash
docker build -t vuln-scanner .
```

Run the container:
```bash
docker run -p 8000:8000 vuln-scanner
```

## Example Scan Result Format

```json
{
  "scan_id": "scan_123456789",
  "timestamp": "2025-01-28T10:30:00Z",
  "scan_status": "completed",
  "resource_type": "container",
  "resource_name": "app-container:latest",
  "vulnerabilities": [
    {
      "id": "CVE-2024-1234",
      "severity": "HIGH",
      "cvss": 8.5,
      "status": "fixed",
      "package_name": "openssl",
      "current_version": "1.1.1t-r0",
      "fixed_version": "1.1.1u-r0",
      "description": "Buffer overflow vulnerability in OpenSSL",
      "published_date": "2024-01-15T00:00:00Z",
      "link": "https://nvd.nist.gov/vuln/detail/CVE-2024-1234",
      "risk_factors": [
        "Remote Code Execution",
        "High CVSS Score",
        "Public Exploit Available"
      ]
    }
  ]
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request