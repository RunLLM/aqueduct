#!/usr/bin/env bash

# This script creates a MariaDB database in a Docker container.
# Example usage: `bash scripts/data/mariadb_test_setup.sh`

docker run --name mdb -e MARIADB_ROOT_PASSWORD=Password123! -e MARIADB_DATABASE=mariadb_test -p 3306:3306 -d mariadb:latest
