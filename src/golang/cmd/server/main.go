package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	"github.com/aqueducthq/aqueduct/cmd/server/server"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/connection"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	confPath = flag.String(
		"config",
		"",
		"The path to .yml config file",
	)
	expose            = flag.Bool("expose", false, "Whether the server will be exposed to the public.")
	verbose           = flag.Bool("verbose", false, "Whether all logs will be shown in the terminal, with filepaths and line numbers.")
	port              = flag.Int("port", connection.ServerInternalPort, "The port that the server listens to.")
	serverLogPath     = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "logs", "server")
	disableUsageStats = flag.Bool("disable-usage-stats", false, "Whether to disable usage statistics reporting.")

	envPath             = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "config", "env")
	allowedEnvironments = map[string]bool{"dev": true, "test": true, "prod": true}
)

func parseEnv() string {
	var environment string
	if _, err := os.Stat(envPath); err == nil {
		b, err := os.ReadFile(envPath)
		if err != nil {
			log.Fatalf("Unexpected error when reading server environment config.: %v", err)
		}

		environment = string(b)

		_, ok := allowedEnvironments[environment]
		if !ok {
			os.Remove(envPath)
			log.Fatalf("Unsupported environment: %v", environment)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		environment = "prod"
	} else {
		log.Fatalf("Unexpected error when reading server environment config.: %v", err)
	}

	return environment
}

func main() {
	flag.Parse()

	// Load all configs from `config.yml` file
	if *confPath == "" {
		cwd, _ := os.Getwd()
		*confPath = filepath.Join(cwd, "config", "server.yml")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableQuote: true,
		ForceColors:  true,
	})

	// Always store all logs to a log file.
	// With lumberjack.Logger we can do log rotation to prevent it from growing infinitely.
	log.SetOutput(&lumberjack.Logger{
		Filename:   serverLogPath,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	// Send logs with level higher than warning to stderr.
	log.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})

	if *verbose {
		// If verbose, also send info and debug logs to stdout.
		log.AddHook(&writer.Hook{
			Writer: os.Stdout,
			LogLevels: []log.Level{
				log.InfoLevel,
				log.DebugLevel,
			},
		})

		// Also print the filepath and line number.
		log.SetReportCaller(true)
	}

	if err := config.Init(*confPath); err != nil {
		log.Fatalf("Failed to initialize server config: %v", err)
	}

	environment := parseEnv()

	s := server.NewAqServer(environment, *disableUsageStats)

	err := s.StartWorkflowRetentionJob(config.RetentionJobPeriod())
	if err != nil {
		log.Fatalf("Failed to start workflow retention cronjob: %v", err)
	}

	err = s.RunMissedCronJobs()
	if err != nil {
		log.Errorf("Failed to run missed workflows: %v", err)
	}

	// Start the HTTP server and listen for requests indefinitely.
	log.Infof("You can use api key %s to connect to the server", config.APIKey())
	s.Run(*expose, *port)
}
