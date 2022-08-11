FROM continuumio/miniconda3:4.10.3

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

COPY ./function/py37_environment.yml .
RUN conda env create -f py37_environment.yml

COPY ./function/py38_environment.yml .
RUN conda env create -f py38_environment.yml

COPY ./function/py39_environment.yml .
RUN conda env create -f py39_environment.yml

COPY ./function/py310_environment.yml .
RUN conda env create -f py310_environment.yml

USER root

# Create a directory in which the application code will live.
RUN mkdir -p /app/function/

RUN apt-get update && \
  python3 -m pip install --upgrade pip && \
  pip3 install \
  aqueduct-ml \
  boto3 \
  pandas \
  pydantic

ENV PYTHONUNBUFFERED 1

COPY ./function/test/start-function-executor.sh /

CMD bash /start-function-executor.sh