package routers

import (
	"encoding/json"
	"net/http"
	"shade_web_server/core/containers"

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
	r.HandleFunc("/container/{id}/status", getDeploymentStatusHandler).Methods("GET")

	return r
}

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

	json.NewEncoder(w).Encode(createdDeployment)
}

func getDeploymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Place holder for fetching deployment status
	json.NewEncoder(w).Encode("Deployment status")
}
