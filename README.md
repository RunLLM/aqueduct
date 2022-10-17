[<img src="https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct-logo-two-tone/1x/aqueduct-logo-two-tone-1x.png" width= "35%" />](https://www.aqueducthq.com)

## Aqueduct: A Production Data Science Platform

[![Downloads](https://pepy.tech/badge/aqueduct-ml/month)](https://pypi.org/project/aqueduct-ml/)
[![Slack](https://img.shields.io/static/v1.svg?label=chat&message=on%20slack&color=27b1ff&style=flat)](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A)
[![GitHub license](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/aqueducthq/aqueduct/blob/master/LICENSE)
[![PyPI version](https://badge.fury.io/py/aqueduct-ml.svg)](https://pypi.org/project/aqueduct-ml/)
[![Tests](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml)

**Aqueduct automates the engineering required to make data science operational**. 
With Aqueduct, data scientists can instantaneously deploy machine learning models to the cloud, connect those models to data and business systems, and gain visibility into everything from inference latency to model accuracy -- all with Python. 

Check out our [docs](https://docs.aqueducthq.com), [ask us anything](https://slack.aqueducthq.com), and [share your feedback](https://github.com/aqueducthq/aqueduct/issues/new/choose)!

To get started with Aqueduct, you need to run two lines in your terminal:
```bash
pip3 install aqueduct-ml
aqueduct start
```

Once you have Aqueduct running, we can create our first workflow:

```python
from aqueduct import Client, op, metric, get_apikey

client = Client(get_apikey(), "localhost:8080")

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

<img width="2160" alt="image" src="https://user-images.githubusercontent.com/867892/183779415-4e42b9b9-e4f3-491a-a2c2-fc0028faa236.png">

## Overview & Examples

Aqueduct is built to allow you to write regular Python code and compose powerful machine learning workflows that **run anywhere, publish predictions everywhere, and ensure prediction quality**.
The core abstraction in Aqueduct is a [Workflow](https://docs.aqueducthq.com/workflows), which is a sequence of [Artifacts](https://docs.aqueducthq.com/artifacts) (data) that are transformed by [Operators](https://docs.aqueducthq.com/operators) (compute). 
The input Artifact(s) for a Workflow is typically loaded from a database, and the output Artifact(s) are typically persisted back to a database. 
Each Workflow can either be run on a fixed schedule or triggered on-demand.

To see Aqueduct in action on some real-world machine learning workflows, check out some of our examples:

* [Churn Ensemble](https://github.com/aqueducthq/aqueduct/blob/main/examples/churn_prediction/Build%20and%20Deploy%20Churn%20Ensemble.ipynb)
* [Sentiment Analysis](https://github.com/aqueducthq/aqueduct/blob/main/examples/sentiment_analysis/Sentiment%20Model.ipynb)
* [Impute Missing Wine Data](https://github.com/aqueducthq/aqueduct/blob/main/examples/training_and_inference/Impute%20Missing%20Wine%20Data.ipynb)
* ... [and more](https://github.com/aqueducthq/aqueduct/tree/main/examples)!

## Why Aqueduct?

The existing (MLOps) tools for deploying models are not designed with data teams in mind -- they are designed, built, and operated by software engineers with years of cloud infrastructure experience.
Rather than abstracting away the repetitive engineering work required to operationalize models, MLOps tools expect data teams to spend their time building Docker containers and managing Kubernetes deployments by hand.
We don't believe that's an efficient use of anyone's time.

Aqueduct is designed for data scientists, with three core design principles in mind:
* *Simplicity*: Data scientists should be able to deploy models with tools they're comfortable with and without having to learn how to use complex, low-level infrastructure systems.
* *Connectedness*: Data science and machine learning can have the greatest impact when everyone in the business has access, and data scientists shouldn't have to bend over backwards to make this happen.
* *Confidence*: Having the whole organization benefit from your work means that data scientists should be able to sleep peacefully, knowing that things are working as expected -- and they'll be alerted as soon as that changes.

## What's next?

Interested in learning more? Check out our [documentation](https://docs.aqueducthq.com/), where you'll find:
* a [Quickstart Guide](https://docs.aqueducthq.com/quickstart-guide)
* [example workflows](https://docs.aqueducthq.com/example-workflows)
* and more details on [creating workflows](https://docs.aqueducthq.com/workflows)

If you have questions or comments or would like to learn more about what we're
building, please [reach out](mailto:hello@aqueducthq.com), [join our Slack
channel](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A), or [start a conversation on GitHub](https://github.com/aqueducthq/aqueduct/issues/new).
We'd love to hear from you!
