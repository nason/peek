// Package cmd defines the primary functionality of the CLI
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"peek/auth"
	"peek/config"
	"peek/spinner"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	jose "gopkg.in/square/go-jose.v2"
)

const devAuth0BaseURL = "https://featurepeek-dev.auth0.com"
const prodAuth0BaseURL = "https://login.featurepeek.com"
const devClientID = "XnNVx0nzQSJdY6ksPGTnnciuGOM8kXsT"
const prodClientID = "oB2RkLUylDTrsSxVa6qdLR3DQMbdh9IR"

var clientID string
var auth0BaseURL string

func userAPIPostForm(tokens auth.Auth) {
	var apiURL string
	if devFlag {
		apiURL = "https://api.dev.featurepeek.com/api/v1/user"
	} else {
		apiURL = "https://api.featurepeek.com/api/v1/user"
	}

	request, err := http.NewRequest("POST", apiURL, strings.NewReader(""))
	request.Header.Add("authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))
	request.Header.Add("X-FEATUREPEEK-CLIENT", Version)
	if debugFlag {
		fmt.Printf("%+v\n", request)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != http.StatusOK &&
		response.StatusCode != http.StatusCreated {
		log.Fatalf("\nCall to FeaturePeek API failed %d", response.StatusCode)
	}
}

func auth0PostForm(reqPath string, data url.Values) (int, []byte, error) {
	if devFlag {
		clientID = devClientID
	} else {
		clientID = prodClientID
	}
	data.Set("client_id", clientID)

	if devFlag {
		auth0BaseURL = devAuth0BaseURL
	} else {
		auth0BaseURL = prodAuth0BaseURL
	}

	reqURL := fmt.Sprintf("%s/oauth%s", auth0BaseURL, reqPath)

	if debugFlag {
		fmt.Printf("-Request-\nurl -> %s\ndata -> %s\n\n", reqURL, data)
	}

	res, err := http.DefaultClient.PostForm(reqURL, data)
	if err != nil {
		return 0, nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	if debugFlag {
		fmt.Printf("-Response-\nObj -> %+v\nbody -> %s\n\n", *res, body)
	}

	return res.StatusCode, body, nil
}

func auth0Get(reqPath string) ([]byte, error) {
	if devFlag {
		auth0BaseURL = devAuth0BaseURL
	} else {
		auth0BaseURL = prodAuth0BaseURL
	}

	u, err := url.Parse(auth0BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, reqPath)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func loginCommand(cmd *cobra.Command, args []string) {

	var oauthAudience string
	if devFlag {
		oauthAudience = "http://api.dev.featurepeek.com/api/v1/"
	} else {
		oauthAudience = "http://api.featurepeek.com/api/v1/"
	}
	data := url.Values{}
	data.Set("scope", "offline_access")
	data.Set("audience", oauthAudience)

	statusCode, body, err := auth0PostForm("/device/code", data)
	if err != nil {
		log.Fatal(err)
	}

	if statusCode != http.StatusOK {
		log.Fatalf("Auth request failed:\n%s\n", body)
	}

	var resp struct {
		DeviceCode              string        `json:"device_code"`
		UserCode                string        `json:"user_code"`
		VerificationURI         string        `json:"verification_uri"`
		ExpiresIn               int           `json:"expires_in"`
		Interval                time.Duration `json:"interval"`
		VerificationURIComplete string        `json:"verification_uri_complete"`
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Fatal(err)
	}

	// print user code that must match on auth screen
	fmt.Println(resp.UserCode)

	// launch browser to user code confirmation screen
	open.Start(resp.VerificationURIComplete)

	// start spinner
	loginSpinner := spinner.New("Logging in")
	loginSpinner.Start()

	// Grab jwks from Auth0
	jwksBody, err := auth0Get(".well-known/jwks.json")
	if err != nil {
		log.Fatal(err)
	}

	var jwks jose.JSONWebKeySet
	if err = json.Unmarshal(jwksBody, &jwks); err != nil {
		log.Fatal(err)
	}
	jwk := jwks.Keys[0]

	// poll for access token
	var tokenBody []byte
	data = url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", resp.DeviceCode)

	for range time.Tick(time.Second * resp.Interval) {
		statusCode, body, err = auth0PostForm("/token", data)
		if err != nil {
			log.Fatal(err)
		}

		if statusCode == http.StatusOK {
			tokenBody = body
			break
		}

		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err = json.Unmarshal(body, &errResp); err != nil {
			log.Fatal(err)
		}

		if errResp.Error == "expired_token" || errResp.Error == "access_denied" {
			log.Fatal(errResp.ErrorDescription)
		}
	}

	// verify jwt
	var tokens auth.Auth
	if err = json.Unmarshal(tokenBody, &tokens); err != nil {
		log.Fatal(err)
	}

	object, err := jose.ParseSigned(tokens.AccessToken)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = object.Verify(&jwk); err != nil {
		log.Fatal(err)
	}

	// save auth to config
	if err = config.SaveAuthToConfigFile(tokens, devFlag); err != nil {
		log.Fatal(err)
	}

	// check in with the api
	userAPIPostForm(tokens)

	loginSpinner.Stop()
	fmt.Println("Logged in to FeaturePeek")
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with your FeaturePeek account",
	Long: `Login to your FeaturePeek account.

This command will send the user through an authentication flow that
will authorize the CLI on the user's behalf. If the user does not have
a FeaturePeek account, one will be created in this flow.`,
	Run: loginCommand,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
