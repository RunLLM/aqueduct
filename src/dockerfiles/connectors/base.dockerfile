FROM python:3.8.5

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.1

ENV PYTHONUNBUFFERED 1

# Update apt-get and install basic utilities via apt-get and pip3.
RUN apt-get update && \
  python3 -m pip install --upgrade pip && \
  pip3 install \
  aqueduct-ml \
  boto3==1.18.0 \
  cloudpickle \
  distro \
  pandas==1.3.0 \
  pydantic==1.9.0 \
  pyyaml \
  SQLAlchemy==1.4.30 \
  typing_extensions
