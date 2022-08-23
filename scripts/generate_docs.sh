#!/bin/bash

rm -rf docs/
mkdir docs

echo "### package aqueduct
* [\`aqueduct_client\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.aqueduct_client)
* [\`decorator\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.decorator)
* [\`enums\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.enums)
* [\`error\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.error)
* [\`flow\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.flow)
* [\`operators\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.operators)
* [\`schedule\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.schedule)
### package aqueduct.artifacts
* [\`bool_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.bool_artifact)
* [\`generic_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.generic_artifact)
* [\`numeric_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.numeric_artifact)
* [\`param_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.param_artifact)
* [\`table_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.table_artifact)
### package aqueduct.constants
* [\`constants.exports\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.exports)
* [\`constants.metrics\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.metrics)
### package aqueduct.integrations
* [\`integrations.google_sheets_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.google_sheets_integration)
* [\`integrations.integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.integration)
* [\`integrations.s3_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.s3_integration)
* [\`integrations.salesforce_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.salesforce_integration)
* [\`integrations.sql_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.sql_integration)" > docs/README.md


pydoc-markdown -I . --render-toc -m aqueduct.aqueduct_client > docs/aqueduct.aqueduct_client.md
pydoc-markdown -I . --render-toc -m aqueduct.decorator > docs/aqueduct.decorator.md
pydoc-markdown -I . --render-toc -m aqueduct.enums > docs/aqueduct.enums.md
pydoc-markdown -I . --render-toc -m aqueduct.error > docs/aqueduct.error.md
pydoc-markdown -I . --render-toc -m aqueduct.flow > docs/aqueduct.flow.md
pydoc-markdown -I . --render-toc -m aqueduct.operators > docs/aqueduct.operators.md
pydoc-markdown -I . --render-toc -m aqueduct.schedule > docs/aqueduct.schedule.md

mkdir docs/package-aqueduct.artifacts

echo "### package aqueduct.artifacts
* [\`bool_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.bool_artifact)
* [\`generic_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.generic_artifact)
* [\`numeric_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.numeric_artifact)
* [\`param_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.param_artifact)
* [\`table_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.artifacts/aqueduct.artifacts.table_artifact)" > docs/package-aqueduct.artifacts/README.md

pydoc-markdown -I . --render-toc -m aqueduct.artifacts.bool_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.bool_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.generic_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.generic_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.numeric_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.numeric_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.param_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.param_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.artifacts.table_artifact > docs/package-aqueduct.artifacts/aqueduct.artifacts.table_artifact.md

mkdir docs/package-aqueduct.constants

echo "### package aqueduct.constants
* [\`constants.exports\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.exports)
* [\`constants.metrics\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.metrics)" > docs/package-aqueduct.constants/README.md

pydoc-markdown -I . --render-toc -m aqueduct.constants.exports > docs/package-aqueduct.constants/aqueduct.constants.exports.md
pydoc-markdown -I . --render-toc -m aqueduct.constants.metrics > docs/package-aqueduct.constants/aqueduct.constants.metrics.md

mkdir docs/package-aqueduct.integrations

echo "### package aqueduct.integrations
* [\`integrations.google_sheets_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.google_sheets_integration)
* [\`integrations.integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.integration)
* [\`integrations.s3_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.s3_integration)
* [\`integrations.salesforce_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.salesforce_integration)
* [\`integrations.sql_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.sql_integration)" > docs/package-aqueduct.integrations/README.md

pydoc-markdown -I . --render-toc -m aqueduct.integrations.google_sheets_integration > docs/package-aqueduct.integrations/aqueduct.integrations.google_sheets_integration.md
pydoc-markdown -I . --render-toc -m aqueduct.integrations.integration > docs/package-aqueduct.integrations/aqueduct.integrations.integration.md
pydoc-markdown -I . --render-toc -m aqueduct.integrations.s3_integration > docs/package-aqueduct.integrations/aqueduct.integrations.s3_integration.md
pydoc-markdown -I . --render-toc -m aqueduct.integrations.salesforce_integration > docs/package-aqueduct.integrations/aqueduct.integrations.salesforce_integration.md
pydoc-markdown -I . --render-toc -m aqueduct.integrations.sql_integration > docs/package-aqueduct.integrations/aqueduct.integrations.sql_integration.md
