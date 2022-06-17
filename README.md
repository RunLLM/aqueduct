[<img src="https://user-images.githubusercontent.com/867892/172955552-f1f29c80-713f-41e9-af0c-7d7c8ee622f0.jpg" width= "35%" />](https://www.aqueducthq.com)

## Aqueduct: Prediction Infrastructure for Data Scientists

[![Downloads](https://pepy.tech/badge/aqueduct-ml/month)](https://pypi.org/project/aqueduct-ml/)
[![Slack](https://img.shields.io/static/v1.svg?label=chat&message=on%20slack&color=27b1ff&style=flat)](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A)
[![GitHub license](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/aqueducthq/aqueduct/blob/master/LICENSE)
[![PyPI version](https://badge.fury.io/py/aqueduct-ml.svg)](https://pypi.org/project/aqueduct-ml/)
[![Tests](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml)

With Aqueduct, data scientists can instantaneously deploy machine learning models to the cloud, connect those models to data and business systems, and gain visibility into the performance of their prediction pipelines -- all from the comfort of a Python notebook. 

The core abstraction in Aqueduct is a [Workflow](https://docs.aqueducthq.com/workflows), which is a sequence of [Artifacts](https://docs.aqueducthq.com/artifacts) (data) that are transformed by [Operators](https://docs.aqueducthq.com/operators) (compute). 
The input Artifact(s) for a Workflow is typically loaded from a database, and the output Artifact(s) are typically persisted back to a database. 
Each Workflow can either be run on a fixed schedule or triggered on-demand. 

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

Once you have the Aqueduct server running, this 25-line code snippet is all you need to create your first prediction pipeline:

```python
import aqueduct as aq
from aqueduct import op, metric
import pandas as pd
# Need to install torch and transformers
#!pip install torch transformers
from transformers import pipeline
import torch

client = aq.Client("YOUR_API_KEY", "localhost:8080")

# This function takes in a DataFrame with the text of user review of
# hotels and returns a DataFrame that has the sentiment of the review.
# This function users the `pipeline` interface from HuggingFace's 
# Transformers package. 
@op()
def sentiment_prediction(reviews):
    model = pipeline("sentiment-analysis")
    predicted_sentiment = model(list(reviews["review"]))
    return reviews.join(pd.DataFrame(predicted_sentiment))

# Load a connection to a database -- here, we use the `aqueduct_demo`
# database, for which you can find the documentation here:
# https://docs.aqueducthq.com/example-workflows/demo-data-warehouse
demo_db = client.integration("aqueduct_demo")

# Once we have a connection to a database, we can run a SQL query against it.
reviews_table = demo_db.sql("select * from hotel_reviews;")

# Next, we apply our annotated function to our data -- this tells Aqueduct 
# to create a workflow spec that applied `sentiment_prediction` to `reviews_table`.
sentiment_table = sentiment_prediction(reviews_table)

# When we call `.save()`, Aqueduct will take the data in `sentiment_table` and 
# write the results back to any database you specify -- in this case, back to the 
# `aqueduct_demo` DB.
sentiment_table.save(demo_db.config(table="sentiment_pred", update_mode="replace"))

# In Aqueduct, a metric is a numerical measurement of a some predictions. Here, 
# we calculate the average sentiment score returned by our machine learning 
# model, which is something we can track over time.


# In Aqueduct, a metric is a numerical measurement of a some predictions. Here, 
# we calculate the average sentiment score returned by our machine learning 
# model, which is something we can track over time.
@metric
def average_sentiment(reviews_with_sent):
    return (reviews_with_sent["label"] == "POSITIVE").mean()

avg_sent = average_sentiment(sentiment_table)

# Once we compute a metric, we can set upper and lower bounds on it -- if 
# the metric exceeds one of those bounds, an error will be raised.
avg_sent.bound(lower=0.5)

# And we're done! With a call to `publish_flow`, we've created a full workflow
# that calculates the sentiment of hotel reviews, creates a metric over those
# predictions, and sets a bound on that metric.
client.publish_flow(name="hotel_sentiment", artifacts=[sentiment_table, avg_sent])
```

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
