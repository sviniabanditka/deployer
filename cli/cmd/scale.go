package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scaleCmd = &cobra.Command{
	Use:   "scale PROCESS=COUNT",
	Short: "Scale application processes",
	Example: "  deployer scale web=3",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scaling is not yet available.")
	},
}

func init() {
	rootCmd.AddCommand(scaleCmd)
}
