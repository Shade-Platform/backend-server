package infrastructure

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewClusterConnection() (*kubernetes.Clientset, error) {
	// create in cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	_clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return _clientset, nil
}
