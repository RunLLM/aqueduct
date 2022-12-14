FROM aqueducthq/base_connector:0.1.6

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.1

USER root

# Install dependencies

# Setup SQL Server Driver for Debian
# https://docs.microsoft.com/en-us/sql/connect/odbc/linux-mac/installing-the-microsoft-odbc-driver-for-sql-server?view=sql-server-ver15#debian17
RUN su
RUN curl https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
RUN curl https://packages.microsoft.com/config/debian/10/prod.list > /etc/apt/sources.list.d/mssql-release.list
RUN exit
RUN apt-get update
RUN ACCEPT_EULA=Y apt-get install -y msodbcsql17

# Install PyODBC
# https://github.com/mkleehammer/pyodbc/wiki/Install#debian-stretch
RUN apt-get install -y g++ unixodbc-dev
RUN pip3 install pyodbc

ENV PYTHONUNBUFFERED 1

CMD python3 -m aqueduct_executor.operators.connectors.data.main --spec "$JOB_SPEC"