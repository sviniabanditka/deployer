package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		if err := apiClient.StartApp(appID); err != nil {
			fmt.Printf("Failed to start app: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("App started.")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
