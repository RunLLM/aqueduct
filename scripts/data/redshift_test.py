import sys
import time

import boto3

CLUSTER_NAME = "integration-test-shared"

STATUS_AVAILABLE = "available"
STATUS_PAUSED = "paused"
STATUS_PAUSING = "pausing"
STATUS_RESUMING = "resuming"


def resume_redshift(aws_access_key_id, aws_secret_access_key, retry=0):
    """
    Resumes the test Redshift cluster.
    """
    client = _create_client(aws_access_key_id, aws_secret_access_key)
    status = _get_cluster_status(client, CLUSTER_NAME)

    if status == STATUS_AVAILABLE:
        # Nothing to do, the cluster is already ready
        print("The cluster is already available. There is no action to take.")
        pass
    elif status == STATUS_PAUSED:
        # Cluster can be resumed
        print("The cluster is currently paused. We will try to resume it.")
        try:
            client.resume_cluster(ClusterIdentifier=CLUSTER_NAME)
        except client.exceptions.InvalidClusterStateFault as e:
            # This exception handling is required because of a transient issue where
            # the cluster has another operation in progress, but the cluster status
            # does not reflect that.
            if retry >= 5:
                sys.exit(f"Unable to resume cluster due to {e} exception even after 5 retries")

            # Sleep and retry
            print(
                "There was an InvalidClusterStateFault when resuming the cluster. We will sleep and retry."
            )
            time.sleep(15)
            resume_redshift(aws_access_key_id, aws_secret_access_key, retry=retry + 1)

        _wait_for_status(client, STATUS_AVAILABLE)
    elif status == STATUS_PAUSING:
        # First need to wait for cluster to completely pause before resuming it
        print(
            "The cluster is currently pausing so we need to wait for it to fully pause, before resuming it."
        )
        _wait_for_status(client, STATUS_PAUSED)
        resume_redshift(aws_access_key_id, aws_secret_access_key)
    elif status == STATUS_RESUMING:
        # Wait for resuming operation to complete
        print("The cluster is currently resuming so we will just wait for it to become available.")
        _wait_for_status(client, STATUS_AVAILABLE)
    else:
        sys.exit(f"Cannot resume {CLUSTER_NAME} cluster because it is in the {status} state")

    print(f"{CLUSTER_NAME} cluster is ready!")


def pause_redshift(aws_access_key_id, aws_secret_access_key, retry=0):
    """
    Pauses the test Redshift cluster.
    """
    client = _create_client(aws_access_key_id, aws_secret_access_key)
    status = _get_cluster_status(client, CLUSTER_NAME)

    if status == STATUS_AVAILABLE:
        # Cluster can be paused
        print("The cluster is available so we will try to pause it.")
        try:
            client.pause_cluster(ClusterIdentifier=CLUSTER_NAME)
        except client.exceptions.InvalidClusterStateFault as e:
            # This exception handling is required because of a transient issue where
            # the cluster has another operation in progress, but the cluster status
            # does not reflect that.
            if retry >= 5:
                sys.exit(f"Unable to pause cluster due to {e} exception even after 5 retries")

            # Sleep and retry
            print(
                "There was an InvalidClusterStateFault when pausing the cluster. We will sleep and retry."
            )
            time.sleep(15)
            pause_redshift(aws_access_key_id, aws_secret_access_key, retry=retry + 1)

        _wait_for_status(client, STATUS_PAUSED)
    elif status == STATUS_PAUSED:
        # Nothing to do, the cluster is already paused
        print("The cluster is already paused. There is no action to take.")
        pass
    elif status == STATUS_PAUSING:
        # Wait for pausing operation to complete
        print("The cluster is currently pausing so we will just wait for it to become paused.")
        _wait_for_status(client, STATUS_PAUSED)
    elif status == STATUS_RESUMING:
        # Wait for cluster to finish resuming, before it can be paused
        print(
            "The cluster is currently resuming so we need to wait for it to become available before pausing it."
        )
        _wait_for_status(client, STATUS_AVAILABLE)
        pause_redshift(aws_access_key_id, aws_secret_access_key)
    else:
        sys.exit(f"Cannot pause {CLUSTER_NAME} cluster because it is in the {status} state")

    print(f"{CLUSTER_NAME} cluster has been paused")


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

    return cluster["ClusterStatus"]


def _wait_for_status(client, desired_status, timeout=900):
    """
    Waits for the test cluster to reach the desired status. Errors if the timeout
    is reached.
    """
    print(f"Waiting for {CLUSTER_NAME} cluster to enter {desired_status} status...")

    status = _get_cluster_status(client, CLUSTER_NAME)
    start = time.time()
    while status != desired_status:
        if time.time() > start + timeout:
            sys.exit(f"Reached timeout waiting for {CLUSTER_NAME} cluster to reach {status} status")
        time.sleep(15)
        status = _get_cluster_status(client, CLUSTER_NAME)

    print(f"{CLUSTER_NAME} cluster has reached {status} status")
