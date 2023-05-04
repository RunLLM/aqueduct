#!/bin/bash

rm -rf docs/
mkdir docs

echo "### package aqueduct_llm
* [\`dolly_v2_3b\`](https://docs.aqueducthq.com/api-reference/aqueduct-llm-reference/package-aqueduct-llm/aqueduct_llm.dolly_v2_3b)
* [\`dolly_v2_7b\`](https://docs.aqueducthq.com/api-reference/aqueduct-llm-reference/package-aqueduct-llm/aqueduct_llm.dolly_v2_7b)
* [\`llama_7b\`](https://docs.aqueducthq.com/api-reference/aqueduct-llm-reference/package-aqueduct-llm/aqueduct_llm.llama_7b)
* [\`vicuna_7b\`](https://docs.aqueducthq.com/api-reference/aqueduct-llm-reference/package-aqueduct-llm/aqueduct_llm.vicuna_7b)" > docs/README.md

pydoc-markdown -I . --render-toc -m aqueduct_llm.dolly_v2_3b > docs/aqueduct_llm.dolly_v2_3b.md
pydoc-markdown -I . --render-toc -m aqueduct_llm.dolly_v2_7b > docs/aqueduct_llm.dolly_v2_7b.md
pydoc-markdown -I . --render-toc -m aqueduct_llm.llama_7b > docs/aqueduct_llm.llama_7b.md
pydoc-markdown -I . --render-toc -m aqueduct_llm.vicuna_7b > docs/aqueduct_llm.vicuna_7b.md