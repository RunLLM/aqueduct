package k8s

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// A helper function that takes in the name of a job, a container image, and
// other configuration parameters. It uses this information to generate a new
// job and run the job.
func LaunchJob(
	name, containerImage string,
	environmentVariables *map[string]string,
	secretEnvVariables []string,
	resourceRequests *map[string]string,
	k8sClient *kubernetes.Clientset,
) error {
	// Currently, all jobs run workflow operators, which should be in the user namespace.
	namespace := UserNamespace
	privileged := false

	k8sEnvironmentVariables, resourceRequirements := generateK8sEnvVarAndResourceReq(environmentVariables, resourceRequests)

	// This is an empty set of create options because we don't need any of these
	// configurations for now.
	createOptions := metav1.CreateOptions{}

	// This means if the job fails, we won't attempt to restart it.
	backoffLimit := int32(0)

	// This means the job, once completed, will be garbage collected after `ttlSeconds`.
	// Currently set to 3 days.
	ttlSeconds := int32(259200)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSeconds,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           containerImage,
							Env:             k8sEnvironmentVariables,
							Resources:       *resourceRequirements,
							ImagePullPolicy: corev1.PullAlways, // Always update the container if there is a new version.
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
						},
					},
					RestartPolicy:    corev1.RestartPolicyNever,
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: DockerSecretName}},
				},
			},
		},
	}

	if len(secretEnvVariables) > 0 {
		// Assign environment variables from secret references
		job.Spec.Template.Spec.Containers[0].EnvFrom = generateK8sEnvVarFromSecrets(secretEnvVariables)
	}

	_, err := k8sClient.BatchV1().Jobs(job.ObjectMeta.Namespace).Create(context.Background(), &job, createOptions)
	if err != nil {
		return errors.Wrap(err, "Error launching job.")
	}
	return nil
}

func GetJob(name string, k8sClient *kubernetes.Clientset) (*batchv1.Job, error) {
	// Currently, all jobs run workflow operators, which should be in the user namespace.
	namespace := UserNamespace

	return k8sClient.BatchV1().Jobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
}
