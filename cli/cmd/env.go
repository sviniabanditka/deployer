package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		app, err := apiClient.GetApp(appID)
		if err != nil {
			fmt.Printf("Failed to get app: %v\n", err)
			os.Exit(1)
		}

		if len(app.EnvVars) == 0 {
			fmt.Println("No environment variables set.")
			return
		}

		for key, value := range app.EnvVars {
			fmt.Printf("%s=%s\n", key, value)
		}
	},
}

var envSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE [KEY2=VALUE2 ...]",
	Short: "Set environment variables",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		// Get current env vars first
		app, err := apiClient.GetApp(appID)
		if err != nil {
			fmt.Printf("Failed to get app: %v\n", err)
			os.Exit(1)
		}

		envVars := app.EnvVars
		if envVars == nil {
			envVars = make(map[string]string)
		}

		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				fmt.Printf("Invalid format: %s (expected KEY=VALUE)\n", arg)
				os.Exit(1)
			}
			envVars[parts[0]] = parts[1]
		}

		if err := apiClient.UpdateEnvVars(appID, envVars); err != nil {
			fmt.Printf("Failed to update env vars: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Environment variables updated. Set %d variable(s).\n", len(args))
	},
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset KEY [KEY2 ...]",
	Short: "Remove environment variables",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		// Get current env vars
		app, err := apiClient.GetApp(appID)
		if err != nil {
			fmt.Printf("Failed to get app: %v\n", err)
			os.Exit(1)
		}

		envVars := app.EnvVars
		if envVars == nil {
			envVars = make(map[string]string)
		}

		for _, key := range args {
			delete(envVars, key)
		}

		if err := apiClient.UpdateEnvVars(appID, envVars); err != nil {
			fmt.Printf("Failed to update env vars: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Removed %d variable(s).\n", len(args))
	},
}

func init() {
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	rootCmd.AddCommand(envCmd)
}
