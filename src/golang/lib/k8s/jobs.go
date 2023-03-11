package k8s

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/errors"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ErrNoPodExists() error {
	return errors.New("No pod exists")
}

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
	namespace := AqueductNamespace
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

	// This also means parallelism == 1, and the success of this one pod means the success of the job.
	numCompletions := int32(1)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSeconds,
			Completions:             &numCompletions,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,

					// We label each pod with the job name, so we can query for it later (when polling).
					// This is a valid assumption, only because we spawn one pod per job.
					Labels: map[string]string{
						"job-name": name,
					},
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

func GetJob(ctx context.Context, name string, k8sClient *kubernetes.Clientset) (*batchv1.Job, error) {
	// Currently, all jobs run workflow operators, which should be in the user namespace.
	namespace := AqueductNamespace

	return k8sClient.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func GetPod(ctx context.Context, name string, k8sClient *kubernetes.Clientset) (*corev1.Pod, error) {
	namespace := AqueductNamespace

	podList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", name),
	})
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		log.Infof("No pod has been created from job %s yet...", name)
		return nil, ErrNoPodExists()
	}

	if len(podList.Items) != 1 {
		return nil, errors.Newf("Expected job %s to have one pod, but instead got %v.", name, len(podList.Items))
	}
	return &podList.Items[0], nil
}
