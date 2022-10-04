import pytest
from aqueduct_executor.operators.connectors.data import dataframe, sqlite
from aqueduct_executor.operators.connectors.tests import conf, utils

_TABLE = "test_sqlite"


@pytest.mark.skipif(conf.SKIP_SQLITE, reason="Skip SQLite Flag Set")
class TestSqlite:
    @classmethod
    def setup_class(cls):
        # Setup connector
        config = conf.SQLITE_CONF
        conn = sqlite.SqliteConnector(config)
        cls.conn = conn

        # Setup test dataframe
        cls.test_df = utils.sample_df()

    @classmethod
    def teardown_class(cls):
        cls.conn.engine.connect().execute("DROP TABLE IF EXISTS {};".format(_TABLE))

    def test_authenticate(self):
        utils.authenticate_test(self.conn)

    @pytest.mark.dependency()
    def test_load(self):
        params = {dataframe.LOAD_PARAMS_TABLE_KEY: _TABLE}
        utils.load_test(self.conn, params, self.test_df)

    @pytest.mark.dependency(depends=["TestSqlite::test_load"])
    def test_extract(self):
        params = {dataframe.EXTRACT_PARAMS_QUERY_KEY: "SELECT * FROM {};".format(_TABLE)}
        utils.extract_test(self.conn, params, expected_df=self.test_df)
