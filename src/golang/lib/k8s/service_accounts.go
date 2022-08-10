package k8s

import (
	"context"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateServiceAccount creates a service account object in the specified namespace with the provided (optional)
// annotations.
func CreateServiceAccount(
	name string,
	namespace string,
	annotations map[string]string,
	k8sClient *kubernetes.Clientset,
) error {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
	if len(annotations) > 0 {
		objectMeta.Annotations = annotations
	}
	serviceAccount := &corev1.ServiceAccount{ObjectMeta: objectMeta}

	_, err := k8sClient.CoreV1().ServiceAccounts(namespace).Create(context.Background(), serviceAccount, metav1.CreateOptions{})
	return err
}

// UpdateServiceAccount updates a service account with new annotations.
func UpdateServiceAccount(
	name string,
	namespace string,
	annotations map[string]string,
	k8sClient *kubernetes.Clientset,
) error {
	serviceAccount, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	serviceAccount.ObjectMeta.Annotations = annotations

	_, err = k8sClient.CoreV1().ServiceAccounts(namespace).Update(context.Background(), serviceAccount, metav1.UpdateOptions{})
	return err
}

// CheckServiceAccountExists returns true if the specified service account exists in
// given namespace; otherwise, it returns false.
func CheckServiceAccountExists(
	name string,
	namespace string,
	k8sClient *kubernetes.Clientset,
) bool {
	serviceAccountsClient := k8sClient.CoreV1().ServiceAccounts(namespace)
	serviceAccountsList, err := serviceAccountsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unexpected error while listing service accounts: %v", err)
		return false
	}

	for _, svcAccount := range serviceAccountsList.Items {
		if svcAccount.Name == name {
			return true
		}
	}
	return false
}
