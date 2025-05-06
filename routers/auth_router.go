package routers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"shade_web_server/core/auth"
	"shade_web_server/core/users"

	"github.com/gorilla/mux"
)

// InitializeAuthRouter sets up authentication routes
func InitializeAuthRouter(dbConn *sql.DB) *mux.Router {
	// Initialize user repository and service
	repo := users.NewMySQLUserRepository(dbConn)
	userService := users.NewUserService(repo)

	// Initialize AuthService
	authService := auth.NewAuthService(userService)

	r := mux.NewRouter()

	// Pass `authService` to handlers
	r.HandleFunc("/auth/signup/", func(w http.ResponseWriter, r *http.Request) {
		signupHandler(w, r, userService)
	}).Methods("POST")

	r.HandleFunc("/auth/login/", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, authService)
	}).Methods("POST")

	return r
}

func signupHandler(w http.ResponseWriter, r *http.Request, userService *users.UserService) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Decode request body
	var requestBody auth.Signup
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Store user in the database
	_, err = userService.CreateUser(requestBody.Name, requestBody.Email, requestBody.Password)
	if err != nil {
		fmt.Printf("%v\n", err) // Use fmt.Printf for formatted output
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	response := map[string]string{"message": "User created successfully"}
	json.NewEncoder(w).Encode(response)
}

func loginHandler(w http.ResponseWriter, r *http.Request, authService *auth.AuthService) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	var requestBody auth.Login

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Authenticate user using the AuthService instance
	token, err := authService.AuthenticateUser(requestBody.Email, requestBody.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Return JWT token as JSON response
	response := map[string]string{"token": token}

	// Marshal the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the response header and write the response body
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
