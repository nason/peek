package auth

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// Auth respresents the json auth structure returned from Auth0 which is also stored locally on login
type Auth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Save will marshall and save the auth in json to disk
func (a Auth) Save() (err error) {
	configPath, err := peekConfigDir()
	if err != nil {
		return
	}
	os.MkdirAll(configPath, 0755)
	tokensPath := filepath.Join(configPath, "tokens.json")

	data, err := json.Marshal(a)
	if err != nil {
		return
	}

	if err = ioutil.WriteFile(tokensPath, data, 0600); err != nil {
		return
	}
	return
}

// LoadFromFile attempts to populate an Auth object from the tokens.json file.
func LoadFromFile() Auth {
	configDir, err := peekConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(configDir, "tokens.json")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("No credentials found. Run `peek login` to login with your FeaturePeek account.")
		} else {
			log.Fatalf("Unable to read authorization file: %v.", err)
		}
	}

	var tokens Auth
	if err = json.Unmarshal(data, &tokens); err != nil {
		log.Fatalf("Unable to parse authorization file: %v.", err)
	}
	return tokens
}

// RemoveFile attempts to remove the auth file from the filesystem
func RemoveFile() {
	// look for auth file
	configDir, err := peekConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(configDir, "tokens.json")
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Unable to remove authorization file: %v.", err)
	}
}

func peekConfigDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "peek"), nil

}
