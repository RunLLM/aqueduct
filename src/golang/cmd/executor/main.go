package main

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/aqueducthq/aqueduct/cmd/executor/executor"
	"github.com/aqueducthq/aqueduct/lib/job"
	log "github.com/sirupsen/logrus"
)

const (
	jobSpecFlagKey = "spec"
)

var spec = flag.String(
	jobSpecFlagKey,
	"",
	"The json-serialized cronjob spec to execute.",
)

func init() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{DisableQuote: true})
}

func main() {
	if err := run(); err != nil {
		log.Errorf("Failure when running executor: %v", err)
	}
}

func run() error {
	ctx := context.TODO()
	spec, err := job.DecodeSpec(*spec, job.GobSerializationType)
	if err != nil {
		return err
	}

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
	return ex.Run(ctx)
}
