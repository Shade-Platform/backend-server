package routers

import (
	"encoding/json"
	"net/http"
	"shade_web_server/core/containers"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
)

// Container Service handling all container operations
var containerService *containers.ContainerService

func InitializeContainersRouter(clientset *kubernetes.Clientset) *mux.Router {
	repo := containers.NewKubernetesContainerRepository(clientset)
	containerService = containers.NewContainerService(repo)

	r := mux.NewRouter()

	r.HandleFunc("/container/create", createDeploymentHandler).Methods("POST")
	r.HandleFunc("/container/{name}", getDeploymentStatusHandler).Methods("GET")
	r.HandleFunc("/container/{name}", deleteDeploymentHandler).Methods("DELETE")
	r.HandleFunc("/container/{name}/stop", stopDeploymentHandler).Methods("PATCH")
	r.HandleFunc("/container/{name}/start", startDeploymentHandler).Methods("PATCH")
	r.HandleFunc("/container/{name}/restart", restartDeploymentHandler).Methods("PATCH")

	// Error encountered when trying to pause a deployment: No supported methods in K8 API
	// r.HandleFunc("/container/{name}/pause", pauseDeploymentHandler).Methods("PATCH")

	return r
}

// TODO: Implement user validation instead of using value from request body
func createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	var container containers.Container
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&container)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Assuming the container service is already initialized, create the deployment
	createdDeployment, err := containerService.CreateContainer(
		container.Name,
		container.Owner,
		container.ImageTag,
		container.Replicas,
		container.MappedPort,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(createdDeployment)
}

func getDeploymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Get the container name from the URL path
	var name = mux.Vars(r)["name"]
	var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	if userToken != "valid_token" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "danny"

	container, err := containerService.GetContainerStatus(user, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(container)
}

func deleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Get the container name from the URL path
	var name = mux.Vars(r)["name"]
	var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	if userToken != "valid_token" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "danny"

	err := containerService.DeleteContainer(user, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Container deleted"})
}

// func pauseDeploymentHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
// 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 	w.Header().Set("Content-Type", "application/json")

// 	// Get the container name from the URL path
// 	var name = mux.Vars(r)["name"]
// 	var userToken = r.Header.Get("Authorization")

// 	// TODO: Validate the user has access to the container (zero trust)
// 	// Placeholder for user validation
// 	if userToken != "valid_token" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// TODO: Get user from token
// 	// Placeholder for user extraction from token
// 	// user, err := getUserFromToken(userToken)
// 	// if err != nil {
// 	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 	// 	return
// 	// }
// 	var user = "danny"

// 	err := containerService.PauseContainer(user, name)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusAccepted)

// 	json.NewEncoder(w).Encode(map[string]string{"message": "Container paused"})
// }

func stopDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Get the container name from the URL path
	var name = mux.Vars(r)["name"]
	var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	if userToken != "valid_token" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "danny"

	err := containerService.StopContainer(user, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func startDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Get the container name from the URL path
	var name = mux.Vars(r)["name"]
	var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	if userToken != "valid_token" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "danny"

	err := containerService.StartContainer(user, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// Restart the deployment by stopping and starting it
func restartDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Get the container name from the URL path
	var name = mux.Vars(r)["name"]
	var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	if userToken != "valid_token" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "danny"

	err := containerService.StopContainer(user, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Wait for 5 seconds before starting the container for nice restart effect
	time.Sleep(5 * time.Second)

	err = containerService.StartContainer(user, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
