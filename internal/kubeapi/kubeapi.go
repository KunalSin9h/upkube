package kubeapi

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RestartDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string) {
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

func UpdateDeploymentImage(clientset *kubernetes.Clientset, namespace, deploymentName, newImage string) {
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
