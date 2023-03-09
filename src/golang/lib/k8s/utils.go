package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func generateK8sEnvVarAndResourceReq(environmentVariables *map[string]string, resourceRequests *map[string]string) ([]corev1.EnvVar, *corev1.ResourceRequirements) {
	// Convert from a `map[string]string` to the Kubernetes representation of
	// environment variables, which has its own special struct.
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
	if gpuName, ok := (*resourceRequests)[GPUResourceName]; ok {
		switch gpuName {
		case "nvidia.com/gpu":
			resourceList["nvidia.com/gpu"] = resource.MustParse(DefaultGPULimit)
		case "amd.com/gpu":
			resourceList["amd.com/gpu"] = resource.MustParse(DefaultGPULimit)
		default:
			resourceList["nvidia.com/gpu"] = resource.MustParse(DefaultGPULimit)
		}
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
