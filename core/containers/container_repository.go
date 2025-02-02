package containers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

// ContainerRepository defines methods for interacting with the cluster.
type ContainerRepository interface {
	Create(container *Container) (*Container, error) // Create a new container
}

// KubernetesContainerRepository is the implementation of ContainerRepository using Kubernetes.
type KubernetesContainerRepository struct {
	CS *kubernetes.Clientset
}

// NewKubernetesContainerRepository creates a new KubernetesContainerRepository
func NewKubernetesContainerRepository(clientset *kubernetes.Clientset) *KubernetesContainerRepository {
	return &KubernetesContainerRepository{CS: clientset}
}

// Save stores a user in the database
func (cluster *KubernetesContainerRepository) Create(container *Container) (*Container, error) {

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: container.UserName,
		},
	}

	namespacesClient := cluster.CS.CoreV1().Namespaces()

	// Check if the namespace already exists
	_, err := namespacesClient.Get(context.Background(), namespace.Name, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("Namespace %q already exists.\n", namespace.Name)
	} else {
		// Create the namespace
		fmt.Println("Creating namespace...")
		_, err = namespacesClient.Create(context.Background(), namespace, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create namespace: %v", err)
		}
		fmt.Printf("Created namespace %q.\n", namespace.Name)
	}

	// Name - deployment?
	// Labels
	// Name for the container - test
	// Port - service to be attached

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "test",
							Image: container.ContainerTag,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: container.MappedPort,
									Name:          "test-port",
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentClient := cluster.CS.AppsV1().Deployments(container.UserName)

	createdDeployment, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-service",
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": "test",
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

	serviceClient := cluster.CS.CoreV1().Services(container.UserName)

	createdService, err := serviceClient.Create(context.Background(), service, metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	container.CreationDate = createdDeployment.GetCreationTimestamp().Time
	container.OpenedPort = createdService.Spec.Ports[0].NodePort

	return container, nil
}
