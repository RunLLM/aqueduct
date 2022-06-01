package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aqueducthq/aqueduct/lib/database"
	log "github.com/sirupsen/logrus"
)

func handleCreate(args []string) {
	name, language := args[0], ScriptLanguage(args[1])
	if language != SqlScriptLanguage && language != GoScriptLanguage {
		log.Fatalf("Unknown script language specified. %s", createErrMsg)
	}

	if err := Create(name, language); err != nil {
		log.Fatalf("Unexpected error running create: %v", err)
	}
}

func handleGoTo(args []string, conf *database.DatabaseConfig) {
	versionStr := args[0]
	version, err := strconv.ParseInt(versionStr, 0, 64)
	if err != nil {
		log.Fatalf("Unable to parse version number: %v.", err)
	}

	db := createDatabaseClient(conf)
	defer db.Close()

	if err := GoTo(context.Background(), version, db); err != nil {
		log.Errorf("Unexpected error running goto: %v", err)
	}

	log.Info("Checking current schema version...")
	handleVersion(conf)
}

func handleUp(conf *database.DatabaseConfig) {
	db := createDatabaseClient(conf)
	defer db.Close()

	if err := Up(context.Background(), db); err != nil {
		log.Errorf("Unexpected error running up: %v", err)
	}

	log.Info("Checking current schema version...")
	handleVersion(conf)
}

func handleDown(conf *database.DatabaseConfig) {
	db := createDatabaseClient(conf)
	defer db.Close()

	if err := Down(context.Background(), db); err != nil {
		log.Errorf("Unexpected error running down: %v", err)
	}

	log.Info("Checking current schema version...")
	handleVersion(conf)
}

func handleVersion(conf *database.DatabaseConfig) {
	db := createDatabaseClient(conf)
	defer db.Close()

	version, dirty, err := Version(context.Background(), db)
	if err != nil {
		log.Fatalf("Unexpected error running version: %v", err)
	}

	if dirty {
		log.Errorf(
			"The current schema version %s is dirty. Please take steps to resolve this.",
			fmt.Sprintf("%06d", version),
		)
	} else {
		log.Infof("The current schema version is %s.", fmt.Sprintf("%06d", version))
	}
}

// createDatabaseClient creates a database.Database client based on config provided.
func createDatabaseClient(conf *database.DatabaseConfig) database.Database {
	db, err := database.NewDatabase(conf)
	if err != nil {
		log.Fatalf("Unable to create database client: %v", err)
	}

	return db
}
