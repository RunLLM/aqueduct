package k8s

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
)

func CreateAwsFullS3Role(
	serviceAccountName string,
	serviceAccountNamespace string,
	roleName string,
	oidcIssuerUri *string,
	openIDConnectProviderArn string,
	awsRegion string,
	clusterName string,
) string {
	sess := session.Must(session.NewSession())
	iamClient := iam.New(sess, aws.NewConfig().WithRegion(awsRegion))

	_, err := iamClient.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(AppendClusterName(clusterName, roleName)),
	})
	if err == nil {
		// AWS IAM role already exists, so we will delete it
		log.Infof("Deleting existing role: %s", AppendClusterName(clusterName, roleName))
		_, err := iamClient.DetachRolePolicy(&iam.DetachRolePolicyInput{
			PolicyArn: aws.String(AwsS3AccessArn),
			RoleName:  aws.String(AppendClusterName(clusterName, roleName)),
		})
		if err != nil {
			log.Fatal("Unable to detach policy from role.")
		}
		_, err = iamClient.DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String(AppendClusterName(clusterName, roleName))})
		if err != nil {
			log.Fatal("Unable to delete AWS IAM role.")
		}
	}

	oidcIssuerWithoutProtocol := strings.Replace(*oidcIssuerUri, "https://", "", 1) // Remove the HTTPS protocol from the issuer URL.

	policyDocument := fmt.Sprintf(
		AwsRoleTrustRelationship,
		openIDConnectProviderArn,
		oidcIssuerWithoutProtocol,
		serviceAccountNamespace,
		serviceAccountName,
	)

	role, err := iamClient.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(policyDocument),
		Description:              aws.String(fmt.Sprintf("Allows service-account %s S3 access", serviceAccountName)),
		PermissionsBoundary:      aws.String(AwsS3AccessArn),
		RoleName:                 aws.String(AppendClusterName(clusterName, roleName)),
	})
	if err != nil {
		log.Fatal("Unexpected error while creating AWS role for full S3 access: ", err)
	}

	_, err = iamClient.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String(AwsS3AccessArn),
		RoleName:  aws.String(AppendClusterName(clusterName, roleName)),
	})
	if err != nil {
		log.Fatal("Unable to attach S3 policy to role: ", err)
	}

	return *role.Role.Arn
}

//	The goal of this helper function is to add the name of the cluster to the name of the resource to be created to
//	prevent creating resources with duplicate names when we spin up multiple clusters.
func AppendClusterName(clusterName, resourceName string) string {
	return fmt.Sprintf("%s-%s", clusterName, resourceName)
}
