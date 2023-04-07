FROM python:3.10

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

# Create a directory in which the application code will live.
RUN mkdir -p /app/function/

RUN pip install numpy==1.22.2 \
pandas==1.4.1 \
pip==22.0.4 \
scipy==1.7.3 \
cloudpickle==2.0.0 \
pyarrow==7.0.0 \
boto3==1.18.0 \
pydantic==1.9.0 \
scikit_learn==1.0.2 \
aqueduct-ml==0.2.9

ENV PYTHONUNBUFFERED 1

COPY ./function/start-function-executor.sh /

CMD bash /start-function-executor.sh