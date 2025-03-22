package routers

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

// InitializeAuthRouter sets up authentication routes
func InitializeClusterRouter(dbConn *sql.DB) *mux.Router {
	r := mux.NewRouter()

	// Pass `authService` to handlers
	r.HandleFunc("/grafana", func(w http.ResponseWriter, r *http.Request) {
		grafanaHandler(w, r)
	}).Methods("GET")

	return r
}

func grafanaHandler(w http.ResponseWriter, r *http.Request) {

	// todo: policies authorization

	http.Redirect(w, r, "http://grafana-url/", http.StatusFound) // 302 Found
}
