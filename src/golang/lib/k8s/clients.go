package k8s

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateK8sClient(kubeconfigPath string, inCluster bool) (*kubernetes.Clientset, error) {
	if inCluster {
		return CreateClientInCluster()
	} else {
		return CreateClientOutsideCluster(kubeconfigPath)
	}
}

// This is a helper function that creates a Kubernetes Clientset using the
// `client-go/rest` library's `InClusterConfig()` function. This function will
// only succeed if it is called from within a Kubernetes cluster. Otherwise,
// the configuration retrieval will fail, and this function will call
// `log.Fatal` and cause the program to crash. If this function is called from
// within the cluster, it should not fail.
func CreateClientInCluster() (*kubernetes.Clientset, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error while creating in-cluster Kubernetes client.")
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error while creating in-cluster Kubernetes client.")
	}

	return k8sClient, nil
}

// This is a helper function that creates a Kubernetes Clientset using the
// kubeconfig that is located at `kubecfgLocation`.
func CreateClientOutsideCluster(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error while creating Kubernetes client.")
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error while creating Kubernetes client.")
	}

	return k8sClient, nil
}

// This is a helper function that creates the user namespace
// and does not return anything. This function should never
// fail, so any errors that are encountered call `log.Fatal` and cause the
// program to crash.
func CreateNamespaces(k8sClient *kubernetes.Clientset) error {
	namespaces := k8sClient.CoreV1().Namespaces()

	// Create the user pod namespace again only after checking if it exists.
	_, err := namespaces.Get(context.TODO(), AqueductNamespace, metav1.GetOptions{})
	if err != nil {
		userNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: AqueductNamespace,
			},
			Spec: corev1.NamespaceSpec{}, // See above for why this is empty.
		}

		_, err = k8sClient.CoreV1().Namespaces().Create(context.TODO(), userNamespace, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "Unable to create namespace.")
		}
		log.Infof("User namespace (name: %s) created successfully.\n", AqueductNamespace)
	}
	return nil
}
