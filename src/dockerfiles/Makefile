.DEFAULT_GOAL = all

# By default, turn off the Makefile practice of printing each command before
# you run it.
ifndef VERBOSE
.SILENT:
endif

build-function:
	docker build . -t aqueducthq/function -f function/function.dockerfile --no-cache

build-param:
	docker build . -t aqueducthq/param -f param/param.dockerfile --no-cache

build-base-connector:
	docker build . -t aqueducthq/connector_base -f connectors/base.dockerfile --no-cache

build-connectors: build-bigquery-connector build-mysql-connector build-postgres-connector \
	build-s3-connector build-snowflake-connector build-sqlserver-connector

build-bigquery-connector:
	docker build . -t aqueducthq/bigquery-connector -f connectors/bigquery.dockerfile --no-cache

build-mysql-connector:
	docker build . -t aqueducthq/mysql-connector -f connectors/mysql.dockerfile --no-cache

build-postgres-connector:
	docker build . -t aqueducthq/postgres-connector -f connectors/postgres.dockerfile --no-cache

build-s3-connector:
	docker build . -t aqueducthq/s3-connector -f connectors/s3.dockerfile --no-cache

build-snowflake-connector:
	docker build . -t aqueducthq/snowflake-connector -f connectors/snowflake.dockerfile --no-cache

build-sqlserver-connector:
	docker build . -t aqueducthq/sqlserver-connector -f connectors/sqlserver.dockerfile --no-cache

publish-connectors:
	docker push aqueducthq/bigquery-connector
	docker push aqueducthq/mysql-connector
	docker push aqueducthq/postgres-connector
	docker push aqueducthq/s3-connector
	docker push aqueducthq/snowflake-connector
	docker push aqueducthq/sqlserver-connector

.PHONY:

