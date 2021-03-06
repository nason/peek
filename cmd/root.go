// Package cmd defines the primary functionality of the CLI
package cmd

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"peek/config"
	"peek/context"
	"peek/git"
	"peek/peekconfig"
	"peek/spinner"
	"runtime/debug"
	"strings"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"
)

// Version is dynamically set by the toolchain.
var Version = "DEV"

var versionOutput = ""

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	Version = strings.TrimPrefix(Version, "v")
	rootCmd.Version = Version
	rootCmd.AddCommand(versionCmd)
	versionOutput = fmt.Sprintf("peek version %s", rootCmd.Version)

	rootCmd.PersistentFlags().StringVar(&targetDir, "dir", "", "target directory to launch from")
	rootCmd.PersistentFlags().StringVar(&targetService, "service", "", "select specific front-end service to launch")
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

const peekCommandLongDesc = `peek is a command-line tool for interacting with FeaturePeek deployments.

The FeaturePeek CLI enables front-end developers, designers, and product owners to interact with and review your changes
by creating deployment previews of the front-end code you are working on.

The goal of this tool is to create deployment previews for any commit, all without needing a CI system.

To get started, simply run ` + "`peek login`" + `to authenticate locally and/or create an account.
Run ` + "`peek init`" + ` and enter your build directory to set up your config.
Make sure your code pushed to your remote and run your build step.
Then run ` + "`peek`" + ` to launch your FeaturePeek deployment.`

const errorMessageCI = `CI environment detected.
The peek CLI is meant to be used interactively at the command line.

To use FeaturePeek in your CI pipeline, sign up for FeaturePeek Teams at https://featurepeek.com/product/teams`

var debugFlag bool
var devFlag bool
var targetDir string
var targetService string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "peek",
	Short: "FeaturePeek Command-line Tool",
	Long:  peekCommandLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		// check if running in CI
		if os.Getenv("CI") != "" {
			log.Fatalln(errorMessageCI)
		}

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
		localConfig, err := config.LoadConfig(devFlag)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal("No credentials found. Run `peek login` to login with your FeaturePeek account.")
			} else {
				log.Fatalf("Error reading config file: %v", err)
			}
		}

		tokens := localConfig.Auth
		if tokens == nil {
			log.Fatal("No credentials found. Run `peek login` to login with your FeaturePeek account.")
		}

		rootDir, err := git.ToplevelDir()
		if err != nil {
			log.Fatal(err)
		}

		peekConfigFilename := filepath.Join(rootDir, "peek.yml")
		service, err := peekconfig.LoadStaticServiceFromFile(peekConfigFilename, targetService)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal("No peek.yml config found.\n\nRun `peek init` to create one!")
			} else {
				log.Fatalf("Cannot read peek.yml config: %v.", err)
			}
		}
		if service == nil {
			log.Fatal("Static app configuration not found in peek.yml")
		}

		assetPath := filepath.Join(rootDir, service.Path)

		// Read info out of local git repo
		branch, err := git.CurrentBranch()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		// make sure peek config exists on remote
		err = git.CheckForFileOnRemoteBranch(branch, "peek.yml")
		if err != nil {
			log.Fatal("peek.yml config not found on remote.\nMake sure to push your config file")
		}

		// warn for uncommited files
		uncommitedFiles, err := git.GetUncommitedFiles()
		if err != nil {
			log.Fatal("Error reading git status")
		}
		if len(uncommitedFiles) > 0 {
			showUncommitedChangesWarning()
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
			log.Fatal("Error: local commit HEAD does not match origin.\nYou may still need to push your changes.")
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
		uploadSpinner := spinner.New("Packaging and Uploading")
		go uploadSpinner.Start()

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

		writer.WriteField("app", service.Name)
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
		defer response.Body.Close()

		resBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		uploadSpinner.Stop()

		if debugFlag {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Header)
		}

		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
			var errorResponse struct {
				Errors []string
			}
			if err = json.Unmarshal(resBody, &errorResponse); err != nil {
				if len(resBody) == 0 {
					log.Fatalf("Upload failed with status %d", response.StatusCode)
				}
				log.Fatalf("Upload Failed with status %d - %s", response.StatusCode, string(resBody))
			}
			log.Fatalf("Upload Failed with status %d - %s", response.StatusCode, errorResponse.Errors)
		}

		if response.StatusCode == http.StatusOK {
			fmt.Println(string(resBody))
		} else {
			fmt.Printf("Assets uploaded successfully! %s\n", randomEmoji())
			fmt.Printf("Visit your deployment preview here: %s\n", string(resBody))
		}
	},
}

func randomEmoji() string {
	emoji := []string{
		"🧡", "💛", "💚", "💙", "💜", "💖", "🆒", "🎉", "✨", "😄", "🚀", "😍", "😁", "💪", "😀", "🥳", "😎", "🤩", "🙌", "✌️", "🤘", "👌", "🤙", "👏", "🌈", "⭐️", "🌟", "💫", "⚡️", "🌶", "🍉", "🍕", "🍦", "🍭", "🍪", "🍻", "🏆", "🎖", "🏅", "🥇", "🏄‍♂️", "⛳️", "🎯", "🎇", "🌠", "🖖", "💯", "🎊", "📈", "🔮", "💎", "🔥", "🌻", "👩‍🎤", "👨‍🎤",
	}
	rand.Seed(time.Now().Unix())
	return emoji[rand.Intn(len(emoji))]
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

func showUncommitedChangesWarning() {
	var input string

	fmt.Println("You have local uncommited changes.")
	fmt.Println("\nIf they effect your deployment,\nthey will not be visible on your remote until you commit and push them.")
	fmt.Println("\nWould you like to continue anyway? (y/n)")

	for input == "" {
		fmt.Print("--> ")
		fmt.Scanln(&input)
	}

	if strings.ToLower(input)[0] != 'y' {
		os.Exit(0)
	}
}
