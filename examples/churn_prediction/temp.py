print("Cell 0")
import pandas as pd
import numpy as np
import aqueduct as aq

# Read some customer data from the Aqueduct repo.
customers_table = pd.read_csv(
    "https://raw.githubusercontent.com/aqueducthq/aqueduct/main/examples/churn_prediction/data/customers.csv"
)
churn_table = pd.read_csv(
    "https://raw.githubusercontent.com/aqueducthq/aqueduct/main/examples/churn_prediction/data/churn_data.csv"
)
pd.merge(customers_table, churn_table, on="cust_id").head()


print("Cell 1")
# The @op decorator here allows Aqueduct to run this function as
# a part of an Aqueduct workflow. It tells Aqueduct that when
# we execute this function, we're defining a step in the workflow.
# While the results can be retrieved immediately, nothing is
# published until we call `publish_flow()` below.
@aq.op
def log_featurize(cust: pd.DataFrame) -> pd.DataFrame:
    """
    log_featurize takes in customer data from the Aqueduct customers table
    and log normalizes the numerical columns using the numpy.log function.
    It skips the cust_id, using_deep_learning, and using_dbt columns because
    these are not numerical columns that require regularization.

    log_featurize adds all the log-normalized values into new columns, and
    maintains the original values as-is. In addition to the original company_size
    column, log_featurize will add a log_company_size column.
    """
    features = cust.copy()
    skip_cols = ["cust_id", "using_deep_learning", "using_dbt"]

    for col in features.columns.difference(skip_cols):
        features["log_" + col] = np.log(features[col] + 1.0)

    return features.drop(columns="cust_id")


print("Cell 2")
# Calling `.local()` on an @op-annotated function allows us to execute the
# function locally for testing purposes. When a function is called with
# `.local()`, Aqueduct does not capture the function execution as a part of
# the definition of a workflow.
features_table = log_featurize.local(customers_table)
features_table.head()


print("Cell 3")
from sklearn.linear_model import LogisticRegression

linear_model = LogisticRegression(max_iter=10000)
linear_model.fit(features_table, churn_table["churn"])


print("Cell 4")
from sklearn.tree import DecisionTreeClassifier

decision_tree_model = DecisionTreeClassifier(max_depth=10, min_samples_split=3)
decision_tree_model.fit(features_table, churn_table["churn"])


print("Cell 5")
@aq.op
def predict_linear(features_table):
    """
    Generates predictions using the logistic regression model and
    returns a new DataFrame with a column called linear that has
    the likelihood of the customer churning.
    """
    return pd.DataFrame({"linear": linear_model.predict_proba(features_table)[:, 1]})

@aq.op
def predict_tree(features_table):
    """
    Generates predictions using the decision tree model and
    returns a new DataFrame with a column called tree that has
    the likelihood of the customer churning.
    """
    return pd.DataFrame({"tree": decision_tree_model.predict_proba(features_table)[:, 1]})

@aq.op
def predict_ensemble(customers_table, linear_pred_table, tree_pred_table):
    """
    predict_ensemble combines the results from our logistic regression
    and decision tree models by taking the average of the two models'
    probabilities that a user might churn. The resulting average is
    then assigned into the `prob_churn` column on the customers_table.
    """
    return customers_table.assign(prob_churn=linear_pred_table.join(tree_pred_table).mean(axis=1))


print("Cell 6")
features_table = log_featurize.local(customers_table)
linear_pred_table = predict_linear.local(features_table)
tree_pred_table = predict_tree.local(features_table)
churn_table = predict_ensemble.local(customers_table, linear_pred_table, tree_pred_table)


print("Cell 7")
churn_table.head()


print("Cell 8")
# If you're running your notebook on a separate machine from your
# Aqueduct server, change this to the address of your Aqueduct server.
address = "localhost:8080"

# If you're running youre notebook on a separate machine from your
# Aqueduct server, you will have to copy your API key here rather than
# using `get_apikey()`.
api_key = "09LOAH7CW3MDUVGQF5JP62BRK1ZX8INS"
client = aq.Client(api_key, address)


print("Cell 9")
warehouse = client.integration(name="aqueduct_demo")

# customers_table is an Aqueduct TableArtifact, which is a wrapper around
# a Pandas DataFrame. A TableArtifact can be used as argument to any operator
# in a workflow; you can also call .get() on a TableArtifact to retrieve
# the underlying DataFrame and interact with it directly.
customers_table = warehouse.sql(query="SELECT * FROM customers;")
print(type(customers_table))


print("Cell 10")
# This gets the head of the underlying DataFrame. Note that you can't
# pass a DataFrame as an argument to a workflow; you must use the Aqueduct
# TableArtifact!
customers_table.get().head()


print("Cell 11")
features_table = log_featurize(customers_table)
print(type(features_table))


print("Cell 12")
features_table.get().head()


print("Cell 13")
linear_pred_table = predict_linear(features_table)
tree_pred_table = predict_tree(features_table)
churn_table = predict_ensemble(customers_table, linear_pred_table, tree_pred_table)


print("Cell 14")
churn_table.get().head()


print("Cell 15")
@aq.check(description="Ensuring valid probabilities.")
def valid_probabilities(df: pd.DataFrame):
    return (df["prob_churn"] >= 0) & (df["prob_churn"] <= 1)


print("Cell 16")
check_result = valid_probabilities(churn_table)


print("Cell 17")
# Use Aqueduct's built-in mean metric to calculate the average value of `prob_churn`.
# Calling .get() on the metric will retrieve the current value.
avg_pred_churn_metric = churn_table.mean("prob_churn")
avg_pred_churn_metric.get()


print("Cell 18")
# Bounds on metrics ensure that the metric stays within a valid range.
# In this case, we'd ideally like churn to be between .1 and .3, and we
# know something's gone wrong if it's above .4.
avg_pred_churn_metric.bound(lower=0.1)
avg_pred_churn_metric.bound(upper=0.3)
avg_pred_churn_metric.bound(upper=0.4, severity="error")


print("Cell 19")
# This tells Aqueduct to save the results in churn_table
# back to the demo DB we configured earlier.
# NOTE: At this point, no data is actually saved! This is just
# part of a workflow spec that will be executed once the workflow
# is published in the next cell.
warehouse.save(churn_table, table_name="pred_churn", update_mode="replace")


print("Cell 20")
# This publishes all of the logic needed to create churn_table
# and avg_pred_churn_metric to Aqueduct. The URL below will
# take you to the Aqueduct UI, which will show you the status
# of your workflow runs and allow you to inspect them.
churn_flow = client.publish_flow(
    name="Demo Churn Ensemble",
    artifacts=[churn_table, avg_pred_churn_metric],
    # Uncomment the following line to schedule on a hourly basis.
    # schedule=aq.hourly(),
)
print(churn_flow.id())