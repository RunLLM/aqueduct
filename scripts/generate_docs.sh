#!/bin/bash

rm -rf docs/
mkdir docs

echo "### package aqueduct
* [\`artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.artifact)
* [\`aqueduct_client\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.aqueduct_client)
* [\`check_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.check_artifact)
* [\`decorator\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.decorator)
* [\`enums\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.enums)
* [\`error\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.error)
* [\`flow\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.flow)
* [\`generic_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.generic_artifact)
* [\`metric_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.metric_artifact)
* [\`operators\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.operators)
* [\`schedule\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.schedule)
* [\`table_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.table_artifact)
* [\`param_artifact\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/aqueduct.param_artifact)
###\ package aqueduct.constants
* [\`constants.exports\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.exports)
* [\`constants.metrics\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.constants/aqueduct.constants.metrics)
###\ package aqueduct.integrations
* [\`integrations.google_sheets_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.google_sheets_integration)
* [\`integrations.integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.integration)
* [\`integrations.s3_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.s3_integration)
* [\`integrations.salesforce_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.salesforce_integration)
* [\`integrations.sql_integration\`](https://docs.aqueducthq.com/api-reference/sdk-reference/package-aqueduct/package-aqueduct.integrations/aqueduct.integrations.sql_integration)" > docs/README.md


pydoc-markdown -I . --render-toc -m aqueduct.artifact > docs/aqueduct.artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.aqueduct_client > docs/aqueduct.aqueduct_client.md
pydoc-markdown -I . --render-toc -m aqueduct.check_artifact > docs/aqueduct.check_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.decorator > docs/aqueduct.decorator.md
pydoc-markdown -I . --render-toc -m aqueduct.enums > docs/aqueduct.enums.md
pydoc-markdown -I . --render-toc -m aqueduct.error > docs/aqueduct.error.md
pydoc-markdown -I . --render-toc -m aqueduct.flow > docs/aqueduct.flow.md
pydoc-markdown -I . --render-toc -m aqueduct.generic_artifact > docs/aqueduct.generic_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.metric_artifact > docs/aqueduct.metric_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.operators > docs/aqueduct.operators.md
pydoc-markdown -I . --render-toc -m aqueduct.schedule > docs/aqueduct.schedule.md
pydoc-markdown -I . --render-toc -m aqueduct.table_artifact > docs/aqueduct.table_artifact.md
pydoc-markdown -I . --render-toc -m aqueduct.param_artifact > docs/aqueduct.param_artifact.md

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
