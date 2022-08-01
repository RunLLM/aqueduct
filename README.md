[<img src="https://user-images.githubusercontent.com/867892/172955552-f1f29c80-713f-41e9-af0c-7d7c8ee622f0.jpg" width= "35%" />](https://www.aqueducthq.com)

## Aqueduct: A Production Data Science Platform

[![Downloads](https://pepy.tech/badge/aqueduct-ml/month)](https://pypi.org/project/aqueduct-ml/)
[![Slack](https://img.shields.io/static/v1.svg?label=chat&message=on%20slack&color=27b1ff&style=flat)](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A)
[![GitHub license](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/aqueducthq/aqueduct/blob/master/LICENSE)
[![PyPI version](https://badge.fury.io/py/aqueduct-ml.svg)](https://pypi.org/project/aqueduct-ml/)
[![Tests](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml)

**Aqueduct automates the engineering required to make data science operational**. 
With Aqueduct, data scientists can instantaneously deploy machine learning models to the cloud, connect those models to data and business systems, and gain visibility into everything from inference latency to model accuracy -- all with Python. 

Check out our [docs](https://docs.aqueducthq.com), [ask us anything](https://slack.aqueducthq.com), and [share your feedback](https://github.com/aqueducthq/aqueduct/issues/new/choose)!

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

<img width="2160" alt="image" src="https://user-images.githubusercontent.com/867892/176579763-6f77fcc0-8b12-446b-ab9a-96095c6d1b5f.png">

You can run the full Aqueduct server in a Google Colab notebook [here](https://colab.research.google.com/drive/1EyKTF9tXjgnlBHVQzgt5Yr79e_8ef27M). Our [`examples`](examples/) directory has a few, more detailed prediction pipelines:

* [Churn Ensemble](https://github.com/aqueducthq/aqueduct/blob/main/examples/churn_prediction/Build%20and%20Deploy%20Churn%20Ensemble.ipynb)
* [Sentiment Analysis](https://github.com/aqueducthq/aqueduct/blob/main/examples/sentiment_analysis/Sentiment%20Model.ipynb)
* [Impute Missing Wine Data](https://github.com/aqueducthq/aqueduct/blob/main/examples/training_and_inference/Training%20and%20Inference%20in%20a%20Single%20Workflow.ipynb)
* more coming soon!


## Getting Started

To get started with Aqueduct:
1. Ensure that you meet the [basic requirements](https://docs.aqueducthq.com/installation-and-deployment/installing-aqueduct).
2. Install the aqueduct server and UI by running: 
    ```bash
    pip3 install aqueduct-ml
    ```
3. Launch both the server and the UI by running: 
    ```bash
    aqueduct start
    ```
4. Get your API Key by running:
    ```bash
    aqueduct apikey
    ```

The core abstraction in Aqueduct is a [Workflow](https://docs.aqueducthq.com/workflows), which is a sequence of [Artifacts](https://docs.aqueducthq.com/artifacts) (data) that are transformed by [Operators](https://docs.aqueducthq.com/operators) (compute). 
The input Artifact(s) for a Workflow is typically loaded from a database, and the output Artifact(s) are typically persisted back to a database. 
Each Workflow can either be run on a fixed schedule or triggered on-demand. 

## Why Aqueduct?

The existing tools for deploying models are not designed with data scientists in mind -- they assume the user will casually build Docker containers, deploy Kubernetes clusters, and writes thousands of lines of YAML to deploy a single model. 
Data scientists are by and large not interested in doing that, and there are better uses for their skills.

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
