package main

import (
    "log"
    "devops-assign/internal/api"
    "devops-assign/internal/db"
)

func main() {
    database, err := db.Initialize()
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    server := api.NewServer(database)
    log.Printf("Server starting on :8000")
    log.Fatal(server.Start(":8000"))
}