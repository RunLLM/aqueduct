package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// A helper function that takes in the name of a pod, a container image, and
// other configuration parameters. It uses this information to generate a new
// pod and deploy that pod to the underlying Kubernetes cluster. Note that
// this method is not being currently used but might be useful for singletons
// to be deployed in the future.
func CreatePod(
	name, containerImage string,
	environmentVariables *map[string]string,
	secretEnvVariables []string,
	resourceRequests *map[string]string,
	labels *map[string]string,
	serviceAccount string,
	privileged bool,
	systemPod bool,
	k8sClient *kubernetes.Clientset,
	priorityClass string,
	enableProbing bool,
	probingPort uint32,
) error {
	pod, createOptions := generatePod(
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
	_, err := k8sClient.CoreV1().Pods(pod.ObjectMeta.Namespace).Create(context.TODO(), pod, createOptions)

	return err
}

// This is a helper function that generates a Kubernetes pod spec based on
// some configuration metadata. This returns the pod to the caller rather than
// creating it (which the `CreatePod` function above can be used for).
func generatePod(
	name, containerImage string,
	environmentVariables *map[string]string,
	secretEnvVariables []string,
	resourceRequests *map[string]string,
	labels *map[string]string,
	systemPod bool,
	serviceAccount string, // The name of a Kubernetes serviceaccount that should be used for this pod.
	privileged bool, // Whether to run this container in privileged mode; only used for Stagehand.
	priorityClass string,
	enableProbing bool,
	probingPort uint32,
) (*corev1.Pod, metav1.CreateOptions) {
	var namespace, podName string
	if systemPod {
		namespace = SystemNamespace
		podName = fmt.Sprintf("%s-system-pod", name)
	} else {
		namespace = UserNamespace
		podName = fmt.Sprintf("%s-user-pod", name)
	}

	k8sEnvironmentVariables, resourceRequirements := generateK8sEnvVarAndResourceReq(environmentVariables, resourceRequests)

	// Add the name of the service as a label to the pod spec.
	(*labels)[ServiceKey] = name

	// This is an empty set of create options because we don't need any of these
	// configurations for now.
	createOptions := metav1.CreateOptions{}

	pod := corev1.Pod{
		// metav1.ObjectMeta specifies the name of the pod being created, the
		// namespace, and any arbitrary labels.
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    *labels,
		},

		// The `PodSpec` contains the interesting stuff about the pod -- what and
		// how many containers to run, where to run them, etc.
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            podName,
					Image:           containerImage,
					Env:             k8sEnvironmentVariables,
					Resources:       *resourceRequirements,
					ImagePullPolicy: corev1.PullAlways, // Always update the container if there is a new version.
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
				},
			},
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: DockerSecretName}},
		},
	}

	if len(secretEnvVariables) > 0 {
		// Assign environment variables from secret references
		pod.Spec.Containers[0].EnvFrom = generateK8sEnvVarFromSecrets(secretEnvVariables)
	}

	if enableProbing {
		// The following probe determines if the container is successfully initialized. It tries to
		// establish connection to the container.
		//
		// The numbers in configs means:
		// Probe starts after StartupProbeInitialDelaySec secs the container starts.
		// It tries to establish connection to the probing port every StartupProbePeriodSec,
		// and once the connection establishes, it's considered
		// as success and the container is considered as successfully started. Otherwise, the
		// probe fails after StartupProbeFailureThreshold tries, or StartupProbeTimeoutSec seconds.
		pod.Spec.Containers[0].StartupProbe = &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(probingPort)},
				},
			},
			InitialDelaySeconds: StartupProbeInitialDelaySec,
			TimeoutSeconds:      StartupProbeTimeoutSec,
			PeriodSeconds:       StartupProbePeriodSec,
			SuccessThreshold:    StartupProbeSuccessThreshold,
			FailureThreshold:    StartupProbeFailureThreshold,
		}
		// The following probe determines if the running container is ready. It tries to
		// establish connection to the container.
		//
		// The numbers in configs means:
		// Probe starts after ReadinessProbeInitialDelaySec secs after the above probe succeeded. It tries to establish connection
		// to the probing port every ReadinessProbePeriodSec, and once the connection establishes, it's considered
		// as success and the container is considered as successfully started.
		//
		// The probe is considered as failure either
		// - Never success after the first ReadinessProbeTimeoutSec and timedout
		// - ReadinessProbeFailureThreshold consecutive failures after a success
		//
		// When the probe failed, it will be removed from the service endpoint
		pod.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(probingPort)},
				},
			},
			InitialDelaySeconds: ReadinessProbeInitialDelaySec,
			TimeoutSeconds:      ReadinessProbeTimeoutSec,
			PeriodSeconds:       ReadinessProbePeriodSec,
			SuccessThreshold:    ReadinessProbeSuccessThreshold,
			FailureThreshold:    ReadinessProbeFailureThreshold,
		}
	}
	podRuntimeClassName := "gvisor"
	pod.Spec.RuntimeClassName = &podRuntimeClassName
	if serviceAccount != "" {
		pod.Spec.ServiceAccountName = serviceAccount
	}

	if priorityClass != "" {
		pod.Spec.PriorityClassName = priorityClass
	}

	return &pod, createOptions
}
