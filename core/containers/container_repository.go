package containers

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/utils/ptr"
)

// KubernetesContainerRepository is the implementation of ContainerRepository using Kubernetes.
type KubernetesContainerRepository struct {
	CS *kubernetes.Clientset
	M  *metrics.Clientset
}

// NewKubernetesContainerRepository creates a new KubernetesContainerRepository
func NewKubernetesContainerRepository(clientset *kubernetes.Clientset, metrics *metrics.Clientset) KubernetesContainerRepository {
	return KubernetesContainerRepository{
		CS: clientset,
		M:  metrics,
	}
}

type ContainerMetrics struct {
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
}

// Deployment (unless AddPods, Remove Pods)
// ContainerRepository defines methods for interacting with the cluster.
type ContainerRepository interface {
	GetByName(namespace, name string) (*Container, error) // Get a container from namespace and name
	Create(container *Container) (*Container, error)      // Create a new container
	Delete(namespace, name string) error                  // Remove a container from the cluster
	Stop(namespace, name string) error                    // Stop container and destroy state
	Start(namespace, name string) error                   // Start a stopped container
	GetAllByNamespace(namespace string) ([]*Container, error)
	GetMetrics(namespace, name string) (*ContainerMetrics, error)

	// Error encountered when trying to pause a deployment: No supported methods in K8 API
	// Pause(namespace, name string) error                   // Pause a container while maintaining state

	// Restart(container *Container) error              // Stop then start a container
	// GetByTagName(tag string) (*[]Container, error) // Get container(s) with a specific tag
	// AddPods(container *Container) error            // Add Pod instances to a container
	// RemovePods(container *Container) error         // Remove Pos instances from a container
}

func (repo KubernetesContainerRepository) GetMetrics(namespace, name string) (*ContainerMetrics, error) {
	deploymentClient := repo.CS.AppsV1().Deployments(namespace)

	deployment, err := deploymentClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}

	selector := metav1.FormatLabelSelector(deployment.Spec.Selector)

	// List pods matching the selector
	podList, err := repo.CS.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}
	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pods found for deployment %s", name)
	}

	// Use the first pod (or aggregate over all pods if you want)
	podName := podList.Items[0].Name

	podMetrics, err := repo.M.MetricsV1beta1().PodMetricses(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var currentCpuUsage, currentMemoryUsage int64 = 0, 0

	for _, c := range podMetrics.Containers {
		currentCpuUsage = c.Usage.Cpu().MilliValue()  // CPU in millicores
		currentMemoryUsage = c.Usage.Memory().Value() // Memory in bytes
	}

	container := deployment.Spec.Template.Spec.Containers[0]

	limitsCPU := float64(0)
	limitsMem := float64(0)

	if cpuLim, ok := container.Resources.Limits["cpu"]; ok {
		limitsCPU = float64(cpuLim.MilliValue())
	}
	if memLim, ok := container.Resources.Limits["memory"]; ok {
		limitsMem = float64(memLim.Value())
	}

	// Avoid division by zero
	var cpuUsage, memUsage float64
	if limitsCPU > 0 {
		cpuUsage = (float64(currentCpuUsage) / limitsCPU) * 100
	}
	if limitsMem > 0 {
		memUsage = (float64(currentMemoryUsage) / limitsMem) * 100
	}

	return &ContainerMetrics{
		CPUUsage:    cpuUsage,
		MemoryUsage: memUsage,
	}, nil
}

func (cluster KubernetesContainerRepository) GetAllByNamespace(namespace string) ([]*Container, error) {
	// Ensure the namespace exists
	_, err := cluster.CS.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("namespace %q does not exist", namespace)
	}

	// Fetch all deployments in the namespace
	deployments, err := cluster.CS.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	var containers []*Container
	for _, deployment := range deployments.Items {
		// Fetch the associated service (optional)
		service, err := cluster.CS.CoreV1().Services(namespace).Get(context.Background(), deployment.Name+"-service", metav1.GetOptions{})
		var nodePort int32
		if err == nil && len(service.Spec.Ports) > 0 {
			nodePort = service.Spec.Ports[0].NodePort
		}

		// Build the container object
		container := &Container{
			Owner:         namespace,
			Name:          deployment.Name,
			ImageTag:      deployment.Spec.Template.Spec.Containers[0].Image,
			Replicas:      *deployment.Spec.Replicas,
			MappedPort:    nodePort,
			CreationDate:  deployment.CreationTimestamp.Time,
			ContainerTags: map[string]string{},
		}

		containers = append(containers, container)
	}

	return containers, nil
}

