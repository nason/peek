package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
func (c Config) Save() (err error) {
	fullConfigFile := configFile
	// fullConfigFile, err := GetConfigFilePath()
	// if err != nil {
	// 	return err
	// }

	data, err := yaml.Marshal(c)
	if err != nil {
		return
	}

	if err = ioutil.WriteFile(fullConfigFile, data, 0644); err != nil {
		return
	}

	return
}

// LoadFromFile attempts to populate a Config object from the peek.yml file.
func LoadFromFile(dir string) Config {
	data, err := ioutil.ReadFile(dir + "/peek.yml")
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

func findConfigFile(dir string) []byte {
	data, err := ioutil.ReadFile(filepath.Join(dir, "peek.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			if filepath.Dir(dir) == dir {
				log.Fatal("No peek.yml config found.\n\nRun `peek init` to create one!")
			} else {
				return findConfigFile(filepath.Dir(dir))
			}
		} else {
			log.Fatalf("Unable to read config file: %v.", err)
		}
	}
	return data
}

// LoadStaticServiceFromFile attempts to populate a Service object from the peek.yml file.
func LoadStaticServiceFromFile(dir string) (*Service, string) {
	data := findConfigFile(dir)

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Unable to decode config file: %v.", err)
	}

	var staticService *Service
	var serviceName string

	for k, v := range config {
		if k == "version" {
			continue
		}

		s, ok := v.(map[interface{}]interface{})
		if !ok {
			log.Fatalln("Error parsing peek.yml")
		}

		serviceConfig := make(map[string]string)
		for sk, sv := range s {
			strKey := fmt.Sprintf("%v", sk)
			strVal := fmt.Sprintf("%v", sv)

			serviceConfig[strKey] = strVal
		}

		serviceType, ok := serviceConfig["type"]
		if !ok {
			continue
		}
		servicePath, ok := serviceConfig["path"]
		if !ok {
			continue
		}
		if serviceType == "static" {
			serviceName = k
			staticService = new(Service)
			staticService.Type = serviceType
			staticService.Path = servicePath
			break
		}
	}

	return staticService, serviceName
}
