package kubeapi

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewClientSet(env string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if strings.EqualFold(env, "PROD") {
		// Create in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Println("Failed to create in-cluster config:", err)
			return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
		}
	} else {
		// Use local kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Println("Failed to create config from kubeconfig file:", err)
			return nil, fmt.Errorf("failed to create config from kubeconfig file: %v", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Failed to create clientset:", err)
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}

func GetAllNameSpaces(clientset *kubernetes.Clientset) ([]string, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list namespaces: %v\n", err)
		//return nil, fmt.Errorf("failed to list namespaces: %v", err)
		// Hypothesis: is service account is in default namespaces, we might not access other namespaces
		return []string{"default"}, nil // Return default namespace if listing fails
	}

	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	return namespaceNames, nil
}

func ListDeployments(clientset *kubernetes.Clientset, namespace string) (*v1.DeploymentList, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to list deployments in namespace", namespace, ":", err)
		return nil, fmt.Errorf("failed to list deployments in namespace %s: %v", namespace, err)
	}

	return deployments, nil
}

func RestartDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string) error {
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
		fmt.Println("Failed to restart deployment in namespace", namespace, ":", retryErr)
		return fmt.Errorf("Update failed: %v", retryErr)
	}

	return nil
}

func UpdateDeploymentImage(clientset *kubernetes.Clientset, namespace, deploymentName, newImage string) error {
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
		fmt.Println("Failed to update deployment image in namespace", namespace, ":", retryErr)
		return fmt.Errorf("Update failed: %v", retryErr)
	}

	return nil
}

// GetImagePullError returns the first image pull error reason and message for pods in a deployment, or empty string if none
func GetDeploymentImageError(clientset *kubernetes.Clientset, namespace, deploymentName string) (string, string, error) {
	// List pods with the deployment's label selector
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	selector := deployment.Spec.Selector.MatchLabels
	labelSelector := []string{}
	for k, v := range selector {
		labelSelector = append(labelSelector, fmt.Sprintf("%s=%s", k, v))
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return "", "", err
	}
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "CrashLoopBackOff" {
					return reason, cs.State.Waiting.Message, nil
				}
			}
		}
	}
	return "", "", nil
}
