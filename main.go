/*
Copyright 2025, Kunal Singh

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/kunalsin9h/upkube/internal/api"
	"github.com/kunalsin9h/upkube/internal/kubeapi"
)

var (
	UPKUBE_HOST = "127.0.0.1"
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
	// Version and Go build version info
	fmt.Println(upkubeInfoMessage())

	clientSet, err := kubeapi.NewClientSet(UPKUBE_ENV)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	serverConfig := api.NewServiceConfig(clientSet,
		api.WithHost(UPKUBE_HOST), api.WithPort(UPKUBE_PORT), api.WithEnv(UPKUBE_ENV))

	log.Infof("Starting Upkube server on %s:%s in %s environment", serverConfig.Host, serverConfig.Port, serverConfig.Env)
	if err := api.StartHttpServer(serverConfig); err != nil {
		log.Fatal("failed to start HTTP server: %v", err)
	}
}
