package kubernetes

import (
	client "k8s.io/client-go/kubernetes"
)

// NewClientWithConfig will create a new client able to talk to a Kubernetes API
// server.
func NewClientWithConfig(config *ClientConfig) (*client.Clientset, error) {
	kubeConfig, err := NewClientConfig(config)
	if err != nil {
		return nil, err
	}

	return client.NewForConfig(kubeConfig)
}
