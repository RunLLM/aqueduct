package config

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ServerConfiguration struct {
	AqPath             string `yaml:"aqPath" json:"aq_path"`
	EncryptionKey      string `yaml:"encryptionKey" json:"encryption_key"`
	RetentionJobPeriod string `yaml:"retentionJobPeriod"`
	ApiKey             string `yaml:"apiKey"`
}

func ParseServerConfiguration(confPath string) *ServerConfiguration {
	bts, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatal("Unable to read server config.yml. Please make sure that the config is properly configured and retry: ", err)
		os.Exit(1)
	}

	var config ServerConfiguration
	err = yaml.Unmarshal(bts, &config)
	if err != nil {
		log.Fatal("Unable to correctly parse server config.yml. Please check the config file and retry: ", err)
		os.Exit(1)
	}

	return &config
}
