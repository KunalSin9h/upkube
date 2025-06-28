package api

import (
	"github.com/pkg/errors"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type ServerConfig struct {
	Host      string
	Port      string
	Env       string
	ClientSet *kubernetes.Clientset
}

type ServerConfigFunc func(cfg *ServerConfig)

func WithHost(host string) ServerConfigFunc {
	return func(config *ServerConfig) {
		config.Host = host
	}
}

func WithPort(port string) ServerConfigFunc {
	return func(config *ServerConfig) {
		config.Port = port
	}
}

func WithEnv(env string) ServerConfigFunc {
	return func(config *ServerConfig) {
		config.Env = env
	}
}

func NewServiceConfig(clientSet *kubernetes.Clientset, funcs ...ServerConfigFunc) *ServerConfig {
	config := &ServerConfig{
		ClientSet: clientSet,
	}

	for _, fn := range funcs {
		fn(config)
	}

	return config
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
