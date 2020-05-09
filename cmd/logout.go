package cmd

import (
	"fmt"
	"peek/config"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of your FeaturePeek Account",
	Long: `Logout of your FeaturePeek Account.

	This command erases your FeaturePeek account credentials
	from your computer. You will need to log back in to launch previews
	using the command-line tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Logging out... ")
		config.RemoveAuthFromConfigFile(devFlag)
		fmt.Print("done\n")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
