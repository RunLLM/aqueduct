package k8s

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// This key is used to create a label for every pod that gives the name of
	// the service that it is running. For user workloads, the value will be
	// the user-specified service name.
	ServiceKey = "serviceName"
)

//	This is a helper function that creates a Kubernetes Clientset using the
//	`client-go/rest` library's `InClusterConfig()` function. This function will
//	only succeed if it is called from within a Kubernetes cluster. Otherwise,
//	the configuration retrieval will fail, and this function will call
//	`log.Fatal` and cause the program to crash. If this function is called from
//	within the cluster, it should not fail.
func CreateClientInCluster() *kubernetes.Clientset {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unexpected error while creating Kubernetes client: %v\n", err)
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Unexpected error while creating Kubernetes client: %v\n", err)
	}

	return k8sClient
}

//	This is a helper function that creates a Kubernetes Clientset using the
//	kubeconfig that is located at `kubecfgLocation`. By default, this location
//	is specified (in `enterprise/config/cluster.yml`) as `~/.kube/config`. Please modify
//	the `cluster.yml` file if you need to change this location. If this
//	location is misconfigured, this function will fail.
func CreateClientOutsideCluster(kubecfgLocation string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubecfgLocation)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error while creating Kubernetes client.")
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error while creating Kubernetes client.")
	}

	return k8sClient, nil
}

//	This is a helper function that creates the two Kubernetes namespaces
//	described above and does not return anything. This function should never
//	fail, so any errors that are encountered call `log.Fatal` and cause the
//	program to crash.
func CreateNamespaces(k8sClient *kubernetes.Clientset) {
	// Check to see if the system namespace has already been created and only
	// create if it doesn't exist.
	namespaces := k8sClient.CoreV1().Namespaces()
	_, err := namespaces.Get(context.TODO(), SystemNamespace, metav1.GetOptions{})
	if err != nil {
		systemNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: SystemNamespace,
			},
			Spec: corev1.NamespaceSpec{}, // The only required field here is a `Finalizers` field that we do not use.
		}

		_, err := namespaces.Create(context.TODO(), systemNamespace, metav1.CreateOptions{})
		if err != nil {
			log.Fatal("Unable to create namespace: ", err)
		}
		log.Infof("System namespace (name: %s) created successfully.\n", SystemNamespace)
	}

	// Create the user pod namespace again only after checking if it exists.
	_, err = namespaces.Get(context.TODO(), UserNamespace, metav1.GetOptions{})
	if err != nil {
		userNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: UserNamespace,
			},
			Spec: corev1.NamespaceSpec{}, // See above for why this is empty.
		}

		_, err = k8sClient.CoreV1().Namespaces().Create(context.TODO(), userNamespace, metav1.CreateOptions{})
		if err != nil {
			log.Fatal("Unable to create namespace: ", err)
		}
		log.Infof("User namespace (name: %s) created successfully.\n", UserNamespace)
	}
}
