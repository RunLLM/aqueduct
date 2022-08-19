import pytest
from aqueduct_executor.operators.connectors.data import dataframe, mysql
from aqueduct_executor.operators.connectors.tests import conf, utils

_TABLE = "test_mysql"


@pytest.mark.skipif(conf.SKIP_MYSQL, reason="Skip MySQL Flag Set")
class TestMySql:
    @classmethod
    def setup_class(cls):
        # Setup connector
        config = conf.MYSQL_CONF
        conn = mysql.MySqlConnector(config)
        cls.conn = conn

        # Setup test dataframe
        cls.test_df = utils.sample_df()

    @classmethod
    def teardown_class(cls):
        cls.conn.engine.connect().execute("DROP TABLE IF EXISTS {} CASCADE;".format(_TABLE))

    def test_authenticate(self):
        utils.authenticate_test(self.conn)

    @pytest.mark.dependency()
    def test_load(self):
        params = {dataframe.LOAD_PARAMS_TABLE_KEY: _TABLE}
        utils.load_test(self.conn, params, self.test_df)

    @pytest.mark.dependency(depends=["TestMySql::test_load"])
    def test_extract(self):
        params = {dataframe.EXTRACT_PARAMS_QUERY_KEY: "SELECT * FROM {};".format(_TABLE)}
        utils.extract_test(self.conn, params, expected_df=self.test_df)
