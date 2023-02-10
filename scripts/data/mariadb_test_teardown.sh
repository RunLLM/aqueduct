#!/usr/bin/env bash

# This script stops and removes the MariaDB Docker container. 

# Example usage: `bash scripts/data/mariadb_test_teardown.sh`

docker stop mdb
docker rm mdb
