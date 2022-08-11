FROM aqueducthq/base_connector:latest

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.1

USER root

# Install dependencies
RUN pip3 install snowflake-sqlalchemy

ENV PYTHONUNBUFFERED 1

COPY ./connectors/test/start-snowflake.sh /

CMD bash /start-snowflake.sh