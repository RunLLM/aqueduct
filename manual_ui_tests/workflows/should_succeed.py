import aqueduct as aq
import pandas as pd
import numpy as np
from sklearn.linear_model import LogisticRegression
from sklearn.tree import DecisionTreeClassifier

NAME="should_succeed"
DESCRIPTION="""* Workflows Page: everything should succeed.
* Workflow Details Page: everything should be green.
* Data Page: pred_churn artifact should appears."""


@aq.op
def log_featurize(cust: pd.DataFrame) -> pd.DataFrame:
    features = cust.copy()
    skip_cols = ["cust_id", "using_deep_learning", "using_dbt"]

    for col in features.columns.difference(skip_cols):
        features["log_" + col] = np.log(features[col] + 1.0)

    return features.drop(columns="cust_id")

def deploy(client, integration):
    customers_table = pd.read_csv(
        "https://raw.githubusercontent.com/aqueducthq/aqueduct/main/examples/churn_prediction/data/customers.csv"
    )
    churn_table = pd.read_csv(
        "https://raw.githubusercontent.com/aqueducthq/aqueduct/main/examples/churn_prediction/data/churn_data.csv"
    )
    features_table = log_featurize.local(customers_table)

    linear_model = LogisticRegression(max_iter=10000)
    linear_model.fit(features_table, churn_table["churn"])
    decision_tree_model = DecisionTreeClassifier(max_depth=10, min_samples_split=3)
    decision_tree_model.fit(features_table, churn_table["churn"])

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

    warehouse = client.integration(name=integration)
    customers_table = warehouse.sql(query="SELECT * FROM customers;")
    features_table = log_featurize(customers_table)
    linear_pred_table = predict_linear(features_table)
    tree_pred_table = predict_tree(features_table)
    churn_table = predict_ensemble(customers_table, linear_pred_table, tree_pred_table)
    
    @aq.check(description="Ensuring valid probabilities.")
    def valid_probabilities(df: pd.DataFrame):
        return (df["prob_churn"] >= 0) & (df["prob_churn"] <= 1)
    
    check_result = valid_probabilities(churn_table)
    avg_pred_churn_metric = churn_table.mean("prob_churn")
    avg_pred_churn_metric.bound(lower=0.1)
    avg_pred_churn_metric.bound(upper=0.3)
    avg_pred_churn_metric.bound(upper=0.4, severity="error")
    churn_table.save(warehouse.config("pred_churn", aq.LoadUpdateMode.REPLACE))
    client.publish_flow(
        name=NAME,
        description=DESCRIPTION,
        artifacts=[churn_table, avg_pred_churn_metric],
        schedule=aq.hourly(),
    )