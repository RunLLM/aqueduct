#!/usr/bin/env bash

# First, start up the container
docker run --name aqueduct-postgres -e POSTGRES_PASSWORD=aqueduct -d postgres -p 5432:5432
# TODO: Figure out how to do commands below in bash script.
# Open terminal inside the container
docker exec -it <container_id> bash
# Run Psql to create a database
psql -U postgres
CREATE DATABASE aqueducttest;
