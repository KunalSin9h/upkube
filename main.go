package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kunalsin9h/upkube/internal/kubeapi"
	"github.com/kunalsin9h/upkube/views"
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

		root := views.Root(userEmail, clientSet, namespace)
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
		err := kubeapi.RestartDeployment(clientSet, namespace, deployment)
		if err != nil {
			http.Error(w, "Failed to restart deployment: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("POST /update-image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		namespace := r.FormValue("namespace")
		deployment := r.FormValue("deployment")
		imagePrefix := r.FormValue("imagePrefix")
		oldTag := r.FormValue("oldTag")
		tag := r.FormValue("tag")

		if namespace == "" || deployment == "" || oldTag == "" || imagePrefix == "" || tag == "" {
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}

		newImage := imagePrefix + ":" + tag

		err := kubeapi.UpdateDeploymentImage(clientSet, namespace, deployment, newImage)
		if err != nil {
			http.Error(w, "Failed to update image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	})

	fmt.Println("Listening on http://" + UPKUBE_HOST + ":" + UPKUBE_PORT)
	http.ListenAndServe(UPKUBE_HOST+":"+UPKUBE_PORT, nil)
}
