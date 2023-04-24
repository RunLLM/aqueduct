FROM nvidia/cuda:11.8.0-runtime-ubuntu22.04

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

RUN apt-get -y update \
    && apt-get install -y wget \
    && apt-get install -y software-properties-common
RUN apt-get -y update

# Install miniconda
ENV CONDA_DIR /opt/conda
RUN wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh -O ~/miniconda.sh && \
     /bin/bash ~/miniconda.sh -b -p /opt/conda

# Put conda in path so we can use conda activate
ENV PATH=$CONDA_DIR/bin:$PATH

COPY ./gpu/py310_env.yml .
RUN conda env create -f py310_env.yml

ENV PYTHONUNBUFFERED 1

# Install Aqueduct LLM wrapper
RUN conda run -n py310_env pip install aqueduct-llm==0.2.11

WORKDIR /

COPY ./gpu/start-function-executor-gpu.sh /

CMD ["bash","/start-function-executor-gpu.sh", "py310_env"]