func (cluster KubernetesContainerRepository) GetByName(namespace, name string) (*Container, error) {

	// Check if the namespace already exists
	namespacesClient := cluster.CS.CoreV1().Namespaces()
	// fmt.Printf("%v", namespacesClient)

	_, err := namespacesClient.Get(context.Background(), namespace, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("Namespace %q in fact exists.\n", namespace)
	} else {
		return nil, fmt.Errorf("namespace %q does not exist", namespace)
	}

	deploymentClient := cluster.CS.AppsV1().Deployments(namespace)

	deployment, err := deploymentClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}

	// TODO: Add option to check if container has ports
	serviceClient := cluster.CS.CoreV1().Services(namespace)

	service, err := serviceClient.Get(context.Background(), name+"-service", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %v", err)
	}

	container := &Container{
		Owner:         namespace,
		Name:          deployment.Name,
		ImageTag:      deployment.Spec.Template.Spec.Containers[0].Image,
		Replicas:      *deployment.Spec.Replicas,
		MappedPort:    service.Spec.Ports[0].NodePort,
		CreationDate:  deployment.GetCreationTimestamp().Time,
		ContainerTags: map[string]string{},
	}

	return container, nil
}

// Creates a deployment with the container attributes
func (cluster KubernetesContainerRepository) Create(container *Container) (*Container, error) {

	// Check if the namespace already exists
	namespacesClient := cluster.CS.CoreV1().Namespaces()

	_, err := namespacesClient.Get(context.Background(), container.Owner, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("Namespace %q already exists.\n", container.Owner)
	} else {
		// Create the namespace if it doesn't exist
		namespace := &apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: container.Owner,
			},
		}

		fmt.Println("Creating namespace...")
		_, err = namespacesClient.Create(context.Background(), namespace, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create namespace: %v", err)
		}
		fmt.Printf("Created namespace %q.\n", namespace.Name)
	}

	// Create the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: container.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &container.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": container.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": container.Name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  container.Name,
							Image: container.ImageTag,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: container.MappedPort,
									Name:          "internal-port",
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("500m"),
									apiv1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("250m"),
									apiv1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentClient := cluster.CS.AppsV1().Deployments(container.Owner)

	createdDeployment, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	// Attach a service to the deployment
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: container.Name + "-service",
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": container.Name,
			},
			Ports: []apiv1.ServicePort{
				{
					Protocol:   apiv1.ProtocolTCP,
					Port:       container.MappedPort,
					TargetPort: intstr.FromInt32(container.MappedPort),
				},
			},
		},
	}

	serviceClient := cluster.CS.CoreV1().Services(container.Owner)

	createdService, err := serviceClient.Create(context.Background(), service, metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	container.CreationDate = createdDeployment.GetCreationTimestamp().Time
	container.MappedPort = createdService.Spec.Ports[0].NodePort

	return container, nil
}

// Deletes a deployment and service
func (cluster KubernetesContainerRepository) Delete(namespace, name string) error {

	// Check if the namespace already exists
	namespacesClient := cluster.CS.CoreV1().Namespaces()

	if _, err := namespacesClient.Get(context.Background(), namespace, metav1.GetOptions{}); err == nil {
		fmt.Printf("Namespace %q in fact exists.\n", namespace)
	} else {
		return fmt.Errorf("namespace %q does not exist", namespace)
	}

	deploymentClient := cluster.CS.AppsV1().Deployments(namespace)

	err := deploymentClient.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %v", err)
	}

	// TODO: Add option to check if container has ports
	serviceClient := cluster.CS.CoreV1().Services(namespace)

	err = serviceClient.Delete(context.Background(), name+"-service", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service: %v", err)
	}

	return nil
}

// Stops a deployment
func (cluster KubernetesContainerRepository) Stop(namespace, name string) error {

	// Check if the namespace already exists
	namespacesClient := cluster.CS.CoreV1().Namespaces()

	if _, err := namespacesClient.Get(context.Background(), namespace, metav1.GetOptions{}); err == nil {
		fmt.Printf("Namespace %q in fact exists.\n", namespace)
	} else {
		return fmt.Errorf("namespace %q does not exist", namespace)
	}

	deploymentClient := cluster.CS.AppsV1().Deployments(namespace)

	deployment, err := deploymentClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}

	// Get the current number of replicas
	replicas := int(*deployment.Spec.Replicas)

	// Store the original number of replicas in an annotation
	deployment.Annotations = map[string]string{
		"original-replicas": strconv.Itoa(replicas),
	}

	deployment.Spec.Replicas = ptr.To(int32(0))

	_, err = deploymentClient.Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	return nil
}

// Starts a stopped deployment
func (cluster KubernetesContainerRepository) Start(namespace, name string) error {

	// Check if the namespace already exists
	namespacesClient := cluster.CS.CoreV1().Namespaces()

	if _, err := namespacesClient.Get(context.Background(), namespace, metav1.GetOptions{}); err == nil {
		fmt.Printf("Namespace %q in fact exists.\n", namespace)
	} else {
		return fmt.Errorf("namespace %q does not exist", namespace)
	}

	deploymentClient := cluster.CS.AppsV1().Deployments(namespace)

	deployment, err := deploymentClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}

	// Get the original number of replicas from the annotation
	originalReplicas, ok := deployment.Annotations["original-replicas"]
	if !ok {
		return fmt.Errorf("original replicas not found in annotations")
	}

	replicas, err := strconv.Atoi(originalReplicas)
	if err != nil {
		return fmt.Errorf("failed to convert replicas to int: %v", err)
	}

	deployment.Spec.Replicas = ptr.To(int32(replicas))

	_, err = deploymentClient.Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	return nil
}
