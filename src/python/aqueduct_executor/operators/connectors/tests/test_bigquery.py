import pytest
from aqueduct_executor.operators.connectors.data import bigquery as bq
from aqueduct_executor.operators.connectors.data.extract import RelationalParams as ExtractParam
from aqueduct_executor.operators.connectors.data.load import RelationalParams as LoadParam
from aqueduct_executor.operators.connectors.tests import conf, utils
from google.cloud import bigquery

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
        cls.client.delete_dataset(_DATASET, delete_contents=True)

    def test_authenticate(self):
        utils.authenticate_test(self.conn)

    @pytest.mark.dependency()
    def test_load(self):
        params = LoadParam(table=_TABLE)
        utils.load_test(self.conn, params, self.test_df)

    @pytest.mark.dependency(depends=["TestBigQuery::test_load"])
    def test_extract(self):
        params = ExtractParam(query="SELECT * FROM {};".format(_TABLE))
        utils.extract_test(self.conn, params, expected_df=self.test_df)

    def test_discover(self):
        tables = set()
        for i in range(2):
            table = f"{_DATASET}.table{i}"
            params = LoadParam(table=table)
            utils.load_test(self.conn, params, self.test_df)
            tables.add(table)
            discover = self.conn.discover()
            assert len(tables.difference(discover)) == 0
