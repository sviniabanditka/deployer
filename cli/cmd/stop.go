package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the application",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		if err := apiClient.StopApp(appID); err != nil {
			fmt.Printf("Failed to stop app: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("App stopped.")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
