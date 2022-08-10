package k8s

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateIngress(
	ctx context.Context,
	name string,
	systemIngress bool,
	annotations *map[string]string,
	ingressPaths []v1.HTTPIngressPath,
	k8sClient *kubernetes.Clientset,
) error {
	namespace := SystemNamespace
	if !systemIngress {
		namespace = UserNamespace
	}

	ingress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: *annotations,
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{Paths: ingressPaths},
					},
				},
			},
		},
	}

	_, err := k8sClient.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

func GetIngressHostname(
	ctx context.Context,
	name string,
	systemIngress bool,
	k8sClient *kubernetes.Clientset,
) (string, error) {
	namespace := SystemNamespace
	if !systemIngress {
		namespace = UserNamespace
	}

	ingress, err := k8sClient.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	loadBalancers := ingress.Status.LoadBalancer.Ingress
	if len(loadBalancers) == 0 {
		return "", errors.New("No load balancers found for ingress.")
	}

	return loadBalancers[0].Hostname, nil
}

func DeleteIngressesByNamespace(
	k8sClient *kubernetes.Clientset,
	namespace string,
) {
	ingressesClient := k8sClient.NetworkingV1().Ingresses(namespace)
	ingressList, err := ingressesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unexpected error while listing ingresses: %v", err)
		return
	}

	for _, ingress := range ingressList.Items {
		err = ingressesClient.Delete(context.TODO(), ingress.ObjectMeta.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Unexpected error while deleting ingress %s: %v", ingress.ObjectMeta.Name, err)
			continue
		}
	}

	log.Infof("Successfully deleted ingresses in namespace: %s", namespace)
}
