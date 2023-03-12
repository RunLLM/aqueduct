.DEFAULT_GOAL = all
VERSION = 0.2.5

# By default, turn off the Makefile practice of printing each command before
# you run it.
ifndef VERBOSE
.SILENT:
endif

# Building Aqueduct Images
build-aqueduct-images:
	docker build . -t aqueducthq/aqueduct-py37:$(VERSION) -f aqueduct/aqueduct-py37.dockerfile --no-cache --build-arg version=$(VERSION)
	docker build . -t aqueducthq/aqueduct-py38:$(VERSION) -f aqueduct/aqueduct-py38.dockerfile --no-cache --build-arg version=$(VERSION)
	docker build . -t aqueducthq/aqueduct-py39:$(VERSION) -f aqueduct/aqueduct-py39.dockerfile --no-cache --build-arg version=$(VERSION)
	docker build . -t aqueducthq/aqueduct-py310:$(VERSION) -f aqueduct/aqueduct-py310.dockerfile --no-cache --build-arg version=$(VERSION)

# Building K8s Images
build-k8s-system-images: build-function build-param build-system-metric build-base-connector build-connectors

build-function:
	docker build . -t aqueducthq/function37:$(VERSION) -f function/function37.dockerfile --no-cache
	docker build . -t aqueducthq/function38:$(VERSION) -f function/function38.dockerfile --no-cache
	docker build . -t aqueducthq/function39:$(VERSION) -f function/function39.dockerfile --no-cache
	docker build . -t aqueducthq/function310:$(VERSION) -f function/function310.dockerfile --no-cache

build-param:
	docker build . -t aqueducthq/param:$(VERSION) -f param/param.dockerfile --no-cache

build-system-metric:
	docker build . -t aqueducthq/system-metric:$(VERSION) -f system-metric/system-metric.dockerfile --no-cache

build-base-connector:
	docker build . -t aqueducthq/base_connector:$(VERSION) -f connectors/base.dockerfile --no-cache

build-connectors: build-athena-connector build-bigquery-connector build-mysql-connector \
	build-postgres-connector build-s3-connector build-snowflake-connector build-sqlserver-connector

build-athena-connector:
	docker build . -t aqueducthq/athena-connector:$(VERSION) -f connectors/athena.dockerfile --no-cache

build-bigquery-connector:
	docker build . -t aqueducthq/bigquery-connector:$(VERSION) -f connectors/bigquery.dockerfile --no-cache

build-mysql-connector:
	docker build . -t aqueducthq/mysql-connector:$(VERSION) -f connectors/mysql.dockerfile --no-cache

build-postgres-connector:
	docker build . -t aqueducthq/postgres-connector:$(VERSION) -f connectors/postgres.dockerfile --no-cache

build-s3-connector:
	docker build . -t aqueducthq/s3-connector:$(VERSION) -f connectors/s3.dockerfile --no-cache

build-snowflake-connector:
	docker build . -t aqueducthq/snowflake-connector:$(VERSION) -f connectors/snowflake.dockerfile --no-cache

build-sqlserver-connector:
	docker build . -t aqueducthq/sqlserver-connector:$(VERSION) -f connectors/sqlserver.dockerfile --no-cache

# Publishing K8s Images
publish-aqueduct:
	docker push aqueducthq/aqueduct-py37:$(VERSION)
	docker push aqueducthq/aqueduct-py38:$(VERSION)
	docker push aqueducthq/aqueduct-py39:$(VERSION)
	docker push aqueducthq/aqueduct-py310:$(VERSION)

publish-k8s: publish-function publish-param publish-system-metric publish-connectors

publish-function:
	docker push aqueducthq/function37:$(VERSION)
	docker push aqueducthq/function38:$(VERSION)
	docker push aqueducthq/function39:$(VERSION)
	docker push aqueducthq/function310:$(VERSION)

publish-param:
	docker push aqueducthq/param:$(VERSION)

publish-system-metric:
	docker push aqueducthq/system-metric:$(VERSION)

publish-connectors:
	docker push aqueducthq/athena-connector:$(VERSION)
	docker push aqueducthq/bigquery-connector:$(VERSION)
	docker push aqueducthq/mysql-connector:$(VERSION)
	docker push aqueducthq/postgres-connector:$(VERSION)
	docker push aqueducthq/s3-connector:$(VERSION)
	docker push aqueducthq/snowflake-connector:$(VERSION)
	docker push aqueducthq/sqlserver-connector:$(VERSION)

