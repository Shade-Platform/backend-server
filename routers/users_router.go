package routers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"shade_web_server/core/users"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var userService *users.UserService

// Sets up all the routes, accepting the DB connection as an argument
func InitializeUsersRouter(dbConn *sql.DB) *mux.Router {
	// Initialize the UserRepository and UserService
	repo := users.NewMySQLUserRepository(dbConn) // Pass dbConn here
	userService = users.NewUserService(repo)

	r := mux.NewRouter()

	// Define routes and pass userService to the handlers
	r.HandleFunc("/users/", getUsers).Methods("GET")
	r.HandleFunc("/users/create/", createUserHandler).Methods("POST")
	r.HandleFunc("/users/sub-users/create/", createSubUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", getUserByID).Methods("GET")
	r.HandleFunc("/users/email/{email}", getUserByEmail).Methods("GET")

	return r
}

// Handler to get all users
func getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := userService.GetAllUsers()
	if err != nil {
		fmt.Printf("%v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Set the response header for JSON
	w.Header().Set("Content-Type", "application/json")
	// Encode the list of users into the response
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode users", http.StatusInternalServerError)
	}
}

// Handler to create a new user
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user users.User

	// Decode JSON body into user struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Create the user
	createdUser, err := userService.CreateUser(user.Name, user.Email, user.Password)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return the created user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)
}

// Handler to create a new sub-user
func createSubUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RootUserID string `json:"root_user_id"` // ID of the root user
		Name       string `json:"name"`         // Name of the sub-user
		Email      string `json:"email"`        // Email of the sub-user
		Password   string `json:"password"`     // Password of the sub-user
	}

	// Decode JSON body into input struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Parse the root user ID
	rootUserID, err := uuid.Parse(input.RootUserID)
	if err != nil {
		http.Error(w, "Invalid root_user_id", http.StatusBadRequest)
		return
	}

	// Create the sub-user
	subUser, err := userService.CreateSubUser(rootUserID, input.Name, input.Email, input.Password)
	if err != nil {
		http.Error(w, "Failed to create sub-user", http.StatusInternalServerError)
		return
	}

	// Return the created sub-user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subUser)
}

// Handler to get a user by ID
func getUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch the user by ID
	user, err := userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getUserByEmail(w http.ResponseWriter, r *http.Request) {
	// Extract the email from the URL
	vars := mux.Vars(r)
	email := vars["email"]

	user, err := userService.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
