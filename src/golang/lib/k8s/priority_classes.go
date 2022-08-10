package k8s

import (
	"context"

	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//	A helper function that creates a new priority class to be attached to pods.
func CreatePriorityClass(
	name string,
	priority int,
	globalDefault bool,
	description string,
	k8sClient *kubernetes.Clientset,
) error {
	// This is an empty set of create options because we don't need any of these
	// configurations for now.
	createOptions := metav1.CreateOptions{}

	priorityClass := schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Value:         int32(priority),
		GlobalDefault: globalDefault,
		Description:   description,
	}

	_, err := k8sClient.SchedulingV1().PriorityClasses().Create(context.TODO(), &priorityClass, createOptions)
	return err
}
