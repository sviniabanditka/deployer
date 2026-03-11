package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage your applications",
	Run: func(cmd *cobra.Command, args []string) {
		// Default to listing apps
		appsListCmd.Run(cmd, args)
	},
}

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		apps, err := apiClient.ListApps()
		if err != nil {
			fmt.Printf("Failed to list apps: %v\n", err)
			os.Exit(1)
		}

		if len(apps) == 0 {
			fmt.Println("No apps found. Run 'deployer init' to create one.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSLUG\tSTATUS")
		fmt.Fprintln(w, "--\t----\t----\t------")
		for _, app := range apps {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", app.ID, app.Name, app.Slug, app.Status)
		}
		w.Flush()
	},
}

var appsInfoCmd = &cobra.Command{
	Use:   "info [app-id]",
	Short: "Show app details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		var appID string
		if len(args) > 0 {
			appID = args[0]
		} else {
			appID = requireAppID()
		}

		app, err := apiClient.GetApp(appID)
		if err != nil {
			fmt.Printf("Failed to get app: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("ID:        %s\n", app.ID)
		fmt.Printf("Name:      %s\n", app.Name)
		fmt.Printf("Slug:      %s\n", app.Slug)
		fmt.Printf("Status:    %s\n", app.Status)
		fmt.Printf("URL:       https://%s.deployer.dev\n", app.Slug)
		if app.CreatedAt != "" {
			fmt.Printf("Created:   %s\n", app.CreatedAt)
		}
		if len(app.EnvVars) > 0 {
			fmt.Printf("Env Vars:  %d configured\n", len(app.EnvVars))
		}
	},
}

var appsDeleteCmd = &cobra.Command{
	Use:   "delete [app-id]",
	Short: "Delete an application",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		var appID string
		if len(args) > 0 {
			appID = args[0]
		} else {
			appID = requireAppID()
		}

		app, err := apiClient.GetApp(appID)
		if err != nil {
			fmt.Printf("Failed to get app: %v\n", err)
			os.Exit(1)
		}

		if !confirmPrompt(fmt.Sprintf("Are you sure you want to delete '%s'? This cannot be undone", app.Name)) {
			fmt.Println("Aborted.")
			return
		}

		if err := apiClient.DeleteApp(appID); err != nil {
			fmt.Printf("Failed to delete app: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("App '%s' deleted.\n", app.Name)
	},
}

func init() {
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsInfoCmd)
	appsCmd.AddCommand(appsDeleteCmd)
	rootCmd.AddCommand(appsCmd)
}
