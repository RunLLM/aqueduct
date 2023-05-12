#!/bin/bash
aws --profile default configure set aws_access_key_id $AWS_ACCESS_KEY_ID
aws --profile default configure set aws_secret_access_key $AWS_SECRET_ACCESS_KEY
aws --profile default configure set region $AWS_REGION

if [ $VERSION_TAG ]
then
    pip install --index-url "https://test.pypi.org/simple/" --extra-index-url "https://pypi.org/simple" aqueduct-ml==$VERSION_TAG
fi
pip install $DEPENDENCIES
conda-pack -o $ENV_FILE_NAME --ignore-editable-packages
aws s3 cp ./$ENV_FILE_NAME $S3_BUCKET