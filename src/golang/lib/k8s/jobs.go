package k8s

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib"
	"github.com/aqueducthq/aqueduct/lib/container_registry"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ErrNoPodExists = errors.New("No pod exists")

// A helper function that takes in the name of a job, a container image, and
// other configuration parameters. It uses this information to generate a new
// job and run the job.
func LaunchJob(
	name, containerImage string,
	environmentVariables *map[string]string,
	secretEnvVariables []string,
	resourceRequests *map[string]string,
	image *operator.ImageConfig,
	k8sClient *kubernetes.Clientset,
) error {
	// Currently, all jobs run workflow operators, which should be in the user namespace.
	namespace := AqueductNamespace
	privileged := false

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

	var imagePullSecretName string

	if image != nil {
		(*environmentVariables)["AQUEDUCT_EXPECTED_VERSION"] = lib.ServerVersionNumber

		// Get ECR credentials to authenticate with ECR registry
		if image.Service != shared.ECR {
			return errors.Newf("Unsupported image service: %s", image.Service)
		}

		registryID, err := uuid.Parse(*image.RegistryID)
		if err != nil {
			return errors.Wrap(err, "Unable to parse container registry ID.")
		}

		storageConfig := config.Storage()
		vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
		if err != nil {
			return errors.Wrap(err, "Unable to initialize vault.")
		}

		config, err := auth.ReadConfigFromSecret(context.Background(), registryID, vaultObject)
		if err != nil {
			return errors.Wrap(err, "Unable to read container registry config from vault.")
		}

		ecrConfig, err := container_registry.UpdateECRCredentialsIfNeeded(config, registryID, vaultObject)
		if err != nil {
			return errors.Wrap(err, "Unable to get ECR config.")
		}

		uid, err := uuid.NewUUID()
		if err != nil {
			return errors.Wrap(err, "Unable to generate UUID for Kubernetes image pull secret name.")
		}

		imagePullSecretName = uid.String()

		authConfig := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      imagePullSecretName,
				Namespace: namespace,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				".dockerconfigjson": []byte(fmt.Sprintf(`{"auths": {"%s": {"username": "AWS", "password": "%s", "email": "none"}}}`, ecrConfig.ProxyEndpoint, ecrConfig.Token)),
			},
		}

		// Create a secret with ECR credentials
		_, err = k8sClient.CoreV1().Secrets(namespace).Create(context.Background(), authConfig, metav1.CreateOptions{})
		if err != nil {
			// Double-check that we didn't race against another process to create this secret.
			if _, secretExistsErr := GetSecret(context.Background(), imagePullSecretName, k8sClient); secretExistsErr != nil {
				return errors.Wrap(err, "Error while creating ECR Secrets")
			}
		}
	}

	k8sEnvironmentVariables, resourceRequirements := generateK8sEnvVarAndResourceReq(environmentVariables, resourceRequests)

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
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if imagePullSecretName != "" {
		job.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: imagePullSecretName}}
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

func DeleteJob(ctx context.Context, name string, k8sClient *kubernetes.Clientset) error {
	// Currently, all jobs run workflow operators, which should be in the user namespace.
	namespace := AqueductNamespace

	backgroundDeletion := metav1.DeletePropagationBackground
	return k8sClient.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &backgroundDeletion})
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
		return nil, ErrNoPodExists
	}

	if len(podList.Items) != 1 {
		return nil, errors.Newf("Expected job %s to have one pod, but instead got %v.", name, len(podList.Items))
	}
	return &podList.Items[0], nil
}
