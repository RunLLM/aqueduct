

<div align="center">
  <a href="https://aqueducthq.com">
    <img src="https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct-logo-two-tone/1x/aqueduct-logo-two-tone-1x.png" width="40%" />
  </a>
  
  <h2 style="border: 0px white;">Run LLMs and ML on any cloud infrastructure</h2>

### 📢 [Slack](https://slack.aqueducthq.com)&nbsp;&nbsp;|&nbsp;&nbsp;🗺️ [Roadmap](https://roadmap.aqueducthq.com)&nbsp;&nbsp;|&nbsp;&nbsp;🐞 [Report a bug](https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D)&nbsp;&nbsp;|&nbsp;&nbsp;✍️ [Blog](https://blog.aqueducthq.com)

  
[![Start Sandbox](https://img.shields.io/static/v1?label=%20&logo=github&message=Start%20Sandbox&color=black)](https://github.com/codespaces/new?hide_repo_select=true&ref=main&repo=496844646)
[![Downloads](https://pepy.tech/badge/aqueduct-ml/month)](https://pypi.org/project/aqueduct-ml/)
[![Slack](https://img.shields.io/static/v1.svg?label=chat&message=on%20slack&color=27b1ff&style=flat)](https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A)
[![GitHub license](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/aqueducthq/aqueduct/blob/master/LICENSE)
[![PyPI version](https://badge.fury.io/py/aqueduct-ml.svg)](https://pypi.org/project/aqueduct-ml/)
[![Tests](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/aqueducthq/aqueduct/actions/workflows/integration-tests.yml)
</div>

**Aqueduct is an MLOps framework that allows you to define and deploy machine learning and LLM workloads on any cloud infrastructure. [Check out our quickstart guide! →](https://docs.aqueducthq.com/quickstart-guide)**

<p align="center">
  <img src="https://user-images.githubusercontent.com/867892/230214641-b0aec53b-4988-4581-84ed-134f97ed9276.png" width="80%" />
</p>

Aqueduct is an open-source MLOps framework that allows you to write code in vanilla Python, run that code on any cloud infrastructure you'd like to use, and gain visibility into the execution and performance of your models and predictions. **[See what infrastructure Aqueduct works with. →](https://aqueducthq.com/integrations/)**

Here's how you can get started: 

```bash
pip3 install aqueduct-ml
aqueduct start
```

### How it works

Aqueduct's Python native API allows you to define ML tasks in regular Python code. You can connect Aqueduct to your existing cloud infrastructure ([docs](https://docs.aqueducthq.com/integrations)), and Aqueduct will seamlessly move your code from your laptop to the cloud or between different cloud infrastructure layers. 

<!--- TODO(vikram): Modify this once we add support for switching into/out of Databricks in a single workflow. --->
For example, we can define a pipeline that trains a model on Kubernetes using a GPU and validates that model in AWS Lambda in a few lines of Python: 

```python
# Use an existing LLM.
vicuna = aq.llm_op('vicuna_7b', engine='eks-us-east-2')
features = vicuna(
    raw_logs,
    { 
        prompt: 
        "Turn this log entry into a CSV: {text}" 
    }
)

# Or write a custom op on your favorite infrastructure!
@op(
  engine='kubernetes',
  # Get a GPU.
  resources={'gpu_resource_name': 'nvidia.com/gpu'}
)
def train(featurized_logs):
  return model.train(features) # Train your model.

train(features)
```

Once you publish this workflow to Aqueduct, you can see it on the UI: 

![image](https://github.com/aqueducthq/aqueduct/assets/867892/d0561772-8799-4046-92ae-3c975d70e47d)

To see how to build your first workflow, check out our **[quickstart guide! →](https://docs.aqueducthq.com/quickstart-guide)**

## Why Aqueduct?

MLOps has become a [tangled mess of siloed infrastructure](https://aqueducthq.com/post/the-mlops-knot/). Most teams need to set up and operate many different cloud infrastructure tools to run ML effectively, but these tools have disparate APIs and interoperate poorly.

Aqueduct provides a single interface to running machine learning tasks on your existing cloud infrastructure — Kubernetes, Spark, Lambda, etc. From the same Python API, you can run code across any or all of these systems seamlessly and gain visibility into how your code is performing.

* **Python-native pipeline API**: Aqueduct’s API allows you define your workflows in vanilla Python, so you can get code into production quickly and effectively. No more DSLs or YAML configs to worry about.
* **Integrated with your infrastructure**: Workflows defined in Aqueduct can run on any cloud infrastructure you use, like Kubernetes, Spark, Airflow, or AWS Lambda. You can get all the benefits of Aqueduct without having to rip-and-replace your existing tooling.
* **Centralized visibility into code, data, & metadata**: Once your workflows are in production, you need to know what’s running, whether it’s working, and when it breaks. Aqueduct gives you visibility into what code, data, metrics, and metadata are generated by each workflow run, so you can have confidence that your pipelines work as expected — and know immediately when they don’t.
* **Runs securely in your cloud**: Aqueduct is fully open-source and runs in any Unix environment. It runs entirely in your cloud and on your infrastructure, so you can be confident that your data and code are secure.

## Overview & Examples

The core abstraction in Aqueduct is a [Workflow](https://docs.aqueducthq.com/workflows), which is a sequence of [Artifacts](https://docs.aqueducthq.com/artifacts) (data) that are transformed by [Operators](https://docs.aqueducthq.com/operators) (compute). 
The input Artifact(s) for a Workflow is typically loaded from a database, and the output Artifact(s) are typically persisted back to a database. 
Each Workflow can either be run on a fixed schedule or triggered on-demand.

To see Aqueduct in action on some real-world machine learning workflows, check out some of our examples:

* [Churn Ensemble](https://github.com/aqueducthq/aqueduct/blob/main/examples/churn_prediction/Customer%20Churn%20Prediction.ipynb)
* [Sentiment Analysis](https://github.com/aqueducthq/aqueduct/blob/main/examples/sentiment-analysis/Sentiment%20Model.ipynb)
* [Impute Missing Wine Data](https://github.com/aqueducthq/aqueduct/blob/main/examples/wine-ratings-prediction/Predict%20Missing%20Wine%20Ratings.ipynb)
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

If you're interested in contributing, please check out our [roadmap](https://roadmap.aqueducthq.com) and join the development channel in [our community Slack](https://slack.aqueducthq.com).
