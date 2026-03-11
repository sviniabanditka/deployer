package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deployer/cli/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new app in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Check if already initialized
		if existingID, err := config.LoadAppConfig(dir); err == nil && existingID != "" {
			fmt.Printf("This directory is already linked to app %s.\n", existingID)
			if !confirmPrompt("Do you want to create a new app instead?") {
				return
			}
		}

		defaultName := filepath.Base(dir)
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("App name [%s]: ", defaultName)
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = defaultName
		}

		app, err := apiClient.CreateApp(name)
		if err != nil {
			fmt.Printf("Failed to create app: %v\n", err)
			os.Exit(1)
		}

		if err := config.SaveAppConfig(dir, app.ID); err != nil {
			fmt.Printf("Error saving app config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("App '%s' created. Subdomain: %s.deployer.dev\n", app.Name, app.Slug)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
