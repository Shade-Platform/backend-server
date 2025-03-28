package infrastructure

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewClusterConnection() (*kubernetes.Clientset, error) {

	// out of cluster config
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	// create in cluster config
	// config, err := rest.InClusterConfig()
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
