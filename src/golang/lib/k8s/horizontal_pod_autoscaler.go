package k8s

import (
	"context"

	log "github.com/sirupsen/logrus"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateHorizontalPodAutoscaler creates an autoscaler to pod deployments specified
// by deploymentName and namespace.
func CreateHorizontalPodAutoscaler(
	deploymentName string, namespace string, k8sClient *kubernetes.Clientset,
) error {
	createOptions := metav1.CreateOptions{}
	minReplicas := int32(MinReplicas)
	targetCPUUtilizationPercentage := int32(TargetCPUUtilizationPercentage)
	stablizationWindowSec := int32(StablizationWindowSec)
	horizontalPodAutoscaler := autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			// The target on which this autoscaler will monitor
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind: "deployment",
				Name: deploymentName,
				// This points to the deployment API version
				APIVersion: "apps/v1",
			},
			MinReplicas: &minReplicas,
			MaxReplicas: MaxReplicas,
			// The following config sets the scaling target to be avg CPU utilization across all pods
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: "cpu",
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: &targetCPUUtilizationPercentage,
						},
					},
				},
			},
			Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
				// The following configs limit the pod upscale / downscale
				// change rate to 3 pods per 120 seconds, based on the metrics collected
				// in a window specified by StablizationWindowSeconds.
				ScaleUp: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: &stablizationWindowSec,
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         HpaPodChangeLimit,
							PeriodSeconds: HpaPodChangeLimitPeriod,
						},
					},
				},
				ScaleDown: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: &stablizationWindowSec,
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         HpaPodChangeLimit,
							PeriodSeconds: HpaPodChangeLimitPeriod,
						},
					},
				},
			},
		},
	}
	_, err := k8sClient.
		AutoscalingV2beta2().
		HorizontalPodAutoscalers(namespace).
		Create(context.Background(), &horizontalPodAutoscaler, createOptions)
	return err
}

func DeleteHorizontalPodAutoscalerByNamespace(namespace string, k8sClient *kubernetes.Clientset) {
	hpaClient := k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace)
	hpaList, err := hpaClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Error listing HPAs: %s", err)
		return
	}
	for _, hpa := range hpaList.Items {
		name := hpa.ObjectMeta.Name
		hpaClient.Delete(context.Background(), name, metav1.DeleteOptions{})
	}
}

func DeleteHorizontalPodAutoscaler(namespace, deploymentName string, k8sClient *kubernetes.Clientset) error {
	hpaClient := k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace)
	_, err := hpaClient.Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err == nil {
		// We delete HPA only if it exists.
		return hpaClient.Delete(context.Background(), deploymentName, metav1.DeleteOptions{})
	}

	return nil
}
