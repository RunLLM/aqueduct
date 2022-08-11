FROM python:3.8.5


MAINTAINER Spiral Labs <hello@spiralai.co> version: 0.1

USER root

RUN apt-get update && \
  python3 -m pip install --upgrade pip && \
  pip3 install boto3==1.18.0 pandas==1.4.1 pydantic==1.9.0 \
  aqueduct-ml

# Copy over the Python codebase
COPY ./python/aqueduct_executor_enterprise /aqueduct_executor_enterprise

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor_enterprise.operators.param_executor.main --spec "$JOB_SPEC"