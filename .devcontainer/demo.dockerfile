FROM python:3.9

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

ENV PYTHONUNBUFFERED 1

RUN pip install aqueduct-ml scikit-learn transformers

COPY ./manual_qa_tests /deploy_notebooks
COPY ./.devcontainer/deploy_demo_workflow.sh /
COPY ./examples /examples

RUN bash /deploy_demo_workflow.sh
RUN rm /deploy_demo_workflow.sh
RUN rm -rf /deploy_notebooks
RUN rm -rf /examples