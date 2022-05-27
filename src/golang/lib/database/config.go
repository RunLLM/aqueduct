package database

type Type string

const (
	PostgresType = "postgres"
	SqliteType   = "sqlite"
	NoopType     = "noop"
)

type PostgresConfig struct {
	Address  string `yaml:"address" json:"address"`
	UserName string `yaml:"userName" json:"user_name"`
	Password string `yaml:"password" json:"password"`
	Database string `yaml:"database" json:"database"`
	Port     string `yaml:"port" json:"port,omitempty"`
}

type SqliteConfig struct {
	File string `yaml:"file" json:"file"`
}

type DatabaseConfig struct {
	Type     Type            `yaml:"type" json:"type"`
	Postgres *PostgresConfig `yaml:"postgres" json:"postgres,omitempty"`
	Sqlite   *SqliteConfig   `yaml:"sqlite" json:"sqlite,omitempty"`
}
