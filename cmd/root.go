// Package cmd defines the primary functionality of the CLI
package cmd

import (
	"bytes"
	"crypto/md5"
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
	"peek/context"
	"peek/git"
	"peek/spinner"
	"runtime/debug"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Version is dynamically set by the toolchain.
var Version = "DEV"

var versionOutput = ""

func init() {
	cobra.OnInitialize(initConfig)

	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	Version = strings.TrimPrefix(Version, "v")
	rootCmd.Version = Version
	rootCmd.AddCommand(versionCmd)
	versionOutput = fmt.Sprintf("peek version %s", rootCmd.Version)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/peek/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&targetDir, "dir", "", "target directory to launch from")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "debug output")
	rootCmd.PersistentFlags().BoolVar(&devFlag, "dev", false, "dev use")
	rootCmd.PersistentFlags().MarkHidden("dev")
	rootCmd.Flags().Bool("version", false, "Show peek version")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var versionCmd = &cobra.Command{
	Use:    "version",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionOutput)
	},
}

const peekCommandLongDesc = `peek is a command-line tool for interacting with FeaturePeek environments.

The FeaturePeek CLI enables front-end developers, designers, and product owners to interact with and review your changes
by launching new running previews of the front-end code you are working on.

The goal of this tool is to enable launching new environments for every branch and commit, all without needing a CI system.

To get started simply run ` + "`peek login`" + `to authenticate locally and/or create an account.
Run ` + "`peek init`" + ` and enter your build directory to set up your config.
Make sure your code pushed to your remote and run your build step.
Then run ` + "`peek`" + ` to launch your FeaturePeek environment.`

var cfgFile string
var debugFlag bool
var devFlag bool
var targetDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "peek",
	Short: "FeaturePeek Command-line Tool",
	Long:  peekCommandLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		if targetDir != "" {
			currentDir, err := os.Getwd()
			if err != nil {
				log.Fatalf("Could not get current directory: %v", err)
			}
			if err = os.Chdir(targetDir); err != nil {
				log.Fatalf("Could not open target directory: %v", err)
			}
			defer os.Chdir(currentDir)
		}

		// Load auth and config files
		tokens := auth.LoadFromFile()
		rootDir, err := git.ToplevelDir()
		if err != nil {
			log.Fatal(err)
		}

		servicePath, serviceName := config.LoadStaticServiceFromFile(rootDir)
		if servicePath == "" {
			log.Fatal("Static app configuration not found in peek.yml")
		}

		assetPath := filepath.Join(rootDir, servicePath)

		// Read info out of local git repo
		branch, err := git.CurrentBranch()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		sha, err := git.CurrentSha()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		originSha, err := git.ShaForRemoteBranch(branch)
		if err != nil {
			log.Fatalf("Error reading remote branch: %v", err)
		}

		if originSha != sha {
			log.Fatal("Error: origin branch sha does not match local branch sha. You may need to push your changes.")
		}

		remotes, err := context.GetRemotes()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		originRemote, err := remotes.FindByName("origin")
		if err != nil {
			log.Fatal("Error: no remote named 'origin' found")
		}

		org := originRemote.Owner
		repo := originRemote.Repo

		// Archive web asset directory
		files, err := ioutil.ReadDir(assetPath)
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}

		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, filepath.Join(assetPath, file.Name()))
		}

		checksum, err := dirChecksum(assetPath)
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}

		// Send ping
		spinnerDone := spinner.StartSpinning("Packaging and Uploading")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("artifacts", "artifacts.tar.gz")
		if err != nil {
			log.Fatal(err)
		}

		tmpDir := os.TempDir()
		tmpFilename := filepath.Join(tmpDir, "artifacts.tar.gz")

		err = archiver.Archive(fileNames, tmpFilename)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(tmpFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = io.Copy(part, file)
		os.Remove(tmpFilename)

		writer.WriteField("app", serviceName)
		writer.WriteField("service", "cli")
		writer.WriteField("org", org)
		writer.WriteField("repo", repo)
		writer.WriteField("sha", sha)
		writer.WriteField("branch", branch)
		writer.WriteField("checksum", checksum)
		if err = writer.Close(); err != nil {
			log.Fatal(err)
		}

		var pingURL string
		if devFlag {
			pingURL = "https://api.dev.featurepeek.com/api/v1/peek"
		} else {
			pingURL = "https://api.featurepeek.com/api/v1/peek"
		}
		request, err := http.NewRequest("POST", pingURL, body)
		request.Header.Add("authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))
		request.Header.Add("X-FEATUREPEEK-CLIENT", Version)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		if debugFlag {
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
		// stop spinner
		spinnerDone <- true

		if debugFlag {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Header)
		}

		if response.StatusCode != http.StatusOK &&
			response.StatusCode != http.StatusCreated {
			fmt.Printf("\n%s\n", body)
			log.Fatalf("\nUpload Failed with status %d", response.StatusCode)

		}

		if response.StatusCode == http.StatusOK {
			fmt.Println(body)
		} else {
			fmt.Println("\n\nAssets uploaded successfully!\nVisit your new feature environment here:")
			fmt.Println(body)
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
		viper.SetConfigName("peek")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func dirChecksum(dir string) (string, error) {
	hashdump := md5.New()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		io.WriteString(hashdump, string(data))
		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hashdump.Sum(nil)), nil
}
