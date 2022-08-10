package k8s

import (
	"context"

	log "github.com/sirupsen/logrus"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

//	A helper function that creates a new pod disruption budget to be applied to
//	deployments.
func CreatePodDisruptionBudget(
	minAvailable int,
	k8sClient *kubernetes.Clientset,
	name string,
	systemPod bool,
) error {
	// This is an empty set of create options because we don't need any of these
	// configurations for now.
	createOptions := metav1.CreateOptions{}

	castedMinAvailable := intstr.FromInt(minAvailable)

	podDisruptionBudget := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &castedMinAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ServiceKey: name,
				},
			},
		},
	}

	var namespace string
	if systemPod {
		namespace = SystemNamespace
	} else {
		namespace = UserNamespace
	}

	_, err := k8sClient.PolicyV1beta1().PodDisruptionBudgets(namespace).Create(context.TODO(), &podDisruptionBudget, createOptions)
	return err
}

//	This is a helper function that deletes all of the pod disruption budgets in a
//	particular namespace, specified by the `namespace` argument. It uses the
//	Kubernetes client API to list all of the pod disruption budgets in a particular
//	namespace and then delete them one by one.
func DeletePodDisruptionBudgetsByNamespace(k8sClient *kubernetes.Clientset, namespace string, exceptions map[string]bool) {
	podDisruptionBudgetClient := k8sClient.PolicyV1beta1().PodDisruptionBudgets(namespace)
	podDisruptionBudgets, err := podDisruptionBudgetClient.List(context.TODO(), metav1.ListOptions{}) // List all of the pod disruption budgets.
	if err != nil {
		log.Errorf("Unexpected error while retrieving pod disruption budgets in namespace %s: %v\n", namespace, err)
		return
	}

	for _, podDisruptionBudget := range podDisruptionBudgets.Items {
		name := podDisruptionBudget.ObjectMeta.Name
		if _, ok := exceptions[name]; ok { // If this pod disruption budget is an exception, we leave it running.
			continue
		}

		err := podDisruptionBudgetClient.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Unable to delete %s pod disruption budget: %s\n", namespace, name)
			break
		}
	}
	log.Infof("Successfully deleted pod disruption budgets in namespace %s\n", namespace)
}

// Helper function to delete a `PodDisruptionBudget`
func DeletePodDisruptionBudget(k8sClient *kubernetes.Clientset, deploymentName, namespace string) error {
	return k8sClient.PolicyV1beta1().PodDisruptionBudgets(namespace).Delete(context.Background(), deploymentName, metav1.DeleteOptions{})
}
