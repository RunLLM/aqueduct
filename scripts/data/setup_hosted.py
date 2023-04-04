import argparse

from redshift_test import resume_redshift


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument("--aws-key-id", required=True, help="AWS Access Key ID")
    parser.add_argument("--aws-secret-key", required=True, help="AWS Secret Access Key")
    args = parser.parse_args()

    aws_access_key_id = args.aws_key_id
    aws_secret_access_key = args.aws_secret_key

    setup_redshift(aws_access_key_id, aws_secret_access_key)


def setup_redshift(aws_access_key_id, aws_secret_access_key):
    resume_redshift(aws_access_key_id, aws_secret_access_key)


if __name__ == "__main__":
    main()
