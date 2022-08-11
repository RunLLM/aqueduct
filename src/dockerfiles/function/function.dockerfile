FROM continuumio/miniconda3:4.10.3

MAINTAINER Aqueduct <hello@spiralai.co> version: 0.0.1

COPY ./golang/dockerfiles/operators/py38_environment.yml .
RUN conda env create -f py38_environment.yml

COPY ./golang/dockerfiles/operators/py39_environment.yml .
RUN conda env create -f py39_environment.yml

COPY ./golang/dockerfiles/operators/py37_environment.yml .
RUN conda env create -f py37_environment.yml

COPY ./golang/dockerfiles/operators/py310_environment.yml .
RUN conda env create -f py310_environment.yml

USER root

# Create a directory in which the application code will live.
RUN mkdir -p /app/function/

RUN apt-get update && \
  python3 -m pip install --upgrade pip && \
  pip3 install boto3==1.18.0 pandas==1.4.1 pydantic==1.9.0 \
  aqueduct-ml

# Copy over the Python codebase
COPY ./python/aqueduct_executor_enterprise /aqueduct_executor_enterprise

ENV PYTHONUNBUFFERED 1

COPY ./golang/dockerfiles/operators/start-function-executor.sh /

CMD bash /start-function-executor.sh