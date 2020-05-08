// Package cmd defines the primary functionality of the CLI
package cmd

import (
	"fmt"
	"log"
	"peek/config"

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
		fmt.Print("Initializing peek.yml config for static app...\n\n")

		var pathInput string
		fmt.Printf("Enter path of statically built assets, relative to repo root:\n--> ")
		fmt.Scanln(&pathInput)

		var spaInput string
		fmt.Printf("Is your project a Single Page Application? (y/n)\n--> ")
		fmt.Scanln(&spaInput)

		yesResponses := []string{"y", "Y", "yes", "Yes", "YES", "true", "t", "1"}

		peekConfig := config.Config{
			Version: 2,
			Main: config.Service{
				Type: "static",
				Path: pathInput,
				Spa:  containsString(yesResponses, spaInput),
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

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}
