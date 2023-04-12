#!/usr/bin/env bash

export AIRFLOW_HOME=~/airflow

AIRFLOW_VERSION=2.5.3
PYTHON_VERSION="$(python3 --version | cut -d " " -f 2 | cut -d "." -f 1-2)"

CONSTRAINT_URL="https://raw.githubusercontent.com/apache/airflow/constraints-${AIRFLOW_VERSION}/constraints-${PYTHON_VERSION}.txt"
pip3 install "apache-airflow==${AIRFLOW_VERSION}" --constraint "${CONSTRAINT_URL}"

AIRFLOW__WEBSERVER__WEB_SERVER_PORT=8000 airflow standalone