FROM python:3.9

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

ENV PYTHONUNBUFFERED 1

RUN pip install aqueduct-ml==0.0.16

CMD aqueduct start --expose

