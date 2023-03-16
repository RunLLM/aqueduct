FROM python:3.9

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

ENV PYTHONUNBUFFERED 1

RUN pip install aqueduct-ml
RUN aqueduct install postgres
RUN pip install scikit-learn

COPY ./.devcontainer/deploy_demo_workflow.sh /

COPY ./.devcontainer/demo_setup.py /

COPY ./.devcontainer/deploy_example.py /

COPY ./.devcontainer/demo.ipynb /notebook/demo.ipynb

RUN bash /deploy_demo_workflow.sh

RUN rm /demo_setup.py
RUN rm /deploy_example.py
RUN rm /deploy_demo_workflow.sh