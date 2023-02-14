#!/bin/bash

# Create the Docker container
docker run --name aqueduct-postgres -e POSTGRES_PASSWORD=aqueduct -e POSTGRES_DB=aqueducttest -d -p 5432:5432 postgres
