package api

import (
	"net/http"
	"strings"

	"github.com/kunalsin9h/upkube/internal/kubeapi"
	"github.com/kunalsin9h/upkube/views"
)

func (c *ServerConfig) WebHome(w http.ResponseWriter, r *http.Request) {
	// Pass the request or extracted info to your component
	userEmail := r.Header.Get("Cf-Access-Authenticated-User-Email")

	if userEmail == "" {
		if strings.EqualFold(c.Env, "PROD") {
			// In production, we expect the user to be authenticated
			// Not authenticated or header missing
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userEmail = "dev.user@upkube"
	}

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	root := views.Root(userEmail, c.ClientSet, namespace)
	root.Render(r.Context(), w)
}

func (c *ServerConfig) RestartDeployment(w http.ResponseWriter, r *http.Request) {
	namespace := r.FormValue("namespace")
	deployment := r.FormValue("deployment")
	if namespace == "" || deployment == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}
	// TODO: Send some notification to the user.
	err := kubeapi.RestartDeployment(c.ClientSet, namespace, deployment)
	if err != nil {
		http.Error(w, "Failed to restart deployment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (c *ServerConfig) UpdateDeploymentImage(w http.ResponseWriter, r *http.Request) {
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

	err := kubeapi.UpdateDeploymentImage(c.ClientSet, namespace, deployment, newImage)
	if err != nil {
		http.Error(w, "Failed to update image: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
