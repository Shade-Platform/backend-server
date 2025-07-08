package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shade_web_server/core/containers"
	"shade_web_server/infrastructure/logger"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Container Service handling all container operations
var containerService *containers.ContainerService

func InitializeContainersRouter(clientset *kubernetes.Clientset, metrics *metrics.Clientset) *mux.Router {
	repo := containers.NewKubernetesContainerRepository(clientset, metrics)
	containerService = containers.NewContainerService(repo)

	r := mux.NewRouter()

	r.HandleFunc("/container/create", createDeploymentHandler).Methods("POST")
	r.HandleFunc("/container/{name}", getDeploymentStatusHandler).Methods("GET")
	r.HandleFunc("/container/delete", deleteDeploymentHandler).Methods("DELETE")
	r.HandleFunc("/container/{name}/stop", stopDeploymentHandler).Methods("PATCH")
	r.HandleFunc("/container/{name}/start", startDeploymentHandler).Methods("PATCH")
	r.HandleFunc("/container/{name}/restart", restartDeploymentHandler).Methods("PATCH")
	r.HandleFunc("/container/namespace/{name}", getDeploymentsByNamespace).Methods("GET")
	r.HandleFunc("/container/metrics", getContainerMetricsHandler).Methods("POST")

	return r
}

func getDeploymentsByNamespace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	var namespace = mux.Vars(r)["name"]

	containers, err := containerService.ContainerRepo.GetAllByNamespace(namespace)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event": "deployments_error",
			"user":  namespace,
			"error": err.Error(),
		}).Error("Failed to get containers belonging to user")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()

	if !query.Has("hours") {
		// Encode containers as JSON
		response, err := json.Marshal(map[string]any{
			"containers": containers,
		})
		if err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"event": "deployments_today_error",
				"user":  namespace,
				"error": err.Error(),
			}).Error("Failed to encode json response")
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		w.Write(response)
		return
	}

	hours, err := strconv.Atoi(query.Get("hours"))

	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event": "deployments_today_error",
			"user":  namespace,
			"error": err.Error(),
		}).Error("Failed to encode json response")
		http.Error(w, "Invalid query parameter hours", http.StatusBadRequest)
		return
	}

	today := 0
	for _, container := range containers {
		if time.Since(container.CreationDate) <= time.Duration(hours)*time.Hour {
			today += 1
		}
	}

	response, err := json.Marshal(map[string]interface{}{
		"deployments": today,
		"hours":       hours,
	})

	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event": "deployments_today_error",
			"user":  namespace,
			"error": err.Error(),
		}).Error("Failed to marshal deployments today response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(response)
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

	logger.Log.WithFields(map[string]interface{}{
		"event":     "container_creation_attempt",
		"container": container.Name,
		"image":     container.ImageTag,
		"ip":        r.RemoteAddr,
		"method":    r.Method,
		"path":      r.URL.Path,
	}).Info("Container creation request")

	// Assuming the container service is already initialized, create the deployment
	createdDeployment, err := containerService.CreateContainer(
		container.Name,
		container.Owner,
		container.ImageTag,
		container.Replicas,
		container.MappedPort,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
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
	// var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	// if userToken != "valid_token" {
	// http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// return
	// }

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "test"

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

	var deleteDeploymentRequest struct {
		ContainerName string `json:"name"`
		User          string `json:"userID"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&deleteDeploymentRequest)

	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event": "container_deletion_request_error",
			"error": err.Error(),
		}).Error("Failed to decode request into json")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = containerService.DeleteContainer(deleteDeploymentRequest.User, deleteDeploymentRequest.ContainerName)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event":          "container_deletion_error",
			"container_name": deleteDeploymentRequest.ContainerName,
			"user_id":        deleteDeploymentRequest.User,
			"error":          err.Error(),
		}).Error("Failed to decode request into json")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	logger.Log.WithFields(map[string]interface{}{
		"event":          "container_deletion_success",
		"container_name": deleteDeploymentRequest.ContainerName,
		"user_id":        deleteDeploymentRequest.User,
	}).Error("Container Deleted Successfully")

	json.NewEncoder(w).Encode(map[string]string{"message": "Container deleted"})
}

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
	// var userToken = r.Header.Get("Authorization")

	// TODO: Validate the user has access to the container (zero trust)
	// Placeholder for user validation
	// if userToken != "valid_token" {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	// TODO: Get user from token
	// Placeholder for user extraction from token
	// user, err := getUserFromToken(userToken)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	var user = "test"

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

func getContainerMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	var containerMetricsRequest struct {
		ContainerName string `json:"name"`
		User          string `json:"userID"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&containerMetricsRequest)

	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"event": "container_metrics_request_error",
			"error": err.Error(),
		}).Error("Failed to decode request into json")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	metrics, err := containerService.ContainerRepo.GetMetrics(containerMetricsRequest.User, containerMetricsRequest.ContainerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}
