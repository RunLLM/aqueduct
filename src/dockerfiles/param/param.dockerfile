FROM python:3.8.5

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

RUN apt-get update && \
  python3 -m pip install --upgrade pip && \
  pip3 install \
  aqueduct-ml \
  boto3 \ 
  pandas \
  pydantic

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor.operators.param_executor.main --spec "$JOB_SPEC"