FROM python:3.9

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

ENV PYTHONUNBUFFERED 1

RUN pip install aqueduct-ml
RUN pip install scikit-learn

COPY ./.devcontainer/deploy_demo_workflow.sh /

COPY ./manual_qa_tests /manual_qa_tests

RUN bash /deploy_demo_workflow.sh