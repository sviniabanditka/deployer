package cmd

import (
	"fmt"
	"os"

	"github.com/deployer/cli/internal/client"
	"github.com/deployer/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	apiURL     string
	appConfig  *config.Config
	apiClient  *client.Client
)

var rootCmd = &cobra.Command{
	Use:   "deployer",
	Short: "Deploy applications with ease",
	Long:  "Deployer CLI - Build, deploy, and manage your applications.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load config from file
		cfg, err := config.Load(cfgFile)
		if err != nil {
			cfg = &config.Config{}
		}
		appConfig = cfg

		// Determine API URL: flag > config file > default
		if !cmd.Flags().Changed("api-url") && cfg.APIUrl != "" {
			apiURL = cfg.APIUrl
		}

		apiClient = client.New(apiURL, cfg.AccessToken, cfg.RefreshToken, cfgFile)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultPath(), "config file")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:3000/api/v1", "API server URL")

	viper.SetDefault("api_url", "http://localhost:3000/api/v1")
}

// requireAuth checks that the user is logged in and exits with an error if not.
func requireAuth() {
	if appConfig == nil || appConfig.AccessToken == "" {
		fmt.Println("Error: You are not logged in. Run 'deployer login' first.")
		os.Exit(1)
	}
}

// requireAppID reads the app ID from .deployer.json in the current directory.
func requireAppID() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error: Could not determine current directory.")
		os.Exit(1)
	}
	appID, err := config.LoadAppConfig(dir)
	if err != nil || appID == "" {
		fmt.Println("Error: No app configured in this directory. Run 'deployer init' first.")
		os.Exit(1)
	}
	return appID
}

// confirmPrompt asks the user for yes/no confirmation.
func confirmPrompt(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes"
}
