package main

import (
	"log"
	"net/http"
	"shade_web_server/infrastructure"
	"shade_web_server/routers"

	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize the database connection
	dbConn, err := infrastructure.NewDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer dbConn.Close()

	// Check for command-line arguments
	infrastructure.MigrationsCliArguments(dbConn)

	// Initialize the router
	// r := routers.InitializeUsersRouter(dbConn)
	r := routers.InitializeAuthRouter(dbConn)

	// Initialize the cluster connection
	// clientset, err := infrastructure.NewClusterConnection()
	// if err != nil {
	// 	log.Fatalf("Failed to connet to the cluster: %v", err)
	// }

	// Initialize the router
	// routers.InitializeContainersRouter(clientset)

	// Start the server

	corsOptions := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Change "*" to specific origins if needed
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "application/json"}),
	)(r)
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsOptions))
}
