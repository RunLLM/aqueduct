[<img src="https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct-logo-two-tone/1x/aqueduct-logo-two-tone-1x.png" width= "35%" />](https://www.aqueducthq.com)

## Aqueduct: Taking Data Science to Production

[![Downloads](https://pepy.tech/badge/aqueduct-ml/month)](https://pypi.org/project/aqueduct-ml/)
[![Slack](https://img.shields.io/static/v1.svg?label=chat&message=on%20slack&color=27b1ff&style=flat)](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A)
[![GitHub license](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/aqueducthq/aqueduct/blob/master/LICENSE)
[![PyPI version](https://badge.fury.io/py/aqueduct-ml.svg)](https://pypi.org/project/aqueduct-ml/)
[![Tests](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml)

**Aqueduct automates the engineering required to take data science & machine learning projects to production**.

With Aqueduct, you can define & deploy machine learning pipelines, connect your code to data and business systems, and monitor the performance and quality of your pipelines -- all using a simple Python API. 

You can install Aqueduct via `pip`:
```bash
pip3 install aqueduct-ml
aqueduct start
```

Once you have Aqueduct running, you can create your first workflow:

```python
from aqueduct import Client, op, metric

client = Client()

@op
def transform_data(reviews):
    reviews['strlen'] = reviews['review'].str.len()
    return reviews


demo_db = client.integration("aqueduct_demo")
reviews_table = demo_db.sql("select * from hotel_reviews;")

strlen_table = transform_data(reviews_table)
strlen_table.save(demo_db.config(table="strlen_table", update_mode="replace")) 

client.publish_flow(name="review_strlen", artifacts=[strlen_table])
```

Once you've created a workflow, you can view that workflow in the Aqueduct UI: 

![image](https://user-images.githubusercontent.com/867892/196529730-3c9582d5-8692-495d-a7df-8eb62ddf305f.png)

## Why Aqueduct?

The engineering required to get data science & machine learning projects in production slows down data teams. Aqueduct automates away that engineering and allows you to define robust data & ML pipelines in a few lines of code and run them anywhere.

* **Simple API**: Get a production-ready pipeline running in a few lines of vanilla Python.
* **Flexible Environments**: Run your pipelines anywhere (locally or in the cloud). You can deploy & update your models on your own infrastructure without building custom Docker containers, managing Kubernetes deployments, or storing passwords in plaintext.
* **Data Integrations**: With out-of-the-box connectors, you can access the freshest data easily & reliably.
* **Custom Monitoring**: Aqueduct's checks and metrics allow you to define contraints on your workflows, so you can quickly debug and fix errors. 

## Overview & Examples

Aqueduct allow you to build powerful machine learning workflows that **run anywhere, publish predictions everywhere, and ensure prediction quality**.
The core abstraction in Aqueduct is a [Workflow](https://docs.aqueducthq.com/workflows), which is a sequence of [Artifacts](https://docs.aqueducthq.com/artifacts) (data) that are transformed by [Operators](https://docs.aqueducthq.com/operators) (compute). 
The input Artifact(s) for a Workflow is typically loaded from a database, and the output Artifact(s) are typically persisted back to a database. 
Each Workflow can either be run on a fixed schedule or triggered on-demand.

To see Aqueduct in action on some real-world machine learning workflows, check out some of our examples:

* [Churn Ensemble](https://github.com/aqueducthq/aqueduct/blob/main/examples/churn_prediction/Build%20and%20Deploy%20Churn%20Ensemble.ipynb)
* [Sentiment Analysis](https://github.com/aqueducthq/aqueduct/blob/main/examples/sentiment_analysis/Sentiment%20Model.ipynb)
* [Impute Missing Wine Data](https://github.com/aqueducthq/aqueduct/blob/main/examples/training_and_inference/Impute%20Missing%20Wine%20Data.ipynb)
* ... [and more](https://github.com/aqueducthq/aqueduct/tree/main/examples)!

## What's next?

Check out our [documentation](https://docs.aqueducthq.com/), where you'll find:
* a [Quickstart Guide](https://docs.aqueducthq.com/quickstart-guide)
* [example workflows](https://docs.aqueducthq.com/example-workflows)
* and more details on [creating workflows](https://docs.aqueducthq.com/workflows)

If you have questions or comments or would like to learn more about what we're
building, please [reach out](mailto:hello@aqueducthq.com), [join our Slack
channel](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A), or [start a conversation on GitHub](https://github.com/aqueducthq/aqueduct/issues/new).
We'd love to hear from you!

If you're interested in contributing, please check out our [roadmap](https://github.com/aqueducthq/aqueduct/wiki/Aqueduct-Roadmap) and join the development channel in [our community Slack](https://slack.aqueducthq.com).
