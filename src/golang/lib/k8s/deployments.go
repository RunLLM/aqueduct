package k8s

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

type ScaleDirection int

const (
	ScaleUp ScaleDirection = iota
	ScaleDown
	DeploymentDeletionInterval = 5
	DeploymentDeletionTimeout  = 60
	WaitForDeploymentInterval  = 5  // seconds
	WaitForDeploymentTimeout   = 10 // minutes
)

//  This function generates a Kubernetes deployment, which can have a dynamic number of pods
//	deployed in it. We start with `numReplicas` pods in the deployment.
func CreateDeployment(
	numReplicas int32,
	name, containerImage string,
	environmentVariables *map[string]string,
	secretEnvVariables []string,
	resourceRequests *map[string]string,
	labels *map[string]string,
	systemPod bool,
	serviceAccount string,
	privileged bool,
	k8sClient *kubernetes.Clientset,
	priorityClass string,
	minAvailable int,
	autoscale bool,
	enableProbing bool,
	ports *map[uint32]uint32,
) error {
	probingPort := getProbingPort(ports)

	pod, _ := generatePod(
		name,
		containerImage,
		environmentVariables,
		secretEnvVariables,
		resourceRequests,
		labels,
		systemPod,
		serviceAccount,
		privileged,
		priorityClass,
		enableProbing,
		probingPort,
	)

	// Any potential `CreateOptions` fields we might want to add would go here.
	// But for now, it's just empty.
	createOptions := metav1.CreateOptions{}

	deployment := &appsv1.Deployment{
		// metav1.ObjectMeta specifies the name of the pod being created, the
		// namespace, and any arbitrary labels.
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: pod.ObjectMeta.Namespace,
			Labels: map[string]string{
				ServiceKey: name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &numReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ServiceKey: name,
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-spec", name),
					Namespace: pod.ObjectMeta.Namespace,
					Labels:    pod.ObjectMeta.Labels,
				},
				Spec: pod.Spec,
			},
		},
	}

	_, err := k8sClient.AppsV1().Deployments(pod.ObjectMeta.Namespace).Create(context.TODO(), deployment, createOptions)
	if err != nil {
		return err
	}
	if minAvailable > 0 {
		err = CreatePodDisruptionBudget(minAvailable, k8sClient, name, systemPod)
	}
	if err != nil {
		return err
	}
	if !autoscale {
		return nil
	}
	return CreateHorizontalPodAutoscaler(name, pod.ObjectMeta.Namespace, k8sClient)
}

// This function simply returns a internal port that can be used to probe pod healthiness
// If no port is provided, it assumes all port can be used to do healthiness check and returns
// a hard-coded default value.
func getProbingPort(ports *map[uint32]uint32) uint32 {
	for _, internalPort := range *ports {
		return internalPort
	}
	// map is empty
	return DefaultProbingPort
}

//	This is a helper function that takes the name of a service that has been
//	previously deployed and returns a list of IP addresses (which are internal
//	to the Kubernetes cluster) on which the service can be accessed.
func GetDeploymentPodIps(k8sClient *kubernetes.Clientset, serviceName string, systemPod bool) ([]string, error) {
	result := []string{}

	namespace := getDeploymentNamespace(systemPod)
	pods, err := getServicePod(context.TODO(), namespace, serviceName, k8sClient)
	if err != nil {
		log.Errorf("Unexpected error while retrieving service pod addresses for service %s:\n%v", serviceName, err)
		return result, err
	}

	if len(pods.Items) == 0 {
		return result, errors.New("there were no available resources for the service/")
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning { // This pod is not currently actively serving requests.
			continue
		}

		for _, podCondition := range pod.Status.Conditions {
			if podCondition.Type == corev1.ContainersReady && podCondition.Status == corev1.ConditionTrue {
				result = append(result, pod.Status.PodIP)
			}
		}
	}

	return result, nil
}

// This function deletes a k8s deployment and its associated PodDisruptionBudget.
// TODO: https://www.notion.so/aqueducthq/Implement-deleting-HorizontalPodAutoscaler-in-DeleteDeployment-c0afa6c736a943e3a7b6078251e639f3
// TODO: Delete HorizontalPodAutoscaler for deployments that have autoscaling enabled
func DeleteDeployment(
	name string,
	systemPod bool,
	k8sClient *kubernetes.Clientset,
) error {
	namespace := getDeploymentNamespace(systemPod)

	err := k8sClient.AppsV1().Deployments(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = DeletePodDisruptionBudget(k8sClient, name, namespace)
	if err != nil {
		return err
	}

	return DeleteHorizontalPodAutoscaler(namespace, name, k8sClient)
}

// This function relies on the fact that every service that is deployed onto
// the system has a label of the form serviceName={name-of-service}. This
// allow us to list all of the pods that have that label and retrieve their
// IP addresses.
func getServicePod(ctx context.Context, namespace, serviceName string, k8sClient *kubernetes.Clientset) (*corev1.PodList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", ServiceKey, serviceName),
	}

	return k8sClient.CoreV1().Pods(namespace).List(ctx, listOptions)
}

