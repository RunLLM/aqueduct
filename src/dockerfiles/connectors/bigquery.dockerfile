FROM aqueducthq/base_connector:latest

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.1

USER root

# Install dependencies
RUN pip3 install google-cloud-bigquery

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor.operators.connectors.tabular.main --spec "$JOB_SPEC"