package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type serviceType string

const (
	InternalService        serviceType = "internal"
	ExternalService        serviceType = "external"
	ExternalIngressService serviceType = "external-ingress"
)

//	This is a helper function that creates a new service in the specified
//	namespace. The service will apply to the pods with the labels specified in
//	the `selectors` argument and will expose the port(s) requested by the
//	`ports` argument. Once the service object is created, this function will
//	synchronously wait for the service to finish creating before returning.
// `systemService` indicates whether this is a system service.
// `serviceType` should be one of the following:
//    - `InternalService`: A k8s service that can only be accessed inside the cluster.
//    - `ExternalService`: A k8s service that can be accessed from outside the cluster.
//    - `ExternalIngressService`: A k8s service that can be accessed from outside the cluster using an ingress. This
//       should be used when a cloud load balancer needs to be created for the service.
func CreateService(
	name string,
	systemService bool,
	svcType serviceType,
	selectors *map[string]string,
	ports map[uint32]uint32,
	annotations *map[string]string,
	k8sClient *kubernetes.Clientset,
) error {
	namespace := SystemNamespace
	if !systemService {
		namespace = UserNamespace
	}

	var k8sServiceType corev1.ServiceType
	switch svcType {
	case InternalService:
		k8sServiceType = corev1.ServiceTypeClusterIP
		break
	case ExternalService:
		k8sServiceType = corev1.ServiceTypeLoadBalancer
		break
	case ExternalIngressService:
		k8sServiceType = corev1.ServiceTypeNodePort
		break
	default:
		return errors.Newf("Unknown service type specified: %v", svcType)
	}

	// Convert from a list of integer ports that will be exposed by the service
	// to Kubernetes `ServicePort`s. For now, we are hardcoding the port names as
	// well as the protocol for now.
	servicePorts := make([]corev1.ServicePort, len(ports))

	idx := 0
	for external, internal := range ports {
		servicePort := corev1.ServicePort{
			Name:       fmt.Sprintf("%s-port-%d", name, idx), // We are hardcoding these names. Will we ever care about these?
			Protocol:   corev1.ProtocolTCP,                   // Again, will we ever want to change this?
			Port:       int32(external),
			TargetPort: intstr.FromInt(int(internal)),
		}

		servicePorts[idx] = servicePort
		idx += 1
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: *annotations,
		},

		Spec: corev1.ServiceSpec{
			Ports:    servicePorts,
			Selector: *selectors,
			Type:     k8sServiceType,
		},
	}

	_, err := k8sClient.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Wait until the service is given an address.
	timeWaited := 0
	serviceAddress := ""
	for serviceAddress == "" {
		if timeWaited > ServiceCreationTimeoutSec {
			return fmt.Errorf("timeout creating service with the latest error: %s", err)
		}
		service, err = GetService(name, k8sClient, systemService)
		if err != nil {
			time.Sleep(5 * time.Second)
			timeWaited += 5
			continue
		}
		serviceAddress, err = GetServiceAddress(service)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
		timeWaited += 5
	}
	log.Infof("Service finished creating! %s ready to be accessed at: %s\n", name, serviceAddress)

	return nil
}

func GetServiceAddress(s *corev1.Service) (string, error) {
	switch s.Spec.Type {
	// We return the cluster IP even for NodePort and LoadBalancer services, because the service address
	// is not used directly for those service types.
	case corev1.ServiceTypeClusterIP, corev1.ServiceTypeNodePort, corev1.ServiceTypeLoadBalancer:
		return s.Spec.ClusterIP, nil
	default:
		return "", errors.New("unknown service type")
	}
}

//	This helper function takes in the name of a service and returns the
//	Kubernetes client object representation of that service.
func GetService(
	name string,
	k8sClient *kubernetes.Clientset,
	systemService bool,
) (*corev1.Service, error) {
	var namespace string
	if systemService {
		namespace = SystemNamespace
	} else {
		namespace = UserNamespace
	}

	return k8sClient.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func GetServiceAddressFromName(
	name string,
	k8sClient *kubernetes.Clientset,
	systemService bool,
) (string, error) {
	svc, err := GetService(name, k8sClient, systemService)
	if err != nil {
		return "", err
	}

	addr, err := GetServiceAddress(svc)
	if err != nil {
		return "", err
	}

	return addr, nil
}

// GetExternalServiceAddressFromName returns the external IP of a service (if there is one).
// If the service type is not LoadBalancer, it returns an error.
func GetExternalServiceAddressFromName(
	name string,
	k8sClient *kubernetes.Clientset,
	systemService bool,
) (string, error) {
	svc, err := GetService(name, k8sClient, systemService)
	if err != nil {
		return "", err
	}

	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return "", errors.Newf("The external IP is not defined for a k8s service of type %v", svc.Spec.Type)
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", errors.Newf("No load balancer ingress resources found for this service.")
	}

	return svc.Status.LoadBalancer.Ingress[0].Hostname, nil
}

// GetServiceNodePort returns the NodePort for a service. This is the port on which this service is
// exposed when the service type is NodePort or LoadBalancer. It assumes that there is only one port exposed.
func GetServiceNodePort(
	name string,
	k8sClient *kubernetes.Clientset,
	systemService bool,
) (int32, error) {
	service, err := GetService(name, k8sClient, systemService)
	if err != nil {
		return -1, err
	}

	return service.Spec.Ports[0].NodePort, nil
}

func DeleteService(
	name string,
	k8sClient *kubernetes.Clientset,
	systemService bool,
) error {
	var namespace string
	if systemService {
		namespace = SystemNamespace
	} else {
		namespace = UserNamespace
	}

	return k8sClient.CoreV1().Services(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

//	This helper function deletes all of the services in a given namespace by
//	using the Kubernetes API to list all services existing in that namespace
//	and then deleting them one-by-one.
func DeleteServicesByNamespace(
	k8sClient *kubernetes.Clientset,
	namespace string,
) {
	servicesClient := k8sClient.CoreV1().Services(namespace)
	serviceList, err := servicesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error("Unexpected error while listing services:", err)
		return
	}

	for _, svc := range serviceList.Items {
		err = servicesClient.Delete(context.TODO(), svc.ObjectMeta.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Error when deleting service %s: %v.\n", svc.ObjectMeta.Name, err)
			continue
		}
	}
	log.Infof("Successfully deleted services in namespace %s", namespace)
}
