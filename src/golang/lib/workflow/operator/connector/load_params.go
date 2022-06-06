package connector

type LoadParams interface {
	isLoadParams()
}

type RelationalDBLoadParams struct {
	Table      string `json:"table"`
	UpdateMode string `json:"update_mode"`
}

type PostgresLoadParams struct{ RelationalDBLoadParams }

type SnowflakeLoadParams struct{ RelationalDBLoadParams }

type MySqlLoadParams struct{ RelationalDBLoadParams }

type RedshiftLoadParams struct{ RelationalDBLoadParams }

type MariaDbLoadParams struct{ RelationalDBLoadParams }

type SqlServerLoadParams struct{ RelationalDBLoadParams }

type BigQueryLoadParams struct{ RelationalDBLoadParams }

type SqliteLoadParams struct{ RelationalDBLoadParams }

type GoogleSheetsLoadParams struct {
	Filepath string `json:"filepath"`
	SaveMode string `json:"save_mode"`
}

type SalesforceLoadParams struct {
	Object string `json:"object"`
}

type S3LoadParams struct {
	Filepath string `json:"filepath"`
	Format   string `json:"format"`
}

func CastToRelationalDBLoadParams(params LoadParams) (*RelationalDBLoadParams, bool) {
	postgres, ok := params.(*PostgresLoadParams)
	if ok {
		return &postgres.RelationalDBLoadParams, true
	}

	snowflake, ok := params.(*SnowflakeLoadParams)
	if ok {
		return &snowflake.RelationalDBLoadParams, true
	}

	mysql, ok := params.(*MySqlLoadParams)
	if ok {
		return &mysql.RelationalDBLoadParams, true
	}

	redshift, ok := params.(*RedshiftLoadParams)
	if ok {
		return &redshift.RelationalDBLoadParams, true
	}

	mariadb, ok := params.(*MariaDbLoadParams)
	if ok {
		return &mariadb.RelationalDBLoadParams, true
	}

	sqlserver, ok := params.(*SqlServerLoadParams)
	if ok {
		return &sqlserver.RelationalDBLoadParams, true
	}

	bigquery, ok := params.(*BigQueryLoadParams)
	if ok {
		return &bigquery.RelationalDBLoadParams, true
	}

	sqlite, ok := params.(*SqliteLoadParams)
	if ok {
		return &sqlite.RelationalDBLoadParams, true
	}

	return nil, false
}

func (*PostgresLoadParams) isLoadParams() {}

func (*SnowflakeLoadParams) isLoadParams() {}

func (*MySqlLoadParams) isLoadParams() {}

func (*RedshiftLoadParams) isLoadParams() {}

func (*MariaDbLoadParams) isLoadParams() {}

func (*SqlServerLoadParams) isLoadParams() {}

func (*BigQueryLoadParams) isLoadParams() {}

func (*SqliteLoadParams) isLoadParams() {}

func (*GoogleSheetsLoadParams) isLoadParams() {}

func (*SalesforceLoadParams) isLoadParams() {}

func (*S3LoadParams) isLoadParams() {}
