"""
CONFIG FILE FOR TABULAR CONNECTOR INTEGRATION TESTS
- To skip a particular test set the relevant `SKIP_` flag to False
- Set the `_CONF` dict for all connectors being tested.
- Commented out config fields are optional.
"""
from aqueduct_executor.operators.connectors.tabular.config import BigQueryConfig
from aqueduct_executor.operators.connectors.tabular import postgres
from aqueduct_executor.operators.connectors.tabular import snowflake
from aqueduct_executor.operators.connectors.tabular import mysql
from aqueduct_executor.operators.connectors.tabular import sql_server
from aqueduct_executor.operators.connectors.tabular import bigquery
from aqueduct_executor.operators.connectors.tabular import sqlite

"""FLAGS TO SKIP TESTS"""
SKIP_POSTGRES = True
SKIP_SNOWFLAKE = True
SKIP_MYSQL = True
SKIP_REDSHIFT = True
SKIP_MARIADB = True
SKIP_SQL_SERVER = True
SKIP_BIGQUERY = False
SKIP_SQLITE = True

# """POSTGRES CONFIG"""
# POSTGRES_CONF = {
#     postgres._CONFIG_USERNAME_KEY: "",
#     postgres._CONFIG_PASSWORD_KEY: "",
#     postgres._CONFIG_DATABASE_KEY: "",
#     postgres._CONFIG_HOST_KEY: "",
#     # postgres._CONFIG_PORT_KEY: "",
# }

# """SNOWFLAKE CONFIG"""
# SNOWFLAKE_CONF = {
#     snowflake._CONFIG_USERNAME_KEY: "",
#     snowflake._CONFIG_PASSWORD_KEY: "",
#     snowflake._CONFIG_ACCOUNT_IDENTIFIER_KEY: "",
#     snowflake._CONFIG_DATABASE_KEY: "",
#     snowflake._CONFIG_WAREHOUSE_KEY: "",
#     # snowflake._CONFIG_SCHEMA_KEY: "",
# }

# """MYSQL CONFIG"""
# MYSQL_CONF = {
#     mysql._CONFIG_USERNAME_KEY: "",
#     mysql._CONFIG_PASSWORD_KEY: "",
#     mysql._CONFIG_DATABASE_KEY: "",
#     mysql._CONFIG_HOST_KEY: "",
#     mysql._CONFIG_PORT_KEY: "",
# }

# """REDSHIFT CONFIG"""
# REDSHIFT_CONF = {
#     postgres._CONFIG_USERNAME_KEY: "",
#     postgres._CONFIG_PASSWORD_KEY: "",
#     postgres._CONFIG_DATABASE_KEY: "",
#     postgres._CONFIG_HOST_KEY: "",
#     postgres._CONFIG_PORT_KEY: "",
# }

# """MARIADB CONFIG"""
# MARIADB_CONF = {
#     mysql._CONFIG_USERNAME_KEY: "",
#     mysql._CONFIG_PASSWORD_KEY: "",
#     mysql._CONFIG_DATABASE_KEY: "",
#     mysql._CONFIG_HOST_KEY: "",
#     mysql._CONFIG_PORT_KEY: "",
# }

# """SQL SERVER CONFIG"""
# SQL_SERVER_CONF = {
#     sql_server._CONFIG_USERNAME_KEY: "",
#     sql_server._CONFIG_PASSWORD_KEY: "",
#     sql_server._CONFIG_DATABASE_KEY: "",
#     sql_server._CONFIG_HOST_KEY: "",
#     sql_server._CONFIG_PORT_KEY: "",
# }

