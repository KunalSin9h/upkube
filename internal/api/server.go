package api

import (
	"github.com/pkg/errors"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type ServerConfig struct {
	Host      string
	Port      string
	Env       string // DEV or PROD
	ClientSet *kubernetes.Clientset
}

func NewServiceConfig(host, port, env string, clientSet *kubernetes.Clientset) *ServerConfig {
	return &ServerConfig{
		Host:      host,
		Port:      port,
		Env:       env,
		ClientSet: clientSet,
	}
}

func StartHttpServer(config *ServerConfig) error {
	mux := http.NewServeMux()

	// Heath check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Application endpoints
	mux.HandleFunc("GET /", config.WebHome)
	mux.HandleFunc("POST /restart", config.RestartDeployment)
	mux.HandleFunc("POST /update-image", config.UpdateDeploymentImage)

	err := http.ListenAndServe(config.Host+":"+config.Port, mux)
	if err != nil {
		return errors.Wrap(err, "failed to start server")
	}

	return nil
}
