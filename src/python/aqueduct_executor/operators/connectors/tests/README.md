In `aqueduct/src` run `PYTHONPATH=. pytest python/aqueduct_executor/operators/connectors/tests/`.

This will install the requirements in `requirements.txt`. If you are not using `mysqlclient` or `pyodbc`. comment them out. Otherwise:
* If you are testing `mysql-client`, you might need to install the mysqlserver first: https://github.com/PyMySQL/mysqlclient. This should have instructions for Mac and Linux.
