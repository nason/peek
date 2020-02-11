/*
Copyright © 2020 Landon Spear <landon@featurepeek.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

// Package cmd defines the primary functionality of the CLI
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"peek/auth"
	"peek/config"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const releaseVersion = "v0.2-alpha.4"

const peekCommandLongDesc = `peek is a command-line tool for interacting with FeaturePeek environments.

The FeaturePeek CLI enables front-end developers, designers, and product owners to interact with and review your changes
by launching new running previews of the front-end code you are working on.

The goal of this tool is to enable launching new environments for every branch and commit, all without needing a CI system.

To get started simply run ` + "`peek login`" + `to authenticate locally and/or create an account.
Run ` + "`peek init`" + ` and enter your build directory to set up your config.
Make sure your code pushed to your remote and run your build step.
Then run ` + "`peek`" + ` to launch your FeaturePeek environment.`

var cfgFile string
var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "peek",
	Short: "FeaturePeek Command-line Tool",
	Long:  peekCommandLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		// Load auth and config files
		tokens := auth.LoadFromFile()
		peekConfig := config.LoadFromFile()

		if peekConfig.Main.Type != "static" {
			log.Fatal("FeaturePeek CLI does not currently support non-static configurations")
		}

		assetPath := peekConfig.Main.Path

		if assetPath == "" {
			log.Fatal("Invalid Path for static assets in config.")
		}

		// Read info out of local git repo
		r, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Cannot read git repository: %v", err)
		}

		headRef, err := r.Head()
		if err != nil {
			log.Fatalf("Cannot read git repository HEAD: %v", err)
		}

		// Cannot be in detatched HEAD state
		if !headRef.Name().IsBranch() {
			log.Fatal("Cannot find current branch. Environments must refence a branch")
		}

		sha := headRef.Hash().String()
		branch := headRef.Name().Short()
		originRev, err := r.ResolveRevision(plumbing.Revision("refs/remotes/origin/" + branch))
		if err != nil {
			log.Fatalf("Error: local branch does not match origin - %s", err)
		}

		if originRev.String() != sha {
			log.Fatal("Error: origin branch does not match local branch. You may need to push your changes.")
		}

		remote, err := r.Remote("origin")
		if err != nil {
			log.Fatalf("Error reading git remote `origin`: %v", err)
		}
		if err = remote.Config().Validate(); err != nil {
			log.Fatalf("git config error: %v", err)
		}
		remoteURL := remote.Config().URLs[0]
		endpoint, err := transport.NewEndpoint(remoteURL)
		if err != nil {
			log.Fatalf("git remote parse error: %v", err)
		}
		if endpoint.Host != "github.com" {
			log.Fatalf("%s is not currently a supported git hosting platform.\nContact us at support@featurepeek.com to request adding your git host!", endpoint.Host)
		}

		slug := endpoint.Path[:len(endpoint.Path)-4]
		splitSlug := strings.Split(slug, "/")
		org := splitSlug[0]
		repo := splitSlug[1]

		// Archive web asset directory
		files, err := ioutil.ReadDir(assetPath)
		if err != nil {
			log.Fatal(err)
		}

		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, filepath.Join(assetPath, file.Name()))
		}

		err = archiver.Archive(fileNames, "artifacts.tar.gz")
		if err != nil {
			_ = os.Remove("artifacts.tar.gz")
			log.Fatal(err)
		}

		// Send ping
		file, err := os.Open("artifacts.tar.gz")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("artifacts", file.Name())
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(part, file)

		writer.WriteField("app", "main")
		writer.WriteField("service", "cli")
		writer.WriteField("org", org)
		writer.WriteField("repo", repo)
		writer.WriteField("sha", sha)
		writer.WriteField("branch", branch)
		if err = writer.Close(); err != nil {
			log.Fatal(err)
		}

		request, err := http.NewRequest("POST", "https://api.dev.featurepeek.com/api/v1/peek", body)
		request.Header.Add("authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))
		request.Header.Add("X-FEATUREPEEK-CLIENT", releaseVersion)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		if debug {
			fmt.Printf("%+v\n", request)
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			log.Fatal(err)
		}

		body = &bytes.Buffer{}
		defer response.Body.Close()
		_, err = body.ReadFrom(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		if debug {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Header)
		}
		fmt.Println("Assets uploaded successfully!\nVisit your new feature environment here:")
		fmt.Println(body)

		// Clean up artifacts archive
		err = os.Remove("artifacts.tar.gz")
		if err != nil {
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log.SetFlags(0)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.peek.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".peek" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".peek")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
