package k8s

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SafeToDeleteCluster checks whether there are pods in the aqueduct namespace of the dynamic k8s
// cluster that are in ContainerCreating or Running status. If so, it returns false. Otherwise it
// returns true.
func SafeToDeleteCluster(
	ctx context.Context,
	useSameCluster bool,
	kubeconfigPath string,
) (bool, error) {
	k8sClient, err := CreateK8sClient(kubeconfigPath, useSameCluster)
	if err != nil {
		return false, errors.Wrap(err, "Error while creating K8sClient")
	}

	pods, err := k8sClient.CoreV1().Pods(AqueductNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, errors.Wrap(err, "Error while listing pods in the aqueduct namespace")
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return false, nil
		}

		if pod.Status.Phase == v1.PodPending {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ContainerCreating" {
					return false, nil
				}
			}
		}
	}

	return true, nil
}
