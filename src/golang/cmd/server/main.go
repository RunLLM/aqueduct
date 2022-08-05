package main

import (
	"flag"
	"io/ioutil"
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
	expose        = flag.Bool("expose", false, "Whether the server will be exposed to the public.")
	verbose       = flag.Bool("verbose", false, "Whether all logs will be shown in the terminal.")
	port          = flag.Int("port", connection.ServerInternalPort, "The port that the server listens to.")
	serverLogPath = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "logs", "server_log")
)

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

	if !*verbose {
		log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

		log.AddHook(&writer.Hook{ // Send logs with level higher than error to stderr
			Writer: os.Stderr,
			LogLevels: []log.Level{
				log.PanicLevel,
				log.FatalLevel,
				log.ErrorLevel,
			},
		})
		log.AddHook(&writer.Hook{ // Send info, debug, warning logs to the log file
			// With lumberjack.Logger we can do log rotation to prevent it from growing infinitely.
			Writer: &lumberjack.Logger{
				Filename:   serverLogPath,
				MaxSize:    100, // megabytes
				MaxBackups: 3,
				MaxAge:     28, // days
			},
			LogLevels: []log.Level{
				log.InfoLevel,
				log.DebugLevel,
				log.WarnLevel,
			},
		})
	}

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
	s.Run(*expose, *port)
}
