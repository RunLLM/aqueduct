#!/bin/bash

rm -rf docs/
mkdir docs

echo "### package aqueduct
* [\`client\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.client)
* [\`decorator\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.decorator)
* [\`error\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.error)
* [\`flow\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.flow)
* [\`schedule\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.schedule)
* [\`llm_op\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.llm_op)
### package aqueduct.artifacts
* [\`bool_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.bool_artifact)
* [\`generic_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.generic_artifact)
* [\`numeric_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.numeric_artifact)
* [\`table_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.table_artifact)
### package aqueduct.constants
* [\`constants.exports\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.exports)
* [\`constants.metrics\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.metrics)
* [\`constants.enums\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.enums)
### package aqueduct.models
* [\`models.resource\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.models/aqueduct.models.resource)
* [\`models.operators\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.models/aqueduct.models.operators)
### package aqueduct.resources
* [\`resources.airflow\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.airflow)
* [\`resources.aws_lambda\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.aws_lambda)
* [\`resources.databricks\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.databricks)
* [\`resources.dynamic\_k8s\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.dynamic\_k8s)
* [\`resources.ecr\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.ecr)
* [\`resources.k8s\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.k8s)
* [\`resources.google\_sheets\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.google\_sheets)
* [\`resources.mongodb\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.mongodb)
* [\`resources.s3\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.s3)
* [\`resources.salesforce\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.salesforce)
* [\`resources.spark\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.spark)
* [\`resources.sql\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.sql)" > docs/README.md


pydoc-markdown -I . --render-toc -m aqueduct.client > docs/aqueduct.client.md
pydoc-markdown -I . --render-toc -m aqueduct.decorator > docs/aqueduct.decorator.md
pydoc-markdown -I . --render-toc -m aqueduct.error > docs/aqueduct.error.md
pydoc-markdown -I . --render-toc -m aqueduct.flow > docs/aqueduct.flow.md
pydoc-markdown -I . --render-toc -m aqueduct.schedule > docs/aqueduct.schedule.md
pydoc-markdown -I . --render-toc -m aqueduct.llm_wrapper > docs/aqueduct.llm_op.md

mkdir docs/package-aqueduct.artifacts

echo "### package aqueduct.artifacts
* [\`bool_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.bool_artifact)
* [\`generic_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.generic_artifact)
* [\`numeric_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.numeric_artifact)
* [\`table_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.table_artifact)" > docs/package-aqueduct.artifacts/README.md

pydoc-markdown -I . --render-toc -m aqueduct.artifacts.bool_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.bool_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.generic_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.generic_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.numeric_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.numeric_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.table_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.table_artifact.md

mkdir docs/package-aqueduct.constants

echo "### package aqueduct.constants
* [\`constants.exports\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.exports)
* [\`constants.metrics\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.metrics)
* [\`constants.enums\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.enums)" > docs/package-aqueduct.constants/README.md

pydoc-markdown -I . --render-toc -m aqueduct.constants.exports > docs/package-aqueduct.constants/aqueduct.constants.exports.md
pydoc-markdown -I . --render-toc -m aqueduct.constants.metrics > docs/package-aqueduct.constants/aqueduct.constants.metrics.md
pydoc-markdown -I . --render-toc -m aqueduct.constants.enums > docs/package-aqueduct.constants/aqueduct.constants.enums.md

mkdir docs/package-aqueduct.models

echo "### package aqueduct.models
* [\`models.resource\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.models/aqueduct.models.resource)
* [\`models.operators\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.models/aqueduct.models.operators)" > docs/package-aqueduct.models/README.md

pydoc-markdown -I . --render-toc -m aqueduct.models.resource > docs/package-aqueduct.models/aqueduct.models.resource.md
pydoc-markdown -I . --render-toc -m aqueduct.models.operators > docs/package-aqueduct.models/aqueduct.models.operators.md

mkdir docs/package-aqueduct.resources

echo "### package aqueduct.resources
* [\`resources.airflow\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.airflow)
* [\`resources.aws_lambda\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.aws_lambda)
* [\`resources.databricks\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.databricks)
* [\`resources.dynamic\_k8s\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.dynamic\_k8s)
* [\`resources.ecr\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.ecr)
* [\`resources.k8s\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.k8s)
* [\`resources.google\_sheets\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.google\_sheets)
* [\`resources.mongodb\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.mongodb)
* [\`resources.s3\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.s3)
* [\`resources.salesforce\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.salesforce)
* [\`resources.spark\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.spark)
* [\`resources.sql\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.resources/aqueduct.resources.sql)" > docs/package-aqueduct.resources/README.md

pydoc-markdown -I . --render-toc -m aqueduct.resources.airflow > docs/package-aqueduct.resources/aqueduct.resources.airflow.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.aws_lambda > docs/package-aqueduct.resources/aqueduct.resources.aws_lambda.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.databricks > docs/package-aqueduct.resources/aqueduct.resources.databricks.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.dynamic_k8s > docs/package-aqueduct.resources/aqueduct.resources.dynamic_k8s.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.ecr > docs/package-aqueduct.resources/aqueduct.resources.ecr.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.k8s > docs/package-aqueduct.resources/aqueduct.resources.k8s.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.google_sheets > docs/package-aqueduct.resources/aqueduct.resources.google_sheets.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.mongodb > docs/package-aqueduct.resources/aqueduct.resources.mongodb.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.s3 > docs/package-aqueduct.resources/aqueduct.resources.s3.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.salesforce > docs/package-aqueduct.resources/aqueduct.resources.salesforce.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.spark > docs/package-aqueduct.resources/aqueduct.resources.spark.md
pydoc-markdown -I . --render-toc -m aqueduct.resources.sql > docs/package-aqueduct.resources/aqueduct.resources.sql.md
