package db

import (
    "database/sql"
    "os"
    _ "github.com/lib/pq"
)

func Initialize() (*sql.DB, error) {
    connectionString := os.Getenv("DATABASE_URL")
    if connectionString == "" {
        connectionString = "postgresql://vulnuser:vulnpass@localhost:5432/vulndb?sslmode=disable"
    }
    
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        return nil, err
    }

    return db, nil
}