# To add a new schema version, please refer to the following checklist
* Add your version package under `versions` directory.
* The new veresion package should pattern-match existing versions. There are two use cases:
    * Migration can be done using SQL queries. The package should contain the following:
        * `up_postgres.go` for postgres query upgrading from previous version to this version.
        * `up_sqlite.go` for sqlite query upgrading from previous version to this version.
        * `down_postgres.go` for postgres query downgrading from this version to previous version.
        * `main.go` to actually run the above queries.
    * Migration requires running a go logic. This typically happens for backfill.
        * `main.go` with `Up()` and `Down()` implementations for the go logic.
* Add your version package to `migrator/register.go` file:
    * Add the package to import.
    * Update the `init()` function.
* Update the `SCHEMA_VERSION` variable in `<aqueduct repo>/src/python/bin/aqueduct`.
* Update the `CurrentSchemaVersion` variable in `<aqueduct repo>/src/golang/lib/models/schema_version.go`.