package config

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const configFile string = "peek.yml"

// Service defines the configuration options for an individual FeaturePeek service
type Service struct {
	Type string
	Path string
}

// Config defines the configuration options for a FeaturePeek project
type Config struct {
	Version int
	Main    Service
}

// Save will marshal and save the peek.yml config to disk
func (c Config) Save() error {
	fullConfigFile := configFile
	// fullConfigFile, err := GetConfigFilePath()
	// if err != nil {
	// 	return err
	// }

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(fullConfigFile, data, 0644); err != nil {
		return err
	}

	return nil
}

// LoadFromFile attempts to populate a Config object from the peek.yml file.
func LoadFromFile() Config {
	data, err := ioutil.ReadFile("./peek.yml")
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("No peek.yml config found.\n\nRun `peek init` to create one!")
		} else {
			log.Fatalf("Unable to read config file: %v.", err)
		}
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Unable to decode config file: %v.", err)
	}

	return config
}
