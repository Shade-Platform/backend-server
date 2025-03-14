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

	// Check for command-line arguments (e.g., migrations)
	infrastructure.MigrationsCliArguments(dbConn)

	// Initialize the routers
	userRouter := routers.InitializeUsersRouter(dbConn)
	authRouter := routers.InitializeAuthRouter(dbConn)

	// Combine all routers into a single router
	mainRouter := http.NewServeMux()
	mainRouter.Handle("/users/", userRouter)
	mainRouter.Handle("/auth/", authRouter)

	// Configure CORS
	corsOptions := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Allow all origins (change to specific domains in production)
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Wrap the main router with CORS middleware
	handler := corsOptions(mainRouter)

	// Start the server
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
