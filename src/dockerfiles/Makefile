.DEFAULT_GOAL = all

# By default, turn off the Makefile practice of printing each command before
# you run it.
ifndef VERBOSE
.SILENT:
endif

build-function:
	docker build . -t aqueducthq/function:0.0.13 -f function/function.dockerfile --no-cache

build-param:
	docker build . -t aqueducthq/param:0.0.13 -f param/param.dockerfile --no-cache

build-system-metric:
	docker build . -t aqueducthq/system-metric:0.0.13 -f system-metric/system-metric.dockerfile --no-cache

build-base-connector:
	docker build . -t aqueducthq/base_connector:0.0.13 -f connectors/base.dockerfile --no-cache

build-connectors: build-athena-connector build-bigquery-connector build-mysql-connector \
	build-postgres-connector build-s3-connector build-snowflake-connector build-sqlserver-connector

build-athena-connector:
	docker build . -t aqueducthq/athena-connector:0.0.13 -f connectors/athena.dockerfile --no-cache

build-bigquery-connector:
	docker build . -t aqueducthq/bigquery-connector:0.0.13 -f connectors/bigquery.dockerfile --no-cache

build-mysql-connector:
	docker build . -t aqueducthq/mysql-connector:0.0.13 -f connectors/mysql.dockerfile --no-cache

build-postgres-connector:
	docker build . -t aqueducthq/postgres-connector:0.0.13 -f connectors/postgres.dockerfile --no-cache

build-s3-connector:
	docker build . -t aqueducthq/s3-connector:0.0.13 -f connectors/s3.dockerfile --no-cache

build-snowflake-connector:
	docker build . -t aqueducthq/snowflake-connector:0.0.13 -f connectors/snowflake.dockerfile --no-cache

build-sqlserver-connector:
	docker build . -t aqueducthq/sqlserver-connector:0.0.13 -f connectors/sqlserver.dockerfile --no-cache

build-lambda-images: build-lambda-function build-lambda-param build-lambda-system-metric build-lambda-snowflake

build-lambda-function:
	docker build . -t hsubbaraj/lambda-function -f lambda/function/function.dockerfile --no-cache

build-lambda-param:
	docker build . -t hsubbaraj/lambda-param -f lambda/param/param.dockerfile --no-cache

build-lambda-system-metric:
	docker build . -t hsubbaraj/lambda-system-metric -f lambda/system-metric/system-metric.dockerfile --no-cache

build-lambda-connectors: build-lambda-snowflake build-lambda-athena build-lambda-bigquery build-lambda-postgres build-lambda-s3

build-lambda-snowflake:
	docker build . -t hsubbaraj/lambda-snowflake -f lambda/connectors/snowflake.dockerfile --no-cache

build-lambda-athena:
	docker build . -t aqueducthq/lambda-athena -f lambda/connectors/athena.dockerfile --no-cache

build-lambda-bigquery:
	docker build . -t aqueducthq/lambda-bigquery -f lambda/connectors/bigquery.dockerfile --no-cache

build-lambda-postgres:
	docker build . -t aqueducthq/lambda-postgres -f lambda/connectors/postgres.dockerfile --no-cache

build-lambda-s3:
	docker build . -t aqueducthq/lambda-s3 -f lambda/connectors/s3.dockerfile --no-cache


publish-function:
	docker push aqueducthq/function:0.0.13

publish-param:
	docker push aqueducthq/param:0.0.13

publish-system-metric:
	docker push aqueducthq/system-metric:0.0.13

publish-connectors:
	docker push aqueducthq/athena-connector:0.0.13
	docker push aqueducthq/bigquery-connector:0.0.13
	docker push aqueducthq/mysql-connector:0.0.13
	docker push aqueducthq/postgres-connector:0.0.13
	docker push aqueducthq/s3-connector:0.0.13
	docker push aqueducthq/snowflake-connector:0.0.13
	docker push aqueducthq/sqlserver-connector:0.0.13

publish-lambda:
	docker push hsubbaraj/lambda-function
	docker push hsubbaraj/lambda-snowflake
	docker push hsubbaraj/lambda-param
	docker push hsubbaraj/lambda-system-metric

.PHONY:

