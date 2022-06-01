# Aqueduct: Prediction Infrastructure for Data Scientists

Aqueduct is open-source prediction infrastructure built for data scientists, by data scientists. 
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
3. Launch the server by running: 
    ```bash
    aqueduct server &
    ```
4. Launch the web-ui by running:
    ```bash
    aqueduct ui &
    ```
5. Get your API Key by running:
    ```bash
    aqueduct apikey
    ```

Once you have the Aqueduct server running, this 25-line code snippet is all you need to create your first prediction pipeline:

```python
import aqueduct as aq
from aqueduct import op, metric
import pandas as pd
from transformers import pipeline
import torch

client = aq.Client("YOUR_API_KEY", "localhost:8080")

demo_db = client.integration("aqueduct_demo/")
reviews_table = demo_db.sql("select * from hotel_reviews;")

@op()
def sentiment_prediction(reviews):
    model = pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews['review']))))

sentiment_table = sentiment_prediction(reviews_table)
sentiment_table.save(demo_db.config(table='sentiment_pred', update_mode='replace'))

@metric
def average_sentiment(reviews_with_sent):
    return reviews_with_sent['review'].mean()
avg_sent = average_sentiment(sentiment_table)
avg_sent.bound(lower=0.5)

client.publish_flow(name="hotel_sentiment", artifacts=[sentiment_table])
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
