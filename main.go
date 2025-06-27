package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kunalsin9h/upkube/internal/api"
	"github.com/kunalsin9h/upkube/internal/kubeapi"
)

var (
	UPKUBE_HOST = "localhost"
	UPKUBE_PORT = "8080"
	UPKUBE_ENV  = "DEV" // or "PROD" based on your environment
)

func init() {
	if os.Getenv("UPKUBE_HOST") != "" {
		UPKUBE_HOST = os.Getenv("UPKUBE_HOST")
	}
	if os.Getenv("UPKUBE_PORT") != "" {
		UPKUBE_PORT = os.Getenv("UPKUBE_PORT")
	}
	if os.Getenv("UPKUBE_ENV") != "" {
		UPKUBE_ENV = os.Getenv("UPKUBE_ENV")
	}
}

func main() {
	clientSet, err := kubeapi.NewClientSet(UPKUBE_ENV)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create Kubernetes clientset: %v", err))
	}

	config := api.NewServiceConfig(UPKUBE_HOST, UPKUBE_PORT, UPKUBE_ENV, clientSet)

	log.Printf("Starting Upkube server on %s:%s in %s environment", config.Host, config.Port, config.Env)
	if err := api.StartHttpServer(config); err != nil {
		log.Fatal(fmt.Errorf("failed to start HTTP server: %v", err))
	}
}
