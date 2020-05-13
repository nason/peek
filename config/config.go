package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"peek/auth"

	"github.com/mitchellh/go-homedir"
)

// Dir returns the config directory
func Dir() string {
	dir, _ := homedir.Expand("~/.config/peek")
	return dir
}

// File returns the full path of the config file
func File(dev bool) string {
	var filename string
	if dev {
		filename = "dev-config.json"
	} else {
		filename = "config.json"
	}
	return path.Join(Dir(), filename)
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

// Config represents the CLI configuration
type Config struct {
	Auth *auth.Auth `json:"auth"`
}

// LoadConfig will load the appropriate config given the dev flag
func LoadConfig(devFlag bool) (*Config, error) {
	return ParseConfigFile(File(devFlag))
}

// ParseConfigFile attempts to load the given config file
func ParseConfigFile(filename string) (*Config, error) {
	data, err := ReadConfigFile(filename)
	if err != nil {
		return nil, err
	}

	var configData Config
	if err = json.Unmarshal(data, &configData); err != nil {
		return nil, err
	}

	return &configData, nil
}

// SaveAuthToConfigFile will marshall and save auth to the config file
func SaveAuthToConfigFile(newAuth auth.Auth, devFlag bool) error {
	// create config dir
	os.MkdirAll(Dir(), 0755)

	data, err := json.Marshal(Config{Auth: &newAuth})
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(File(devFlag), data, 0600); err != nil {
		return err
	}
	return nil
}

// RemoveAuthFromConfigFile attempts to remove the auth file from the filesystem
func RemoveAuthFromConfigFile(devFlag bool) error {
	err := os.Remove(File(devFlag))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
