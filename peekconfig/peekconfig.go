package peekconfig

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const configFile string = "peek.yml"

// Service defines the configuration options for an individual FeaturePeek service
type Service struct {
	Type string
	Path string
	Spa  bool
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

// SimpleService is a simple representation of a service with a dynamic name
type SimpleService struct {
	Path string
	Name string
}

// LoadStaticServiceFromFile attempts to load a specific static service from the peek.yml file.
func LoadStaticServiceFromFile(filename string, serviceName string) (*SimpleService, error) {
	data, err := ReadConfigFile(filename)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	var service *SimpleService

	for k, v := range config {
		if k == "version" {
			continue
		}

		s, ok := v.(map[interface{}]interface{})
		if !ok {
			continue
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

		if serviceType == "static" && (serviceName == "" || serviceName == k) {
			service = new(SimpleService)
			service.Name = k
			service.Path = servicePath
			break
		}
	}

	return service, nil
}

// LoadFromFile attempts to populate a SimpleService struct from the given peek.yml file.
func LoadFromFile(filename string) (*SimpleService, error) {
	return LoadStaticServiceFromFile(filename, "")
}

// ReadConfigFile reads and returns the contents of the given file (mockable)
var ReadConfigFile = func(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}
