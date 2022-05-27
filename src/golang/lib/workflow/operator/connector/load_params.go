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
