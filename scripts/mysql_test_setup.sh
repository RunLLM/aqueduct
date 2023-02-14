#!/bin/bash

# Create the Docker container
docker run --name aqueduct-mysql -e MYSQL_ROOT_PASSWORD='secret' -d -p 3306:3306 mysql

# Wait for the database to be ready
echo "Waiting for MySQL database container to be ready ..."
until docker exec aqueduct-mysql mysql -uroot -psecret -e 'show databases;' &> /dev/null; do
  sleep 1
done

# Create the 'aqueducttest' database
echo "Creating aqueducttest database ..."
docker exec -it aqueduct-mysql mysql -uroot -psecret -e "CREATE DATABASE aqueducttest;"

echo "MySQL container with the 'aqueducttest' database created successfully."
