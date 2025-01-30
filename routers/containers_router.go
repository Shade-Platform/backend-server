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

	r.HandleFunc("/container", createDeploymentHandler).Methods("POST")

	return r
}

func createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	var container containers.Container

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&container)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Assuming the container service is already initialized, create the deployment
	createdDeployment, err := containerService.CreateContainer(container.UserName, container.ContainerTag)
	if err != nil {
		http.Error(w, "Failed to create container", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdDeployment)
}