"""BIGQUERY CONFIG"""
BIGQUERY_CONF = BigQueryConfig(project_id = "aqueduct-connector-test", service_account_credentials = r"""
{
  "type": "service_account",
  "project_id": "aqueduct-connector-test",
  "private_key_id": "8743d365f5b0cef32c31cd5aa0f80c82681ab34a",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCqeF+qUtzOqMIK\nDfo9PVKgTjSzGHmwLplVnQ+sYE9ceP5IOL28DT+bwNkg9OImFZblHMgE7EatF1A7\nYuZGrNooZePVDr1EMlZtvbnTSqNAmOO7vyQ1FWObWnYF73GhlxrhVjyCeENHo0OY\nUcUWIcOghtapKcdD0e2CNKJmz29C1sS/mQnYFdGEmimWMBwgZCFKXlF+AF/9t5/V\n9F+aiiAN60d0vEbbuodZoz5mtMFjnLR7faSt0zI4PkWBdi/luJm3CZ+D7sLR5Xnk\nJj7gn/ralYZt2e6NNvxrcbjv5A0vTNI4arxjyzrJphIbMyHEpwCt54ZMVybSFEUI\nKAowVq4bAgMBAAECggEAAPJUwGaBqEfOg9KRXPrMU4dml+I2rD99/tnyV7qR60zG\nEYwgKqW20V4LEZEtE8h7snivhaevzlzwR0z8NmuSndD4Sk23hIxm55FB9NOXxqOb\nH+v0MX2whTKdRF+/08I56mFqcnHd1SUIE+AB1whimMprGg1Shd33evUCp9q03rIB\nZZn1PT/mWdXov05hWYHTHKlWGrC1ApA2ji8N/rZ08BQlxA1ip1XITxEkz1sxFxUQ\nNL4mGy5oAaPttdirxJu2/4qO/HF+xJkedogU0jdiRjBFpzCwTFQaGuCBjhraP5nd\nHzSNfzfnm/91fE2izcWU1NeNuY9+daXkYbkQru4B6QKBgQDWXx1aYBlYQgtcSS1m\n3YKDQ/abOD7XvCMwMHyrZvzA4hZfTOv0ydKo14t7e9wRNDbUi9unOZjFx3C4opNH\n+I6mrLqhFui9GpdxGWnlNAUsjOYo/fCAVL1sdeqsJkpzm24AJd95abb0Cq9X9MtL\n6sOIt3t3VWsUUeAnSY4hJSNrUwKBgQDLktOLJowlstpKWETCe94gmUI+BPx7KciE\n10LRx8E6gugjkUfDGJ1sp8K7CAUUjHEFCmRShjaD4X49VeiG8z9w1UTw8qxsr+yQ\nSP4BFuKLvVDYDIpmJ+cpfhEQ/FTzPpmdW5RzBymE/uSBMJgc6KrvV/x6SjN1T0Bn\n1yqu5tahGQKBgQCYZVm6q+KYqarl2mfaXtKveptP0XZra6YgVffq6fX5MUDyUv7T\nML7/pOvVx0G1QUdRZnOqt/lxcM0jlP/bBEp1Fwo+BslB1iufDZAIjyi2eRwOPCjD\nMnrPJizEYRxAf1h95m6uI4caipYIk1ALEkQbZ0TwmtrawTH2/AV8bqh1XQKBgQCZ\nWhTDmRkv+OhZ4t6BR0BQfEMbZzQvL42fDG2IjBqykhR/XpyZijxksoeNzv/Mt/MX\noflq9TGx7Tbky4dryWf7/px9icF76pahJms5tNyZ+dYhumizhdGsPwxqKDtyNbEQ\nigFtGXMcfcryywF7nYXO4RAPqz/SWg4ha0P7F2eNWQKBgQC1SaLivYdtt/qSWtl0\nemMWHVoLMxbFmuCQT9HB0TOMoUWSPgGRmLfS5utjFL08QpSPKowxt5iUbpXdfNA5\nFErZ/lOBapduTbqGdBnTIQMAQ+8HKjbz5MrvDGeDC5UrbcQaN/fR/9RgQCDB8dn3\nUjCRBRMmNKmbI1e2OiBJMFFJ/g==\n-----END PRIVATE KEY-----\n",
  "client_email": "connection-test@aqueduct-connector-test.iam.gserviceaccount.com",
  "client_id": "116285665412906810374",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/connection-test%40aqueduct-connector-test.iam.gserviceaccount.com"
}
""")

# """SQLITE CONFIG"""
# SQLITE_CONF = {
#     sqlite._CONFIG_DATABASE_PATH_KEY: "",
# }
