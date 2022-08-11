FROM aqueducthq/connector_base:latest

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.1

USER root

# Install dependencies
RUN pip3 install pyarrow

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor_enterprise.operators.connectors.tabular.main --spec "$JOB_SPEC"