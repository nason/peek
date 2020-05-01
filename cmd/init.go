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

		var input string
		fmt.Printf("Enter path of statically built assets, relative to repo root:\n--> ")
		fmt.Scanln(&input)

		peekConfig := config.Config{
			Version: 2,
			Main: config.Service{
				Type: "static",
				Path: input,
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
