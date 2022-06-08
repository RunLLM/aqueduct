package main

import (
	"flag"
	"fmt"

	"github.com/aqueducthq/aqueduct/cmd/migrator/migrator"
	"github.com/aqueducthq/aqueduct/lib/database"
	log "github.com/sirupsen/logrus"
)

var (
	dbType = flag.String("type", "postgres", "The type of database to connect to: postgres or sqlite.")

	// Postgres Config
	host = flag.String("host", "", "The host of the Postgres server to connect to.")
	port = flag.String("port", "5432", "The port number to connect to the Postgres database.")
	user = flag.String("username", "postgres-username", "The username for connecting to the Postgres database.")
	pwd  = flag.String("password", "postgres-password", "The password for connecting to the Postgres database.")
	db   = flag.String("database", "aqueduct", "The Postgres database to connect to.")
	file = flag.String("file", "", "The sqlite file path")
)

const (
	createUsage  = `create NAME TYPE       Creates a migration script with NAME of TYPE [sql|go].`
	gotoUsage    = `goto V                 Migrate to version V.`
	upUsage      = `up                     Migrate up one version.`
	downUsage    = `down                   Migrate down one version.`
	versionUsage = `version                Print out the current version.`

	createCmd  = "create"
	gotoCmd    = "goto"
	upCmd      = "up"
	downCmd    = "down"
	versionCmd = "version"

	createErrMsg = "create must be of form: migrate create NAME [sql|go]"
	goToErrMsg   = "goto must be of form: migrate goto V"
)

func printUsage() {
	fmt.Printf(`Usage: migrate OPTIONS COMMAND [arg...]
		OPTIONS:
			- help 		Print usage
			- type 		The SQL database dialect: postgres or sqlite.
			- host 		IP Address or hostname of Postgres instance to connect to
			- port 		The port number that the Postgres instance is running at.
			- username 	Username for connecting to Postgres
			- password 	Password for connecting to Postgres
			- database 	The Postgres database to connect to
		
		COMMANDS:
			%s 
			%s
			%s
			%s
			%s
	`, createUsage, gotoUsage, upUsage, downUsage, versionUsage)
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	databaseType := database.Type(*dbType)
	if databaseType != database.PostgresType && databaseType != database.SqliteType {
		log.Fatalf("Unknown database type specified: %v", databaseType)
	}

	var pgxConfig *database.PostgresConfig = nil
	if databaseType == database.PostgresType {
		pgxConfig = &database.PostgresConfig{
			Address:  *host,
			UserName: *user,
			Password: *pwd,
			Database: *db,
			Port:     *port,
		}
	}

	var sqliteConfig *database.SqliteConfig = nil
	if databaseType == database.SqliteType {
		sqliteConfig = &database.SqliteConfig{File: *file}
	}

	conf := &database.DatabaseConfig{
		Type:     databaseType,
		Postgres: pgxConfig,
		Sqlite:   sqliteConfig,
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		log.Fatal("Command was not specified.")
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case createCmd:
		if len(cmdArgs) != 2 {
			log.Fatal(createErrMsg)
		}
		migrator.HandleCreate(cmdArgs)
	case gotoCmd:
		if len(cmdArgs) != 1 {
			log.Fatal(goToErrMsg)
		}
		migrator.HandleGoTo(cmdArgs, conf)
	case upCmd:
		migrator.HandleUp(conf)
	case downCmd:
		migrator.HandleDown(conf)
	case versionCmd:
		migrator.HandleVersion(conf)
	default:
		printUsage()
		log.Fatal("Unknown command specified.")
	}
}
