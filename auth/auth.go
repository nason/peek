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
func (a Auth) Save() error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", "peek")
	os.MkdirAll(path, 0755)
	tokensPath := filepath.Join(path, "tokens.json")

	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(tokensPath, data, 0600); err != nil {
		return err
	}
	return nil
}

// LoadFromFile attempts to populate an Auth object from the tokens.json file.
func LoadFromFile() Auth {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(home, ".config", "peek", "tokens.json")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("No authorization file found. Run `peek login` to authorize with your FeaturePeek account.")
		} else {
			log.Fatalf("Unable to read authorization file: %v.", err)
		}
	}

	var tokens Auth
	if err = json.Unmarshal(data, &tokens); err != nil {
		log.Fatal(err)
	}

	return tokens
}
