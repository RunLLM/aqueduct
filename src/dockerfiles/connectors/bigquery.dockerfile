FROM aqueducthq/base_connector:0.3.4

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.1

USER root

# Install dependencies
RUN pip3 install google-cloud-bigquery

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor.operators.connectors.data.main --spec "$JOB_SPEC" --version-tag "$VERSION_TAG"