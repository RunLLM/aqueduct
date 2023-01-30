from os import cpu_count

import pytest
from aqueduct.constants.enums import ExecutionStatus, ServiceType
from aqueduct.error import AqueductError, InvalidUserArgumentException

from aqueduct import global_config, op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test
from .extract import extract
from .save import save


@pytest.mark.enable_only_for_engine_type(ServiceType.DATABRICKS)
def test_spark_function(client, flow_name, data_integration, engine):
    """Test against PySpark code on Spark-based compute engine."""
    global_config({"engine": engine, "lazy": True})

    @op
    def _log_featurize_spark(cust):
        import numpy as np
        import pyspark.sql.functions as F
        from pyspark.sql.types import FloatType

        """
        log_featurize takes in customer data from the Aqueduct customers table
        and log normalizes the numerical columns using the numpy.log function.
        It skips the cust_id, using_deep_learning, and using_dbt columns because
        these are not numerical columns that require regularization.

        log_featurize adds all the log-normalized values into new columns, and
        maintains the original values as-is. In addition to the original company_size
        column, log_featurize will add a log_company_size column.
        """

        def udf_np_log(a):
            # actual function
            return float(np.log(a + 1.0))

        np_log = F.udf(udf_np_log, FloatType())
        features = cust.alias("features")
        skip_cols = ["cust_id", "using_deep_learning", "using_dbt"]

        for col in [c for c in features.schema.names if c not in skip_cols]:
            features = features.withColumn("LOG_" + col, np_log(features[col]))

        return features.drop("cust_id")

    table_artifact = extract(data_integration, DataObject.CUSTOMERS)
    output_artifact = _log_featurize_spark(table_artifact)
    save(data_integration, output_artifact)

    default_cpus_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[output_artifact],
        engine=engine,
    )
