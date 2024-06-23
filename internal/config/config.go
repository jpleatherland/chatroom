package config

import (
	"encoding/json"
	"os"
	"path"
)

/*
./config.json example
{
  "host": "localhost",
  "port": "42069",
  "databaseFile": "/absolute/path/to/test.db",
  "hostKeyPath": "/absolute/path/to/.ssh/id_rsa"
}
*/

type Config struct {
	Host         string
	Port         string
	DatabaseFile string
	HostKeyPath  string
}

func NewConfig() (*Config, error) {
	var cf *Config
	confFileDir, _ := os.Getwd()
	confFile, err := os.Open(path.Join(confFileDir, "config.json"))
	if err != nil {
		return nil, err
	}

	defer confFile.Close()

	jsonParser := json.NewDecoder(confFile)
	decErr := jsonParser.Decode(&cf)

	if decErr != nil {
		return nil, err
	}

	return cf, nil
}
