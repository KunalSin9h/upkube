package kubeapi

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
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
			return nil, errors.Wrap(err, "failed to create in-cluster config")
		}
	} else {
		// Use local kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create config from kubeconfig file")
		}
	}

	// Create clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientSet")
	}

	return clientSet, nil
}

func GetAllNameSpaces(clientSet *kubernetes.Clientset) ([]string, error) {
	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Warnf("Failed to list namespaces, permission not allowed. %v", err)
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

func ListDeployments(clientSet *kubernetes.Clientset, namespace string) (*v1.DeploymentList, error) {
	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list deployments in namespace: %s", namespace)
	}

	return deployments, nil
}

func RestartDeployment(clientSet *kubernetes.Clientset, namespace, deploymentName string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, getErr := clientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = map[string]string{}
		}
		deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = fmt.Sprintf("%v", metav1.Now())
		_, updateErr := clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "Failed to restart deployment in namespace %s", namespace)
	}

	return nil
}

func UpdateDeploymentImage(clientSet *kubernetes.Clientset, namespace, deploymentName, newImage string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, getErr := clientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		// Assuming the first container is the one to update
		deployment.Spec.Template.Spec.Containers[0].Image = newImage
		_, updateErr := clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "Failed to update deployment image in namespace: %s", namespace)
	}

	return nil
}

func GetDeploymentImageError(clientSet *kubernetes.Clientset, namespace, deploymentName string) (string, string, error) {
	// List pods with the deployment's label selector
	deployment, err := clientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get deployments")
	}

	selector := deployment.Spec.Selector.MatchLabels

	var labelSelector []string
	for k, v := range selector {
		labelSelector = append(labelSelector, fmt.Sprintf("%s=%s", k, v))
	}

	pods, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: strings.Join(labelSelector, ","),
	})

	if err != nil {
		log.Warn("failed to list pods, permission not granted")
		return "", "", errors.Wrap(err, "failed to list pods for deployment")
	}

	for _, pod := range pods.Items {
		for i, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason

				if reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "CrashLoopBackOff" {
					image := ""

					if i < len(pod.Spec.Containers) {
						image = pod.Spec.Containers[i].Image
					}

					msg := cs.State.Waiting.Message
					if image != "" {
						msg = fmt.Sprintf("Image: %s\n%s", image, msg)
					}

					return reason, msg, nil
				}
			}
		}
	}

	return "", "", nil
}
