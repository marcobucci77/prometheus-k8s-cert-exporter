package kubeclient

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() (kubernetes.Interface, string, error) {
	// First try to use the Kubernetes ServiceAccount.
	config, err := rest.InClusterConfig()

	if err == nil {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, "", fmt.Errorf("create in-cluster Kubernetes client: %w", err)
		}

		return clientset, "in-cluster", nil
	}

	// If we are not running inside Kubernetes,
	// use the local kubeconfig.
	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, "", fmt.Errorf("find user home directory: %w", err)
		}

		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, "", fmt.Errorf("load kubeconfig %q: %w", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", fmt.Errorf("create Kubernetes client: %w", err)
	}

	return clientset, kubeconfig, nil
}
