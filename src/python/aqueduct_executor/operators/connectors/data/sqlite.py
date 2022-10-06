import pandas as pd
from aqueduct_executor.operators.connectors.data import config, load, relational
from aqueduct_executor.operators.utils.enums import ArtifactType
from sqlalchemy import create_engine, engine

SQLITE_MAX_VARIABLE_NUMBER = 32766


class SqliteConnector(relational.RelationalConnector):
    def __init__(self, config: config.SqliteConfig):
        conn_engine = _create_engine(config)
        super().__init__(conn_engine)

    def load(
        self, params: load.RelationalParams, df: pd.DataFrame, artifact_type: ArtifactType
    ) -> None:
        if artifact_type != ArtifactType.TABLE:
            raise Exception("The data being loaded must be of type table, found %s" % artifact_type)
        # NOTE (saurav): df._to_sql has known performance issues. Using `method="multi"` helps incrementally,
        # since pandas will pass multiple rows in a single INSERT. If this still remains an issue, we can pass in a
        # callable function for `method` that does bulk loading.
        # See: https://pandas.pydata.org/docs/user_guide/io.html#io-sql-method
        df.to_sql(
            params.table,
            con=self.engine,
            if_exists=params.update_mode.value,
            index=False,
            method="multi",
            chunksize=int(SQLITE_MAX_VARIABLE_NUMBER / len(df.columns)),
        )


def _create_engine(config: config.SqliteConfig) -> engine.Engine:
    # SQLite Dialect:
    # https://docs.sqlalchemy.org/en/14/dialects/sqlite.html#dialect-sqlite-pysqlite-connect
    url = "sqlite:///{database}".format(
        database=config.database,
    )
    return create_engine(url)
