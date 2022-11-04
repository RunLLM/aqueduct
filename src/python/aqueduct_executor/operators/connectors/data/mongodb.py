from typing import Any, List, Optional

import pandas as pd
from aqueduct_executor.operators.connectors.data import common, connector, extract, load
from aqueduct_executor.operators.connectors.data.config import MongoDBConfig
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.utils import delete_object
from pymongo import MongoClient
from pymongo.client_session import ClientSession
from pymongo.database import Database


class MongoDBConnector(connector.DataConnector):
    _client: MongoClient[Any]
    _db_name: str

    def __init__(self, config: MongoDBConfig):
        self._client = MongoClient(
            config.auth_uri,
            connect=True,
        )
        self._db_name = config.database
        self._test()

    def _test(self) -> None:
        try:
            self._client.test
            assert (
                self._db_name in self._client.list_database_names()
            ), f"Database {self._db_name} does not exist."
        except Exception as e:
            raise ConnectionError("Unable to connect") from e

    def __del__(self) -> None:
        self._client.close()

    def _connect_db(self, session: Optional[ClientSession] = None) -> Database[Any]:
        if session:
            return session.client[self._db_name]
        return self._client[self._db_name]

    def authenticate(self) -> None:
        self._test()

    def _discover(self, session: Optional[ClientSession] = None) -> List[str]:
        return self._connect_db(session).list_collection_names()

    def discover(self) -> List[str]:
        return self._discover()

    def extract(self, params: extract.MongoDBParams) -> Any:
        assert params.usable()
        query = params.query
        assert query is not None

        db = self._connect_db()
        collection = params.collection
        assert (
            collection in db.list_collection_names()
        ), f"Collection `{collection}` does not exist."

        raw_results = db[collection].find(*(query.args or []), **(query.kwargs or {}))
        return pd.DataFrame(raw_results)

    def load(
        self, params: load.RelationalParams, df: pd.DataFrame, artifact_type: ArtifactType
    ) -> None:
        if artifact_type != ArtifactType.TABLE:
            raise Exception("The data being loaded must be of type table, found %s" % artifact_type)

        with self._client.start_session() as session:
            with session.start_transaction():
                db = self._connect_db(session)
                collections = self._discover(session)
                exists = params.table in collections
                collection = None
                replace = False
                if exists:
                    if params.update_mode == common.UpdateMode.FAIL:
                        raise Exception(f"Specified collection {params.table} already exists.")

                    collection = db[params.table]
                    if params.update_mode == common.UpdateMode.REPLACE:
                        replace = True
                else:
                    collection = db.create_collection(params.table)

                if replace:
                    collection.delete_many({})
                collection.insert_many(df.to_dict("records"))

    def delete(self, tables: List[str]) -> List[SavedObjectDelete]:
        results = []
        db = self._connect_db()

        # This helper simply bypass type check of `delete_object`,
        # which requires the callback to return None.
        def _delete_table(t: str) -> None:
            db.drop_collection(t)

        for table in tables:
            results.append(delete_object(table, lambda t: _delete_table(t)))
        return results
