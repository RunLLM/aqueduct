package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/aqueducthq/aqueduct/cmd/executor/executor"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/job"
	log "github.com/sirupsen/logrus"
)

const (
	jobSpecFlagKey      = "spec"
	logsFilePathFlagKey = "logs-path"
)

var specSerialized = flag.String(
	jobSpecFlagKey,
	"",
	"The json-serialized cronjob spec to execute.",
)

var logsFilePath = flag.String(
	logsFilePathFlagKey,
	"",
	"The path to the file the executor will log to. If not set, we'll log to stdout/err.",
)

var confPath = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "config", "config.yml")

func init() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{DisableQuote: true})

	// Create the directory for the logs if it doesn't already exist.
	if len(*logsFilePath) > 0 {
		logsDir := filepath.Dir(*logsFilePath)
		if _, err := os.Stat(logsDir); os.IsNotExist(err) {
			err = os.Mkdir(logsDir, os.ModePerm)
			if err != nil {
				log.Infof("Unable to make directory %s: %v ", logsDir, err)
			}
		}
	}

	// Initialize config, the process should exit if this fails
	if err := config.Init(confPath); err != nil {
		log.Fatalf("Unable to initialize config: %v", err)
	}
}

func redirectLogOutput(filepath string) (*os.File, error) {
	logsFile, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening log file.")
	}
	log.SetOutput(logsFile)
	return logsFile, nil
}

func main() {
	if len(*logsFilePath) > 0 {
		logFile, err := redirectLogOutput(*logsFilePath)
		if err != nil {
			log.Errorf("Unable to redirect log output. %v", err)
			return
		}
		defer logFile.Close()
	}

	spec, err := job.DecodeSpec(*specSerialized, job.GobSerializationType)
	if err != nil {
		log.Errorf("Unable to decode spec. %v", err)
		return
	}

	if err := run(spec); err != nil {
		log.Errorf("Failure when running executor: %v", err)
	}
}

func run(spec job.Spec) error {
	logBytes, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	log.Info(string(logBytes))

	ex, err := executor.NewExecutor(spec)
	if err != nil {
		return err
	}
	defer ex.Close()

	ctx := context.TODO()
	return ex.Run(ctx)
}
