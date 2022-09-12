.DEFAULT_GOAL = all

# By default, turn off the Makefile practice of printing each command before
# you run it.
ifndef VERBOSE
.SILENT:
endif

build-function:
	docker build . -t aqueducthq/function:0.0.14 -f function/function.dockerfile --no-cache

build-param:
	docker build . -t aqueducthq/param:0.0.14 -f param/param.dockerfile --no-cache

build-system-metric:
	docker build . -t aqueducthq/system-metric:0.0.14 -f system-metric/system-metric.dockerfile --no-cache

build-base-connector:
	docker build . -t aqueducthq/base_connector:0.0.14 -f connectors/base.dockerfile --no-cache

build-connectors: build-athena-connector build-bigquery-connector build-mysql-connector \
	build-postgres-connector build-s3-connector build-snowflake-connector build-sqlserver-connector

build-athena-connector:
	docker build . -t aqueducthq/athena-connector:0.0.14 -f connectors/athena.dockerfile --no-cache

build-bigquery-connector:
	docker build . -t aqueducthq/bigquery-connector:0.0.14 -f connectors/bigquery.dockerfile --no-cache

build-mysql-connector:
	docker build . -t aqueducthq/mysql-connector:0.0.14 -f connectors/mysql.dockerfile --no-cache

build-postgres-connector:
	docker build . -t aqueducthq/postgres-connector:0.0.14 -f connectors/postgres.dockerfile --no-cache

build-s3-connector:
	docker build . -t aqueducthq/s3-connector:0.0.14 -f connectors/s3.dockerfile --no-cache

build-snowflake-connector:
	docker build . -t aqueducthq/snowflake-connector:0.0.14 -f connectors/snowflake.dockerfile --no-cache

build-sqlserver-connector:
	docker build . -t aqueducthq/sqlserver-connector:0.0.14 -f connectors/sqlserver.dockerfile --no-cache

publish-function:
	docker push aqueducthq/function:0.0.14

publish-param:
	docker push aqueducthq/param:0.0.14

publish-system-metric:
	docker push aqueducthq/system-metric:0.0.14

publish-connectors:
	docker push aqueducthq/athena-connector:0.0.14
	docker push aqueducthq/bigquery-connector:0.0.14
	docker push aqueducthq/mysql-connector:0.0.14
	docker push aqueducthq/postgres-connector:0.0.14
	docker push aqueducthq/s3-connector:0.0.14
	docker push aqueducthq/snowflake-connector:0.0.14
	docker push aqueducthq/sqlserver-connector:0.0.14

.PHONY:

