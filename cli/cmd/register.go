package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/deployer/cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Create a new Deployer account",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			fmt.Println("Error: Name is required.")
			os.Exit(1)
		}

		fmt.Print("Email: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if email == "" {
			fmt.Println("Error: Email is required.")
			os.Exit(1)
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}
		password := string(passwordBytes)
		if password == "" {
			fmt.Println("Error: Password is required.")
			os.Exit(1)
		}

		resp, err := apiClient.Register(name, email, password)
		if err != nil {
			fmt.Printf("Registration failed: %v\n", err)
			os.Exit(1)
		}

		appConfig.AccessToken = resp.AccessToken
		appConfig.RefreshToken = resp.RefreshToken
		appConfig.Email = resp.User.Email
		appConfig.APIUrl = apiURL

		if err := config.Save(cfgFile, appConfig); err != nil {
			fmt.Printf("Warning: Could not save config: %v\n", err)
		}

		fmt.Printf("Account created. Logged in successfully as %s\n", resp.User.Email)
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
