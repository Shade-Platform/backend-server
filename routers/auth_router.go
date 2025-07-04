package routers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"shade_web_server/core/auth"
	"shade_web_server/core/users"
	"shade_web_server/infrastructure/logger"
	"shade_web_server/middleware"

	"github.com/gorilla/mux"
)

// InitializeAuthRouter sets up authentication routes
func InitializeAuthRouter(dbConn *sql.DB) *mux.Router {
	repo := users.NewMySQLUserRepository(dbConn)
	userService := users.NewUserService(repo)
	authService := auth.NewAuthService(userService)

	r := mux.NewRouter()

	r.HandleFunc("/auth/signup/", func(w http.ResponseWriter, r *http.Request) {
		signupHandler(w, r, userService)
	}).Methods("POST")

	r.HandleFunc("/auth/login/", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, authService)
	}).Methods("POST")

	r.Handle("/auth/me/", middleware.JWTAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)

		logger.Log.WithFields(map[string]interface{}{
			"event":   "auth_me",
			"user_id": userID,
			"ip":      r.RemoteAddr,
		}).Info("Token verified successfully")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Authenticated successfully",
			"user_id": userID,
		})
	}))).Methods("GET")

	return r
}

func signupHandler(w http.ResponseWriter, r *http.Request, userService *users.UserService) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	var requestBody auth.Signup
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event":  "signup",
			"ip":     r.RemoteAddr,
			"method": r.Method,
			"path":   r.URL.Path,
			"error":  err.Error(),
		}).Warn("Invalid signup input")
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = userService.CreateUser(requestBody.Name, requestBody.Email, requestBody.Password)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event":  "signup_failed",
			"user":   requestBody.Email,
			"ip":     r.RemoteAddr,
			"method": r.Method,
			"path":   r.URL.Path,
			"error":  err.Error(),
		}).Error("Failed to create user")
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Log.WithFields(map[string]interface{}{
		"event":  "signup_success",
		"user":   requestBody.Email,
		"ip":     r.RemoteAddr,
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("User signed up")

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
		logger.Log.WithFields(map[string]interface{}{
			"event":  "login_invalid_input",
			"ip":     r.RemoteAddr,
			"method": r.Method,
			"path":   r.URL.Path,
			"error":  err.Error(),
		}).Warn("Invalid login input")
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	token, err := authService.AuthenticateUser(requestBody.Email, requestBody.Password)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event":  "login_failed",
			"user":   requestBody.Email,
			"ip":     r.RemoteAddr,
			"method": r.Method,
			"path":   r.URL.Path,
			"error":  err.Error(),
		}).Warn("Login failed")
		http.Error(w, "Invalid credentials: "+err.Error(), http.StatusUnauthorized)
		return
	}

	logger.Log.WithFields(map[string]interface{}{
		"event":  "login_success",
		"user":   requestBody.Email,
		"ip":     r.RemoteAddr,
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Login successful")

	response := map[string]string{"token": token}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event": "login_json_error",
			"user":  requestBody.Email,
			"error": err.Error(),
		}).Error("Failed to marshal login response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}
