package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/SarathLUN/go-auth-service/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	// Access configuration values:
	fmt.Printf("Database Host: %s\n", cfg.DBHost)
	fmt.Printf("JWT Secret: %s\n", cfg.JWTSecret) // Be careful about logging secrets!
	fmt.Printf("Database Connection String: %s\n", cfg.GetDBConnectionString())
	fmt.Printf("Application Port: %s\n", cfg.AppPort)

	// Example HTTP server (replace with your actual application logic)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!  The database host is: %s", cfg.DBHost)
	})

	log.Printf("Server starting on port %s...\n", cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatal(err)
	}
}
