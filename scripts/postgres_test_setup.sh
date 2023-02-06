#!/bin/bash

# Create the Docker container
docker run --name aqueduct-postgres -e POSTGRES_PASSWORD=aqueduct -d -p 5432:5432 postgres

# Wait for the database to be ready
echo "Waiting for database container to be ready ..."
until docker exec aqueduct-postgres psql -U postgres -c '\l' &> /dev/null; do
  sleep 1
done

# Create the 'aqueducttest' database
echo "Creating aqueducttest database ..."
docker exec -it aqueduct-postgres psql -U postgres -c "CREATE DATABASE aqueducttest;"

echo "PostgreSQL container with the 'aqueducttest' database created successfully."
