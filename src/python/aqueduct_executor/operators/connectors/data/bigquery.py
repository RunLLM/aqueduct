import json
from typing import Any, Dict, List, Optional

import pandas as pd
from aqueduct_executor.operators.connectors.data import common, config, connector, extract, load
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.utils import delete_object
from google.cloud import bigquery
from google.oauth2 import service_account


class BigQueryConnector(connector.DataConnector):
    def __init__(self, config: config.BigQueryConfig):
        self.project_id = config.project_id

        credentials_info = json.loads(config.service_account_credentials)
        self.credentials = service_account.Credentials.from_service_account_info(credentials_info)
        self.client = bigquery.Client(credentials=self.credentials, project=self.project_id)

    def authenticate(self) -> None:
        # There is no explicit authenticate method, so we execute a test query
        self.client.query("SELECT 1;")

    def discover(self) -> List[str]:
        all_tables = []
        for dataset in self.client.list_datasets():
            tables = self.client.list_tables(dataset.dataset_id)
            all_tables.extend([table.full_table_id.split(":")[-1] for table in tables])
        return all_tables

    def extract(self, params: extract.RelationalParams) -> Any:
        assert params.usable(), "Query is not usable. Did you forget to expand placeholders?"
        query = self.client.query(params.query)
        df = query.result().to_dataframe()
        return df

    def _delete_object(self, name: str, context: Optional[Dict[str, Any]] = None) -> None:
        self.client.delete_table(name, not_found_ok=False)

    def delete(self, tables: List[str]) -> List[SavedObjectDelete]:
        results = []
        for table in tables:
            results.append(delete_object(table, self._delete_object))
        return results

    def load(self, params: load.RelationalParams, df: Any, artifact_type: ArtifactType) -> None:
        if artifact_type != ArtifactType.TABLE:
            raise Exception("The data being loaded must be of type table, found %s" % artifact_type)

        update_mode = params.update_mode
        write_disposition = bigquery.WriteDisposition.WRITE_TRUNCATE  # Default
        if update_mode == common.UpdateMode.APPEND:
            write_disposition = bigquery.WriteDisposition.WRITE_APPEND
        if update_mode == common.UpdateMode.REPLACE:
            write_disposition = bigquery.WriteDisposition.WRITE_TRUNCATE
        if update_mode == common.UpdateMode.FAIL:
            write_disposition = bigquery.WriteDisposition.WRITE_EMPTY
        # Since string columns use the "object" dtype, pass in a (partial) schema
        # to ensure the correct BigQuery data type.
        partial_schema = []
        for column in df:
            if df[column].dtype == object:
                partial_schema.append(bigquery.SchemaField(column, "STRING"))

        job_config = bigquery.LoadJobConfig(
            schema=partial_schema, write_disposition=write_disposition
        )
        job = self.client.load_table_from_dataframe(df, params.table, job_config=job_config)

        # Wait for the load job to complete.
        job.result()
