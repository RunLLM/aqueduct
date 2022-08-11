FROM spiralco/connector_base:latest

MAINTAINER Aqueduct <hello@spiralai.co> version: 0.1

USER root

# Install dependencies
RUN apt-get install -y python3-dev default-libmysqlclient-dev build-essential
RUN pip3 install mysqlclient==2.1.0

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor_enterprise.operators.connectors.tabular.main --spec "$JOB_SPEC"