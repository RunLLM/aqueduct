from aqueduct_executor.operators.connectors.data import config, relational, utils
from sqlalchemy import create_engine, engine


class MySqlConnector(relational.RelationalConnector):
    def __init__(self, config: config.MySqlConfig):
        conn_engine = _create_engine(config)
        super().__init__(conn_engine)


def _create_engine(config: config.MySqlConfig) -> engine.Engine:
    # MySQL Dialect:
    # https://docs.sqlalchemy.org/en/14/dialects/mysql.html#module-sqlalchemy.dialects.mysql.mysqldb
    url = "mysql+mysqldb://{username}:{password}@{host}:{port}/{database}".format(
        username=config.username,
        password=utils.url_encode(config.password),
        host=config.host,
        port=config.port,
        database=config.database,
    )
    return create_engine(url)
