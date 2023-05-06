#!/usr/bin/env bash

export AIRFLOW_HOME=~/airflow

AIRFLOW_VERSION=2.5.3
PYTHON_VERSION="$(python3 --version | cut -d " " -f 2 | cut -d "." -f 1-2)"

CONSTRAINT_URL="https://raw.githubusercontent.com/apache/airflow/constraints-${AIRFLOW_VERSION}/constraints-${PYTHON_VERSION}.txt"
pip3 install "apache-airflow==${AIRFLOW_VERSION}" --constraint "${CONSTRAINT_URL}"

export AIRFLOW__WEBSERVER__WEB_SERVER_PORT=8000
# Allows basic username and password based auth 
export AIRFLOW__API__AUTH_BACKENDS=airflow.api.auth.backend.basic_auth
# Has scheduler check for new DAGs every 15s
export AIRFLOW__SCHEDULER__DAG_DIR_LIST_INTERVAL=15

# Safety check to ensure that the AIRFLOW_HOME/dags directory actually exists
mkdir ${AIRFLOW_HOME}/dags

airflow standalone
