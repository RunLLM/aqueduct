from typing import Any, Dict, List, Optional

import awswrangler as wr
import pandas as pd
from aqueduct_executor.operators.connectors.data import connector, extract, load
from aqueduct_executor.operators.connectors.data.config import AthenaConfig
from aqueduct_executor.operators.connectors.data.utils import construct_boto_session
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete

DEFAULT_CATALOG = "AwsDataCatalog"
LIST_TABLES_QUERY_ATHENA = "AQUEDUCT_ATHENA_LIST_TABLE"


class AthenaConnector(connector.DataConnector):
    def __init__(self, config: AthenaConfig):
        self.session = construct_boto_session(config)
        self.output_location = config.output_location
        self.database = config.database

    def _list_tables(self) -> List[str]:
        client = self.session.client("athena")
        tables = client.list_table_metadata(CatalogName=DEFAULT_CATALOG, DatabaseName=self.database)
        return [table["Name"] for table in tables["TableMetadataList"]]

    def authenticate(self) -> None:
        self._list_tables()
        # This checks whether the S3 output path is valid.
        wr.athena.read_sql_query(
            sql="SELECT 1;",
            database=self.database,
            boto3_session=self.session,
            ctas_approach=False,
            s3_output=self.output_location,
            keep_files=False,
        )

    def discover(self) -> List[str]:
        return self._list_tables()

    def extract(self, params: extract.RelationalParams) -> pd.DataFrame:
        assert params.usable(), "Query is not usable. Did you forget to expand placeholders?"
        if params.query == LIST_TABLES_QUERY_ATHENA:
            return pd.DataFrame(self._list_tables(), columns=["Tables"])
        else:
            return wr.athena.read_sql_query(
                sql=params.query,
                database=self.database,
                boto3_session=self.session,
                # Disabling ctas improves generality at the cost of performance.
                # More details here: https://aws-sdk-pandas.readthedocs.io/en/stable/stubs/awswrangler.athena.read_sql_query.html
                ctas_approach=False,
                s3_output=self.output_location,
                keep_files=False,
            )

    def load(
        self, params: load.RelationalParams, df: pd.DataFrame, artifact_type: ArtifactType
    ) -> None:
        raise Exception("Save operation not supported for Athena.")

    def delete(self, objects: List[str]) -> List[SavedObjectDelete]:
        raise Exception("Delete operation not supported for Athena.")
