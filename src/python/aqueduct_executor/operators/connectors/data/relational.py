from typing import Any, Callable, Dict, List, Optional

import pandas as pd
from aqueduct_executor.operators.connectors.data import connector, extract, load
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.utils import delete_object
from sqlalchemy import MetaData, engine, inspect
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.types import VARCHAR


class RelationalConnector(connector.DataConnector):
    def __init__(self, conn_engine: engine.Engine):
        self.engine = conn_engine

    def __del__(self) -> None:
        self.engine.dispose()

    def authenticate(self) -> None:
        try:
            self.engine.connect()
        except SQLAlchemyError as e:
            raise ConnectionError("Unable to connect.") from e

    def discover(self) -> List[str]:
        return inspect(self.engine).get_table_names()  # type: ignore

    def extract(self, params: extract.RelationalParams) -> Any:
        assert params.usable(), "Query is not usable. Did you forget to expand placeholders?"
        return pd.read_sql(params.query, con=self.engine)

    def _delete_object(self, name: str, context: Optional[Dict[str, Any]] = None) -> None:
        if context:
            metadata = context["metadata"]
            base = context["base"]
        else:
            raise Exception("Unexpectedly cannot find context for deletion.")
        sql_table = metadata.tables[name]
        base.metadata.drop_all(self.engine, [sql_table], checkfirst=True)

    def delete(self, tables: List[str]) -> List[SavedObjectDelete]:
        results = []
        base = declarative_base()
        metadata = MetaData()
        metadata.reflect(bind=self.engine)
        delete_helper: Callable[[str], None] = lambda name: self._delete_object(
            name, context={"metadata": metadata, "base": base}
        )
        for table in tables:
            results.append(
                delete_object(
                    table,
                    delete_helper,
                )
            )
        return results

    def _map_object_dtype_to_varchar(self, df: pd.DataFrame) -> Dict[str, VARCHAR]:
        """Given a DataFrame, for each string column (i.e. object dtype), it
        returns a mapping of column name to the appropriate SQL type VARCHAR(N),
        where N is large enough for all values in the column. Non-string
        columns will not be included in the returned map.
        """
        col_to_type = {}
        for col in df.select_dtypes(include=["object"]):
            max_size = df[col].astype(str).str.len().max()
            # Use powers of 2 to determine how large the column needs to be
            # in terms of number of characters
            if max_size < 256:
                col_to_type[col] = VARCHAR(256)
            elif max_size < 512:
                col_to_type[col] = VARCHAR(512)
            elif max_size < 1024:
                col_to_type[col] = VARCHAR(1024)
            elif max_size < 4096:
                col_to_type[col] = VARCHAR(4096)
            elif max_size < 16384:
                col_to_type[col] = VARCHAR(16384)
            elif max_size < 32768:
                col_to_type[col] = VARCHAR(32768)
            elif max_size < 65536:
                # The max VARCHAR size supported by many RDBMS is 65535 (2^16 - 1)
                col_to_type[col] = VARCHAR(65535)
            else:
                raise Exception(
                    "Cannot support saving string columns with length greater than 65535 bytes"
                )
        return col_to_type

    def load(
        self, params: load.RelationalParams, df: pd.DataFrame, artifact_type: ArtifactType
    ) -> None:
        if artifact_type != ArtifactType.TABLE:
            raise Exception("The data being loaded must be of type table, found %s" % artifact_type)

        # Map of only string columns to their SQL type
        # If a column is not in this map, we rely on Panda's default mapping
        col_to_type = self._map_object_dtype_to_varchar(df)

        # NOTE (saurav): df._to_sql has known performance issues. Using `method="multi"` helps incrementally,
        # since pandas will pass multiple rows in a single INSERT. If this still remains an issue, we can pass in a
        # callable function for `method` that does bulk loading.
        # See: https://pandas.pydata.org/docs/user_guide/io.html#io-sql-method
        df.to_sql(
            params.table,
            con=self.engine,
            if_exists=params.update_mode.value,
            index=False,
            dtype=col_to_type,
            method="multi",
        )
