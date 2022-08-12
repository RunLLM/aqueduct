package k8s

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func generateK8sEnvVarAndResourceReq(environmentVariables *map[string]string, resourceRequests *map[string]string) ([]corev1.EnvVar, *corev1.ResourceRequirements) {
	// Convert from a `map[string]string` to the Kubernetes representation of
	// environment varaibles, which has its own special struct.
	k8sEnvironmentVariables := make([]corev1.EnvVar, len(*environmentVariables))
	index := 0
	for key, value := range *environmentVariables {
		k8sEnvironmentVariables[index] = corev1.EnvVar{
			Name:  key,
			Value: value,
		}
		index += 1
	}

	// Currently, we both request and set the limits for each container to
	// whatever is specified by the client. We might want to change that in the
	// future. Note that we deference the result of the `resource.NewQuantity`
	// call because it returns a pointer, but we do not need that object anywhere
	// else.
	resourceList := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse((*resourceRequests)["cpu"]),
		corev1.ResourceMemory: resource.MustParse((*resourceRequests)["memory"]),
	}

	resourceRequirements := corev1.ResourceRequirements{
		Limits:   resourceList,
		Requests: resourceList,
	}
	return k8sEnvironmentVariables, &resourceRequirements
}

// Helper function to generate k8s environment variable references from secrets.
func generateK8sEnvVarFromSecrets(k8sSecretNames []string) []corev1.EnvFromSource {
	k8sEnvVarRefs := make([]corev1.EnvFromSource, 0, len(k8sSecretNames))
	for _, name := range k8sSecretNames {
		envRef := corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		}
		k8sEnvVarRefs = append(k8sEnvVarRefs, envRef)
	}
	return k8sEnvVarRefs
}

// This function uses `kubectl` to create a secret that contains the Docker Hub credential.
// We could have done it programmatically in Go, but it requires work to manually put all the
// variables into the right data structure and perform base 64 encoding, which is error-prone.
// This way we can get the secret right with one command.
func CreateDockerSecrets(username, password, email, kubeconfigPath string) {
	namespaces := []string{DefaultNamespace, SystemNamespace, UserNamespace}
	for _, namespace := range namespaces {
		err := exec.Command(
			"kubectl",
			"create",
			"secret",
			"docker-registry",
			"-n",
			namespace,
			DockerSecretName,
			fmt.Sprintf("--docker-server=%s", DockerServer),
			fmt.Sprintf("--docker-username=%s", username),
			fmt.Sprintf("--docker-password=%s", password),
			fmt.Sprintf("--docker-email=%s", email),
			"--kubeconfig",
			kubeconfigPath,
		).Run()
		if err != nil {
			log.Errorf("Failed to create the Docker Hub secret in the %s namespace: %v\n", namespace, err)
		}
	}
}

func DeleteDockerSecrets(kubeconfigPath string) {
	namespaces := []string{DefaultNamespace, SystemNamespace, UserNamespace}
	for _, namespace := range namespaces {
		err := exec.Command(
			"kubectl",
			"delete",
			"secret",
			DockerSecretName,
			"-n",
			namespace,
			"--kubeconfig",
			kubeconfigPath,
		).Run()
		if err != nil {
			log.Errorf("Failed to delete the Docker Hub secret in the %s namespace: %v\n", namespace, err)
		}
	}
}

func CreateImagePuller(kubeconfigPath string, images []string) error {
	args := []string{
		ImagePullerDeployScript,
		ImagePullerManifest,
		UpdatedImagePullerManifest,
		kubeconfigPath,
	}

	args = append(args, images...)

	var errb bytes.Buffer

	cmd := exec.Command(
		"bash",
		args...,
	)

	cmd.Stderr = &errb

	err := cmd.Run()
	if len(errb.String()) != 0 {
		return errors.New(errb.String())
	}

	return err
}

func DeleteImagePuller(kubeconfigPath string) error {
	return exec.Command(
		"kubectl",
		"delete",
		"daemonset",
		ImagePullerDaemonsetName,
		"--kubeconfig",
		kubeconfigPath,
	).Run()
}

// CreateAwsCredentialsSecret creates a k8s secret for the AWS credentials.
func CreateAwsCredentialsSecret(accessKeyId, secretAccessKey, kubeconfigPath string) {
	namespaces := []string{SystemNamespace, UserNamespace}
	for _, namespace := range namespaces {
		if err := exec.Command(
			"kubectl",
			"create",
			"secret",
			"generic",
			"-n",
			namespace,
			AwsCredentialsSecretName,
			fmt.Sprintf("--from-literal=AWS_ACCESS_KEY_ID=%s", accessKeyId),
			fmt.Sprintf("--from-literal=AWS_SECRET_ACCESS_KEY=%s", secretAccessKey),
			"--kubeconfig",
			kubeconfigPath,
		).Run(); err != nil {
			log.Errorf("Failed to create the AWS credentials secret in the %s namespace: %v", namespace, err)
		}
	}
}

// DeleteAwsCredentialsSecret deletes the k8s secret for the AWS credentials.
func DeleteAwsCredentialsSecret(kubeconfigPath string) {
	namespaces := []string{SystemNamespace, UserNamespace}
	for _, namespace := range namespaces {
		if err := exec.Command(
			"kubectl",
			"delete",
			"secret",
			AwsCredentialsSecretName,
			"-n",
			namespace,
			"--kubeconfig",
			kubeconfigPath,
		).Run(); err != nil {
			log.Errorf("Failed to delete the AWS credentials secret in the %s namespace: %v", namespace, err)
		}
	}
}
