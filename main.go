package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/kunalsin9h/upkube/views"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"

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
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := "default"
	deploymentName := "my-deployment"

	// Restart Deployment
	restartDeployment(clientset, namespace, deploymentName)

	// Update Deployment Image
	updateDeploymentImage(clientset, namespace, deploymentName, "nginx:1.21")

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Pass the request or extracted info to your component
		userEmail := r.Header.Get("Cf-Access-Authenticated-User-Email")
		// If your Root templ expects userEmail:

		// if userEmail == "" || true {
		// 	// Not authenticated or header missing
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		userEmail = "kunal.singh@safedep.io"

		root := views.Root(userEmail)
		root.Render(r.Context(), w)
	})

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Println("Listening on http://" + HOST + ":" + PORT)
	http.ListenAndServe(HOST+":"+PORT, nil)
}

func restartDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, getErr := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = map[string]string{}
		}
		deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = fmt.Sprintf("%v", metav1.Now())
		_, updateErr := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Deployment restarted successfully.")
}

func updateDeploymentImage(clientset *kubernetes.Clientset, namespace, deploymentName, newImage string) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, getErr := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		// Assuming the first container is the one to update
		deployment.Spec.Template.Spec.Containers[0].Image = newImage
		_, updateErr := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Deployment image updated successfully.")
}
