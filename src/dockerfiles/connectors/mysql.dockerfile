FROM aqueducthq/base_connector:0.2.9

MAINTAINER Aqueduct <hello@spiralai.co> version: 0.1

USER root

# Install dependencies
RUN apt-get install -y python3-dev default-libmysqlclient-dev build-essential
RUN pip3 install mysqlclient==2.1.0

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor.operators.connectors.data.main --spec "$JOB_SPEC"