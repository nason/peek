/*
Copyright © 2020 Landon Spear <phyujin@gmail.com>

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

package cmd

import (
	"fmt"
	"peek/auth"

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
		auth.RemoveFile()
		fmt.Print("done\n")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