# Building Lambda Images
build-lambda-images: build-lambda-function build-lambda-param build-lambda-system-metric build-lambda-connectors

build-lambda-function:
	docker build . -t aqueducthq/lambda-function-37:$(VERSION) -f lambda/function/function37.dockerfile --no-cache
	docker build . -t aqueducthq/lambda-function-38:$(VERSION) -f lambda/function/function38.dockerfile --no-cache
	docker build . -t aqueducthq/lambda-function-39:$(VERSION) -f lambda/function/function39.dockerfile --no-cache

build-lambda-param:
	docker build . -t aqueducthq/lambda-param:$(VERSION) -f lambda/param/param.dockerfile --no-cache

build-lambda-system-metric:
	docker build . -t aqueducthq/lambda-system-metric:$(VERSION) -f lambda/system-metric/system-metric.dockerfile --no-cache

build-lambda-connectors: build-lambda-snowflake build-lambda-athena build-lambda-bigquery build-lambda-postgres build-lambda-s3

build-lambda-snowflake:
	docker build . -t aqueducthq/lambda-snowflake-connector:$(VERSION) -f lambda/connectors/snowflake.dockerfile --no-cache

build-lambda-athena:
	docker build . -t aqueducthq/lambda-athena-connector:$(VERSION) -f lambda/connectors/athena.dockerfile --no-cache

build-lambda-bigquery:
	docker build . -t aqueducthq/lambda-bigquery-connector:$(VERSION) -f lambda/connectors/bigquery.dockerfile --no-cache

build-lambda-postgres:
	docker build . -t aqueducthq/lambda-postgres-connector:$(VERSION) -f lambda/connectors/postgres.dockerfile --no-cache

build-lambda-s3:
	docker build . -t aqueducthq/lambda-s3-connector:$(VERSION) -f lambda/connectors/s3.dockerfile --no-cache

# Publishing Lambda Images
publish-lambda:
	docker push aqueducthq/lambda-function-37:$(VERSION)
	docker push aqueducthq/lambda-function-38:$(VERSION)
	docker push aqueducthq/lambda-function-39:$(VERSION)
	docker push aqueducthq/lambda-param:$(VERSION)
	docker push aqueducthq/lambda-system-metric:$(VERSION)
	docker push aqueducthq/lambda-athena-connector:$(VERSION)
	docker push aqueducthq/lambda-bigquery-connector:$(VERSION)
	docker push aqueducthq/lambda-postgres-connector:$(VERSION)
	docker push aqueducthq/lambda-s3-connector:$(VERSION)
	docker push aqueducthq/lambda-snowflake-connector:$(VERSION)

# Building GPU Images
build-function-gpu:
	docker build . -t aqueducthq/function37gpu:$(VERSION) -f gpu/function37gpu.dockerfile
	docker build . -t aqueducthq/function38gpu:$(VERSION) -f gpu/function38gpu.dockerfile
	docker build . -t aqueducthq/function39gpu:$(VERSION) -f gpu/function39gpu.dockerfile
	docker build . -t aqueducthq/function310gpu:$(VERSION) -f gpu/function310gpu.dockerfile

publish-gpu:
	docker push aqueducthq/function37gpu:$(VERSION)
	docker push aqueducthq/function38gpu:$(VERSION)
	docker push aqueducthq/function39gpu:$(VERSION)
	docker push aqueducthq/function310gpu:$(VERSION)

# Building Spark Conda env images
build-spark-conda:
	docker build . -t aqueducthq/spark-py37-env:$(VERSION) -f spark/spark-py37-env.dockerfile
	docker build . -t aqueducthq/spark-py38-env:$(VERSION) -f spark/spark-py38-env.dockerfile
	docker build . -t aqueducthq/spark-py39-env:$(VERSION) -f spark/spark-py39-env.dockerfile
	docker build . -t aqueducthq/spark-py310-env:$(VERSION) -f spark/spark-py310-env.dockerfile

publish-spark-conda:
	docker push aqueducthq/spark-py37-env:$(VERSION)
	docker push aqueducthq/spark-py38-env:$(VERSION)
	docker push aqueducthq/spark-py39-env:$(VERSION)
	docker push aqueducthq/spark-py310-env:$(VERSION)

.PHONY:

