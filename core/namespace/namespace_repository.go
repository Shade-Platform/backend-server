package namespace

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesNamespaceRepository struct {
	CS *kubernetes.Clientset
}

// NewKubernetesContainerRepository creates a new KubernetesContainerRepository
func NewKubernetesNamespaceRepository(clientset *kubernetes.Clientset) KubernetesNamespaceRepository {
	return KubernetesNamespaceRepository{CS: clientset}
}

type NamespaceRepository interface {
	CreateNamespace(name string) error
	Exists(name string) (bool, error)
}

func (repo KubernetesNamespaceRepository) CreateNamespace(name string) error {
	namespaceClient := repo.CS.CoreV1().Namespaces()

	// First ensure no namespace already has that name
	_, err := namespaceClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err = namespaceClient.Create(context.Background(), namespace, metav1.CreateOptions{})

	return err
}

func (repo KubernetesNamespaceRepository) Exists(name string) (bool, error) {
	namespaceClient := repo.CS.CoreV1().Namespaces()

	_, err := namespaceClient.Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	return true, nil
}
