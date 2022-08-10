package k8s

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateSecret(
	ctx context.Context,
	name string,
	secrets map[string]string,
	k8sClient *kubernetes.Clientset,
) error {
	// Convert the values of `secrets` to type []byte.
	castedSecrets := map[string][]byte{}
	for key := range secrets {
		castedSecrets[key] = []byte(secrets[key])
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: UserNamespace,
		},
		Data: castedSecrets,
	}

	// Call 'create' or 'update' according to whether if a secret already exists
	// TODO (likawind): use proper context and options
	// https://www.notion.so/aqueducthq/Use-proper-context-and-options-33da1baeb12144a2a2381c641a44cf7c
	_, err := GetSecret(ctx, secret.ObjectMeta.Name, k8sClient)
	// The secret doesn't exist
	if err != nil {
		_, err := k8sClient.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(ctx, &secret, metav1.CreateOptions{})
		return err
	}

	// Already exists
	_, err = k8sClient.CoreV1().Secrets(secret.ObjectMeta.Namespace).Update(ctx, &secret, metav1.UpdateOptions{})
	return err
}

func GetSecret(ctx context.Context, name string, k8sClient *kubernetes.Clientset) (map[string]string, error) {
	secret, err := k8sClient.CoreV1().Secrets(UserNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secretMap := make(map[string]string, len(secret.Data))
	for key, castedSecret := range secret.Data {
		secretMap[key] = string(castedSecret)
	}
	return secretMap, nil
}

func DeleteSecret(ctx context.Context, name string, k8sClient *kubernetes.Clientset) error {
	return k8sClient.CoreV1().Secrets(UserNamespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func DeleteSecretsByNamespace(ctx context.Context, k8sClient *kubernetes.Clientset, namespace string) {
	// Don't delete service account token secrets, as those are managed by service account deletion
	// Don't delete docker config secrets, as those are managed by `DeleteDockerSecret` in lib/k8s/utils
	fieldSelector := createExclusiveSecretFieldSelector(corev1.SecretTypeServiceAccountToken, corev1.SecretTypeDockerConfigJson)
	listOptions := metav1.ListOptions{FieldSelector: fieldSelector}

	err := k8sClient.CoreV1().Secrets(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, listOptions)
	if err != nil {
		log.Errorf("Unexpected error while deleting secrets: %v", err)
	} else {
		log.Infof("Successfully deleted secrets in namespace: %s", namespace)
	}
}

// Helper function that returns string representing a k8s field selector
// that excludes secrets of the types listed in `secretTypes`
// K8s syntax for field selectors can be found at: https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
func createExclusiveSecretFieldSelector(secretTypes ...corev1.SecretType) string {
	selectors := make([]string, 0, len(secretTypes))
	for _, secretType := range secretTypes {
		selectors = append(selectors, fmt.Sprintf("type!=%s", secretType))
	}
	// Commas are used to chain field selectors
	return strings.Join(selectors, ",")
}
