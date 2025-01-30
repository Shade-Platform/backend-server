package main

import (
	"log"
	"net/http"
	"shade_web_server/infrastructure"
	"shade_web_server/routers"

	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize the database connection
	// dbConn, err := infrastructure.NewDBConnection()
	// if err != nil {
	// 	log.Fatalf("Failed to connect to the database: %v", err)
	// }
	// defer dbConn.Close()

	// Check for command-line arguments
	// infrastructure.MigrationsCliArguments(dbConn)

	// Initialize the router
	// r := routers.InitializeUsersRouter(dbConn)
	
	// Initialize the cluster connection
	clientset, err := infrastructure.NewClusterConnection()
	if err != nil {
		log.Fatalf("Failed to connet to the cluster: %v", err)
	}

	// Initialize the router
	r := routers.InitializeContainersRouter(clientset)

	// Start the server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
