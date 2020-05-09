// Package cmd defines the primary functionality of the CLI
package cmd

import (
	"fmt"
	"log"
	"peek/peekconfig"
	"strings"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a peek.yml config",
	Long: `Initialize a new peek.yml config.

This command will run the user through a wizard to determine what settings to use
when creating the local peek.yml config file. Once the questions are answered, the new
config file will be created in the local directory and it should be immedeately commited
to git and pushed to remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		var pathInput string
		var spaInput string

		fmt.Println("Initializing peek.yml config for static app...")
		fmt.Println("\nEnter path of statically built assets, relative to repo root:")
		for pathInput == "" {
			fmt.Print("--> ")
			fmt.Scanln(&pathInput)
		}

		fmt.Println("Is your project a Single Page Application? (y/n)")
		for spaInput == "" {
			fmt.Print("--> ")
			fmt.Scanln(&spaInput)
		}

		peekConfig := peekconfig.Config{
			Version: 2,
			Main: peekconfig.Service{
				Type: "static",
				Path: pathInput,
				Spa:  strings.ToLower(spaInput)[0] == 'y',
			},
		}
		if err := peekConfig.Save(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("\npeek.yml saved!")
		fmt.Println("\nMake sure to commit and push this file before deploying a preview")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
