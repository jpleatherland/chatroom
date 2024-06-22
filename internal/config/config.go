package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

/*
./config.json example
{
  "host": "localhost",
  "port": "42069",
  "databaseFile": "test.db",
  "hostKeyPath": ".ssh/id_ed10000"
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
	fmt.Println(confFileDir)
	fmt.Println(path.Join(confFileDir, "config.json"))
	confFile, err := os.Open(path.Join(confFileDir, "config.json"))
	if err != nil {
		return nil, err
	}

	defer confFile.Close()

	jsonParser := json.NewDecoder(confFile)
	decErr := jsonParser.Decode(&cf)
	cf.DatabaseFile = path.Join(confFileDir, cf.DatabaseFile)
	cf.HostKeyPath = path.Join(confFileDir, cf.HostKeyPath)

	if decErr != nil {
		return nil, err
	}

	fmt.Println("databasefile: ", cf.DatabaseFile)

	return cf, nil
}
