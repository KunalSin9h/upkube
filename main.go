package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kunalsin9h/upkube/internal/kubeapi"
	"github.com/kunalsin9h/upkube/views"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	HOST = "localhost"
	PORT = "8080"
)

func init() {
	if os.Getenv("PORT") != "" {
		PORT = os.Getenv("PORT")
	}

	if os.Getenv("HOST") != "" {
		HOST = os.Getenv("HOST")
	}
}

func main() {
	// Create in-cluster config
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// Use local kubeconfig
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// namespace := "default"
	// deploymentName := "my-deployment"

	// Restart Deployment
	// restartDeployment(clientset, namespace, deploymentName)

	// Update Deployment Image
	// updateDeploymentImage(clientset, namespace, deploymentName, "nginx:1.21")

	ctx := context.Background()

	// List Deployments in a namespace (e.g., "default")
	deployments, err := clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, deploy := range deployments.Items {
		fmt.Printf("Deployment Name: %s\n", deploy.Name)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, ns := range namespaces.Items {
		fmt.Println(ns.Name)
	}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Pass the request or extracted info to your component
		userEmail := r.Header.Get("Cf-Access-Authenticated-User-Email")
		// If your Root templ expects userEmail:

		// TODO: unauthorized page.
		// if userEmail == "" || true {
		// 	// Not authenticated or header missing
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		userEmail = "kunal.singh@safedep.io"

		namespace := r.URL.Query().Get("namespace")
		if namespace == "" {
			namespace = "default"
		}

		root := views.Root(userEmail, clientset, namespace)
		root.Render(r.Context(), w)
	})

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("POST /restart", func(w http.ResponseWriter, r *http.Request) {
		namespace := r.FormValue("namespace")
		deployment := r.FormValue("deployment")
		if namespace == "" || deployment == "" {
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}
		fmt.Println(namespace, deployment)
		// TODO: Send some notification to the user.
		go kubeapi.RestartDeployment(clientset, namespace, deployment)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	fmt.Println("Listening on http://" + HOST + ":" + PORT)
	http.ListenAndServe(HOST+":"+PORT, nil)
}
