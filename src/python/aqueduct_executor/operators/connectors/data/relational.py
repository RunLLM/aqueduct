from typing import Any, Callable, Dict, List, Optional

import pandas as pd
from aqueduct_executor.operators.connectors.data import connector, extract, load
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.utils import delete_object
from sqlalchemy import MetaData, engine, inspect
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.declarative import declarative_base


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
        )
