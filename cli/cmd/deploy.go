package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/deployer/cli/internal/archive"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current directory to your app",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("Uploading... ")
		zipPath, err := archive.ZipDirectory(dir, archive.DefaultExcludes)
		if err != nil {
			fmt.Printf("\nFailed to create archive: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(zipPath)

		resp, err := apiClient.Deploy(appID, zipPath)
		if err != nil {
			fmt.Printf("\nDeploy failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("done")

		if resp.ID != "" {
			fmt.Print("Building... ")
			// Poll deployment status
			for i := 0; i < 120; i++ {
				status, err := apiClient.GetDeploymentStatus(appID, resp.ID)
				if err != nil {
					// If endpoint doesn't exist, just show success
					break
				}

				switch status.Status {
				case "running", "live", "succeeded", "success":
					fmt.Println("done")
					fmt.Printf("Deployed successfully! Deployment ID: %s\n", resp.ID)
					return
				case "failed", "error":
					fmt.Println("failed")
					if status.Logs != "" {
						fmt.Println("\nBuild logs:")
						fmt.Println(status.Logs)
					}
					fmt.Println("Deployment failed.")
					os.Exit(1)
				}

				time.Sleep(2 * time.Second)
			}
		}

		fmt.Printf("Deployment started. ID: %s\n", resp.ID)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
