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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
		fmt.Println("Saved peek.yml config")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
