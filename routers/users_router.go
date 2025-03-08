package routers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"shade_web_server/core/users"

	"github.com/gorilla/mux"
)

// Initialize the userService variable
var userService *users.UserService

// InitializeRouter sets up all the routes, accepting the DB connection as an argument
func InitializeUsersRouter(dbConn *sql.DB) *mux.Router {
	// Initialize the UserRepository and UserService
	repo := users.NewMySQLUserRepository(dbConn) // Pass dbConn here
	userService = users.NewUserService(repo)     // Create the UserService with the repository

	r := mux.NewRouter()

	// Define routes and pass userService to the handlers
	r.HandleFunc("/", aboutHandler).Methods("GET")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/create", createUserHandler).Methods("POST")

	return r
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := userService.GetAllUsers()
	if err != nil {
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

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello go user, This is the Home page!!")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user users.User

	// Decode JSON body into user struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Assuming userService is already initialized, create the user
	createdUser, err := userService.CreateUser(user.Name, user.Email, user.Password)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return the created user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)
}
