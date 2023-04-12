import sys

import boto3

CLUSTER_NAME = "integration-test-shared"

STATUS_AVAILABLE = "available"
STATUS_PAUSED = "paused"


def resume_redshift(aws_access_key_id, aws_secret_access_key):
    client = _create_client(aws_access_key_id, aws_secret_access_key)

    status = _get_cluster_status(client, CLUSTER_NAME)
    if status == STATUS_AVAILABLE:
        print(f"The {CLUSTER_NAME} cluster is already available, it does not need to be resumed")
    elif status == STATUS_PAUSED:
        client.resume_cluster(ClusterIdentifier=CLUSTER_NAME)
    else:
        sys.exit(f"Cannot resume {CLUSTER_NAME} cluster because it is in the {status} state")


def pause_redshift(aws_access_key_id, aws_secret_access_key):
    client = _create_client(aws_access_key_id, aws_secret_access_key)

    status = _get_cluster_status(client, CLUSTER_NAME)
    if status == STATUS_PAUSED:
        print(f"The {CLUSTER_NAME} cluster is already paused, it does not need to be paused")
    elif status == STATUS_AVAILABLE:
        client.pause_cluster(ClusterIdentifier=CLUSTER_NAME)
    else:
        sys.exit(f"Cannot pause {CLUSTER_NAME} cluster because it is in the {status} state")


def _create_client(aws_access_key_id, aws_secret_access_key):
    return boto3.client(
        "redshift",
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
    )


def _get_cluster_status(client, cluster_identifier):
    response = client.describe_clusters(ClusterIdentifier=cluster_identifier)

    if "Clusters" not in response:
        sys.exit(f"Unable to {cluster_identifier} cluster response")

    clusters = response["Clusters"]
    if not clusters or len(clusters) == 0:
        sys.exit(f"Unable to find {cluster_identifier} cluster in response")

    cluster = clusters[0]
    if "ClusterStatus" not in cluster:
        sys.exit(f"Unable to find {cluster_identifier} cluster status in response")

    status = cluster["ClusterStatus"]
    if status == "available":
        return STATUS_AVAILABLE
    elif status == "paused":
        return STATUS_PAUSED
    return status
