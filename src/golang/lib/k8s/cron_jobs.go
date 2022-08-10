package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// DefaultCronJobSchedule is only used when the cron job is suspended
	DefaultCronJobSchedule  = "0 0 1 * *"
	CronJobDeletionInterval = 5
	CronJobDeletionTimeout  = 60
)

//	A helper function that takes in the name of a cron job, a container image, and
//	other configuration parameters. It uses this information to generate a new
//	cron job and deploy that cron job to the underlying Kubernetes cluster.
func CreateCronJob(
	name, containerImage string,
	systemPod bool,
	environmentVariables *map[string]string,
	secretEnvVariables []string,
	resourceRequests *map[string]string,
	labels *map[string]string,
	serviceAccount string,
	privileged bool,
	k8sClient *kubernetes.Clientset,
	jobPeriod string,
) error {
	namespace := getDeploymentNamespace(systemPod)

	k8sEnvironmentVariables, resourceRequirements := generateK8sEnvVarAndResourceReq(environmentVariables, resourceRequests)
	// Add the name of the service as a label to the pod spec.
	(*labels)[ServiceKey] = name

	// This is an empty set of create options because we don't need any of these
	// configurations for now.
	createOptions := metav1.CreateOptions{}

	// This means the job, once completed, will be garbage collected after `ttlSeconds`.
	// Currently set to 3 days.
	ttlSeconds := int32(259200)

	cronJob := batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: jobPeriod,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &ttlSeconds,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
							Labels:    *labels,
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
			},
		},
	}

	podRuntimeClassName := "gvisor"
	cronJob.Spec.JobTemplate.Spec.Template.Spec.RuntimeClassName = &podRuntimeClassName
	if jobPeriod == "" {
		cronJob.Spec.Schedule = DefaultCronJobSchedule
		suspend := true
		cronJob.Spec.Suspend = &suspend
	}

	if serviceAccount != "" {
		cronJob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = serviceAccount
	}

	if len(secretEnvVariables) > 0 {
		// Assign environment variables from secret references
		cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].EnvFrom = generateK8sEnvVarFromSecrets(secretEnvVariables)
	}

	_, err := k8sClient.BatchV1beta1().CronJobs(cronJob.ObjectMeta.Namespace).Create(context.TODO(), &cronJob, createOptions)
	return err
}

// InvokeJob takes in the name of a cron job and invokes a single
// instance of it as a Kubernetes job. The name is used to find the previously
// created cron job.
func InvokeJob(name string, k8sClient *kubernetes.Clientset) error {
	cronJob, err := k8sClient.BatchV1beta1().CronJobs(UserNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, time.Now().Unix()),
			Namespace: UserNamespace,
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}

	_, err = k8sClient.BatchV1().Jobs(UserNamespace).Create(context.TODO(), &job, metav1.CreateOptions{})
	return err
}

func DeleteCronJob(name string, systemPod bool, k8sClient *kubernetes.Clientset) error {
	var namespace string
	if systemPod {
		namespace = SystemNamespace
	} else {
		namespace = UserNamespace
	}

	err := k8sClient.BatchV1beta1().CronJobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})

	deletionConditionFunction := func() (bool, error) {
		_, err := k8sClient.BatchV1beta1().CronJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return true, nil // meaning that the cronjob is already deleted, which is what we want.
		}
		log.Infof("Waiting for cronjob %s to be deleted...\n", name)
		return false, nil
	}

	// Wait for the cronjob that we just deleted to actually go away.
	err = wait.PollImmediate(CronJobDeletionInterval*time.Second, CronJobDeletionTimeout*time.Second, deletionConditionFunction)
	if err != nil {
		log.Errorf("Error while waiting for cronjob to be deleted: %v\n", err)
	}

	return err
}

func DeleteCronJobsByNamespace(k8sClient *kubernetes.Clientset, namespace string) {
	cronjobsClient := k8sClient.BatchV1beta1().CronJobs(namespace)
	cronjobs, err := cronjobsClient.List(context.TODO(), metav1.ListOptions{}) // List all of the cron jobs.
	if err != nil {
		log.Errorf("Unexpected error while retrieving cronjobs in namespace %s: %v\n", namespace, err)
		return
	}

	for _, cronjob := range cronjobs.Items {
		name := cronjob.ObjectMeta.Name

		err := cronjobsClient.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Unable to delete %s cronjob: %s\n", namespace, name)
			continue
		}
	}
	log.Infof("Successfully deleted cron jobs in namespace %s\n", namespace)
}

func DeleteJobsByNamespace(k8sClient *kubernetes.Clientset, namespace string) {
	jobsClient := k8sClient.BatchV1().Jobs(namespace)
	jobs, err := jobsClient.List(context.TODO(), metav1.ListOptions{}) // List all of the jobs.
	if err != nil {
		log.Errorf("Unexpected error while retrieving jobs in namespace %s: %v\n", namespace, err)
		return
	}

	for _, job := range jobs.Items {
		name := job.ObjectMeta.Name

		// Including this in the propagation policy will make sure the associated pods spun up by the jobs are also deleted.
		deleteDependencyInBackground := metav1.DeletePropagationBackground
		err := jobsClient.Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &deleteDependencyInBackground})
		if err != nil {
			log.Errorf("Unable to delete %s job: %s\n", namespace, name)
			continue
		}
	}
	log.Infof("Successfully deleted jobs in namespace %s\n", namespace)
}

func CronJobExists(
	name string, k8sClient *kubernetes.Clientset,
) bool {
	_, err := k8sClient.BatchV1beta1().CronJobs(UserNamespace).Get(context.Background(), name, metav1.GetOptions{})
	return err == nil
}

func EditCronJob(
	name string, jobPeriod string, k8sClient *kubernetes.Clientset,
) error {
	cronJob, err := k8sClient.BatchV1beta1().CronJobs(UserNamespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to fetch cron job")
	}

	// Suspend the cron job if the job period is empty
	if jobPeriod == "" {
		cronJob.Spec.Schedule = DefaultCronJobSchedule
		suspend := true
		cronJob.Spec.Suspend = &suspend
	} else if *cronJob.Spec.Suspend && jobPeriod != "" {
		// If the cron job is currently suspended, and a non-empty job period is provided,
		// unsuspend the cron job
		suspend := false
		cronJob.Spec.Suspend = &suspend
		cronJob.Spec.Schedule = jobPeriod
	} else {
		cronJob.Spec.Schedule = jobPeriod
	}

	_, err = k8sClient.BatchV1beta1().CronJobs(UserNamespace).Update(context.Background(), cronJob, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to update cron job")
	}
	return nil
}