// We immediately delete all the system pods, causing Kubernetes to spin up a new one,
// since our system deployments are part of a replica set and cannot have 0 pods.
func RestartSystemDeployment(
	ctx context.Context,
	serviceName string,
	k8sClient *kubernetes.Clientset,
) error {
	namespace := getDeploymentNamespace(true) // This is always a system pod.
	pods, err := getServicePod(ctx, namespace, serviceName, k8sClient)
	if err != nil {
		return err
	}

	if len(pods.Items) < 1 {
		return errors.Newf("No %s pods detected!", serviceName)
	}

	// We want this pod to be deleted immediately, without any draining.
	gracePeriod := int64(0)
	for _, podToDelete := range pods.Items {
		log.Infof("Deleting pod %v", podToDelete.Name)
		err = k8sClient.CoreV1().Pods(namespace).Delete(ctx, podToDelete.Name, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
		if err != nil {
			return errors.Wrap(err, "Failed to delete pod")
		}
	}
	return nil
}

// Waits until all pods in the deployment are ready to service requests.
// The deployment must also have at least one pod.
func WaitForSystemDeploymentReady(
	ctx context.Context,
	serviceName string,
	k8sClient *kubernetes.Clientset,
) error {
	fmt.Printf("Waiting for %s to be ready.\n", serviceName)

	readyConditionFunction := func() (bool, error) {
		namespace := getDeploymentNamespace(true) // This is always a system pod.
		pods, err := getServicePod(ctx, namespace, serviceName, k8sClient)
		if err != nil {
			return false, err
		}

		if len(pods.Items) < 1 {
			return false, nil
		}

		numReady := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodPending {
				continue
			}
			if pod.Status.Phase != corev1.PodRunning {
				return false, errors.Newf("Service %s has a pod with status %s", serviceName, pod.Status.Phase)
			}

			for _, podCondition := range pod.Status.Conditions {
				if podCondition.Type == corev1.PodReady && podCondition.Status == corev1.ConditionTrue {
					numReady++
				}
			}
		}

		// Only break the loop if all pods are ready.
		return numReady == len(pods.Items), nil
	}

	err := wait.PollImmediate(
		WaitForDeploymentInterval*time.Second,
		WaitForDeploymentTimeout*time.Minute,
		readyConditionFunction,
	)
	if err != nil {
		return err
	}

	fmt.Printf("%s is ready.\n", serviceName)
	return nil
}

//	This is a helper function that deletes all of the deployments in a
//	particular namespace, specified by the `namespace` argument. It uses the
//	Kubernetes client API to list all of the deployments in a particular
//	namespace and then delete them one by one.
func DeleteDeploymentsByNamespace(k8sClient *kubernetes.Clientset, namespace string, exceptions map[string]bool) {
	DeleteHorizontalPodAutoscalerByNamespace(namespace, k8sClient)
	deploymentsClient := k8sClient.AppsV1().Deployments(namespace)
	deployments, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{}) // List all of the deployments.
	if err != nil {
		log.Errorf("Unexpected error while retrieving deployments in namespace %s: %v\n", namespace, err)
		return
	}

	for _, deployment := range deployments.Items {
		name := deployment.ObjectMeta.Name
		if _, ok := exceptions[name]; ok { // If this deployment is an exception, we leave it running.
			continue
		}

		err := deploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Unable to delete %s deployment: %s\n", namespace, name)
			continue
		}

		deletionConditionFunction := func() (bool, error) {
			_, err := deploymentsClient.Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				return true, nil // meaning that the deployment is already deleted, which is what we want.
			}
			log.Infof("Waiting for deployment %s to be deleted...\n", name)
			return false, nil
		}

		// Wait for the deployment that we just deleted to actually go away.
		err = wait.PollImmediate(DeploymentDeletionInterval*time.Second, DeploymentDeletionTimeout*time.Second, deletionConditionFunction)
		if err != nil {
			log.Errorf("Error while waiting for deployment to be deleted: %v\n", err)
			continue
		}
	}
	log.Infof("Successfully deleted deployments in namespace %s\n", namespace)
}

func getDeploymentNamespace(systemPod bool) string {
	if systemPod {
		return SystemNamespace
	} else {
		return UserNamespace
	}
}

func SetDeploymentReplicas(
	k8sClient *kubernetes.Clientset,
	name string,
	systemPod bool,
	count int32,
) (int32, error) {
	namespace := getDeploymentNamespace(systemPod)

	scale, err := k8sClient.AppsV1().Deployments(namespace).GetScale(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}
	newCount := int32(math.Max(0, float64(scale.Spec.Replicas+count)))
	scale.Spec.Replicas = newCount

	_, err = k8sClient.AppsV1().Deployments(namespace).UpdateScale(context.TODO(), name, scale, metav1.UpdateOptions{})
	return newCount, err
}

func WaitForReplicas(
	k8sClient *kubernetes.Clientset,
	name string,
	systemPod bool,
	desiredCount int,
	direction ScaleDirection,
	interval,
	timeout time.Duration,
) error {
	return wait.PollImmediate(interval, timeout, checkReplicaCountCondition(k8sClient, name, systemPod, desiredCount, direction))
}

func checkReplicaCountCondition(k8sClient *kubernetes.Clientset, name string, systemPod bool, desiredCount int, direction ScaleDirection) wait.ConditionFunc {
	return func() (bool, error) {
		runningPods, err := GetDeploymentPodIps(k8sClient, name, systemPod)
		if err != nil {
			return false, err
		}
		if direction == ScaleUp {
			return len(runningPods) >= desiredCount, nil
		}
		return len(runningPods) <= desiredCount, nil
	}
}
