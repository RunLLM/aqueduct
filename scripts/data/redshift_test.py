import boto3

CLUSTER_NAME = 'integration-test-shared'

def resume_redshift(aws_access_key_id, aws_secret_access_key):
    client = _create_client(aws_access_key_id, aws_secret_access_key)

    response = client.resume_cluster(
        ClusterIdentifier=CLUSTER_NAME
    )


def pause_redshift(aws_access_key_id, aws_secret_access_key):
    client = _create_client(aws_access_key_id, aws_secret_access_key)

    client.pause_cluster(
        ClusterIdentifier=CLUSTER_NAME
    )


def _create_client(aws_access_key_id, aws_secret_access_key):
    return boto3.client(
        'redshift',
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key
    )
