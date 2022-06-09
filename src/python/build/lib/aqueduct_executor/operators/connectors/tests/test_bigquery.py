import pytest
from google.cloud import bigquery

from aqueduct_executor.operators.connectors.tabular import bigquery as bq
from aqueduct_executor.operators.connectors.tabular import dataframe

from aqueduct_executor.operators.connectors.tests import conf
from aqueduct_executor.operators.connectors.tests import utils

_DATASET = "testdata"
_TABLE = "testdata.bigquery"


@pytest.mark.skipif(conf.SKIP_BIGQUERY, reason="Skip BigQuery Flag Set")
class TestBigQuery:
    @classmethod
    def setup_class(cls):
        # Setup connector
        config = conf.BIGQUERY_CONF
        conn = bq.BigQueryConnector(config)
        cls.conn = conn

        # Setup test dataframe
        cls.test_df = utils.sample_df()

        # Create BigQuery client to perform setup and teardown
        cls.client = bigquery.Client(project=conn.project_id, credentials=conn.credentials)
        cls.client.create_dataset(_DATASET)

    @classmethod
    def teardown_class(cls):
        cls.client.delete_table(_TABLE)
        cls.client.delete_dataset(_DATASET)

    def test_authenticate(self):
        utils.authenticate_test(self.conn)

    @pytest.mark.dependency()
    def test_load(self):
        params = {dataframe.LOAD_PARAMS_TABLE_KEY: _TABLE}
        utils.load_test(self.conn, params, self.test_df)

    @pytest.mark.dependency(depends=["TestBigQuery::test_load"])
    def test_extract(self):
        params = {dataframe.EXTRACT_PARAMS_QUERY_KEY: "SELECT * FROM {};".format(_TABLE)}
        utils.extract_test(self.conn, params, expected_df=self.test_df)
