from aqueduct_executor.operators.connectors.data import config, relational, utils
from sqlalchemy import create_engine, engine


class PostgresConnector(relational.RelationalConnector):
    def __init__(self, config: config.PostgresConfig):
        conn_engine = _create_engine(config)
        super().__init__(conn_engine)


def _create_engine(config: config.PostgresConfig) -> engine.Engine:
    # Postgres Dialect:
    # https://docs.sqlalchemy.org/en/14/dialects/postgresql.html#module-sqlalchemy.dialects.postgresql.psycopg2
    if config.password:
        url = "postgresql://{username}:{password}@{host}:{port}/{database}".format(
            username=config.username,
            password=utils.url_encode(config.password),
            host=config.host,
            port=config.port,
            database=config.database,
        )
    else:
        url = "postgresql://{username}@{host}:{port}/{database}".format(
            username=config.username,
            host=config.host,
            port=config.port,
            database=config.database,
        )
    return create_engine(url)
