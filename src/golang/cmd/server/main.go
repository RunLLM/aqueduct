package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/aqueducthq/aqueduct/cmd/server/server"
	"github.com/aqueducthq/aqueduct/config"
	log "github.com/sirupsen/logrus"
)

var (
	confPath = flag.String(
		"config",
		"",
		"The path to .yml config file",
	)
	expose = flag.Bool("expose", false, "Whether you want to expose the server to the public.")
)

func main() {
	flag.Parse()

	// Load all configs from `config.yml` file
	if *confPath == "" {
		cwd, _ := os.Getwd()
		*confPath = filepath.Join(cwd, "config", "server.yml")
	}

	log.SetFormatter(&log.TextFormatter{DisableQuote: true})

	serverConfig := config.ParseServerConfiguration(*confPath)

	s := server.NewAqServer(serverConfig)

	err := s.StartWorkflowRetentionJob(serverConfig.RetentionJobPeriod)
	if err != nil {
		log.Fatalf("Failed to start workflow retention cronjob: %v", err)
	}

	err = s.RunMissedCronJobs()
	if err != nil {
		log.Errorf("Failed to run missed workflows: %v", err)
	}

	// Start the HTTP server and listen for requests indefinitely.
	log.Infof("You can use api key %s to connect to the server", serverConfig.ApiKey)
	s.Run(*expose)
}
