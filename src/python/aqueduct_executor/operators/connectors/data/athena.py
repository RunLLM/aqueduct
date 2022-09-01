from typing import Any, Dict, List, Optional

import awswrangler as wr
import pandas as pd
from aqueduct_executor.operators.connectors.data import connector, extract, load
from aqueduct_executor.operators.connectors.data.config import AthenaConfig
from aqueduct_executor.operators.connectors.data.utils import construct_boto_session
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete


class AthenaConnector(connector.DataConnector):
    def __init__(self, config: AthenaConfig):
        self.session = construct_boto_session(config)
        self.output_location = config.output_location
        self.database = config.database

    def authenticate(self) -> None:
        wr.athena.read_sql_query(
            sql="SELECT 1;",
            database=self.database,
            boto3_session=self.session,
            ctas_approach=False,
            s3_output=self.output_location,
            keep_files=False,
        )

    def discover(self) -> List[str]:
        raise Exception("Discover is not supported for Athena.")

    def extract(self, params: extract.RelationalParams) -> pd.DataFrame:
        assert params.usable(), "Query is not usable. Did you forget to expand placeholders?"
        return wr.athena.read_sql_query(
            sql=params.query,
            database=self.database,
            boto3_session=self.session,
            ctas_approach=False,
            s3_output=self.output_location,
            keep_files=False,
        )

    def load(
        self, params: load.RelationalParams, df: pd.DataFrame, artifact_type: ArtifactType
    ) -> None:
        raise Exception("Save operation not supported for Athena.")

    def _delete_object(self, name: str, context: Optional[Dict[str, Any]] = None) -> None:
        return

    def delete(self, objects: List[str]) -> List[SavedObjectDelete]:
        raise Exception("Delete operation not supported for Athena.")
