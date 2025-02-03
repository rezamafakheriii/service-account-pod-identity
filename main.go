package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/rezamafakheriii/pod-identity-agent/lib" // Replace with the actual import path for your library
)

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopCh := make(chan struct{})
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		log.Println("Received termination signal. Shutting down...")
		close(stopCh)
	}()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "sample-app"
	}

	coreClient := clientset.CoreV1()
	controller := lib.NewServiceAccountController(coreClient, namespace)

	// Start the controller
	log.Printf("Starting ServiceAccountController in namespace: %s\n", namespace)
	controller.Run(stopCh)

	// Block until the stopCh is closed
	<-stopCh
	log.Println("ServiceAccountController stopped.")
}
