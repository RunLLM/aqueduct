import json
from typing import List

import pandas as pd
import pandas_gbq
from google.oauth2 import service_account

from aqueduct_executor.operators.connectors.tabular import config, connector, extract, load


class BigQueryConnector(connector.TabularConnector):
    def __init__(self, config: config.BigQueryConfig):
        self.project_id = config.project_id

        credentials_info = json.loads(config.service_account_credentials)
        self.credentials = service_account.Credentials.from_service_account_info(credentials_info)

    def authenticate(self) -> None:
        # pandas_gbq does not have an explicit authenticate method, so we execute a test query
        pandas_gbq.read_gbq("SELECT 1;", project_id=self.project_id, credentials=self.credentials)

    def discover(self) -> List[str]:
        # TBD eng-708-investigate-how-to-list-tables-for-bigquery
        return []

    def extract(self, params: extract.RelationalParams) -> pd.DataFrame:
        df = pandas_gbq.read_gbq(
            params.query, project_id=self.project_id, credentials=self.credentials
        )
        return df

    def load(self, params: load.RelationalParams, df: pd.DataFrame) -> None:
        pandas_gbq.to_gbq(
            df,
            params.table,
            project_id=self.project_id,
            if_exists=params.update_mode.value,
            credentials=self.credentials,
        )
