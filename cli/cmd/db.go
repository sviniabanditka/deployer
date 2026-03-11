package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var validEngines = []string{"postgres", "mysql", "mongodb", "redis"}

func isValidEngine(engine string) bool {
	for _, e := range validEngines {
		if e == engine {
			return true
		}
	}
	return false
}

func maskURL(url string) string {
	if url == "" {
		return ""
	}
	// Mask password in connection URLs like postgres://user:pass@host/db
	if idx := strings.Index(url, "://"); idx != -1 {
		rest := url[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			if colonIdx := strings.Index(rest[:atIdx], ":"); colonIdx != -1 {
				return url[:idx+3] + rest[:colonIdx+1] + "****" + rest[atIdx:]
			}
		}
	}
	return url
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage databases",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// --- db create ---

var dbCreateName string
var dbCreateAppID string

var dbCreateCmd = &cobra.Command{
	Use:   "create <engine>",
	Short: "Create a new database",
	Long:  "Create a new managed database. Supported engines: postgres, mysql, mongodb, redis.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		engine := args[0]
		if !isValidEngine(engine) {
			fmt.Printf("Invalid engine '%s'. Supported engines: %s\n", engine, strings.Join(validEngines, ", "))
			os.Exit(1)
		}

		name := dbCreateName
		if name == "" {
			name = fmt.Sprintf("my-%s-1", engine)
		}

		db, err := apiClient.CreateDatabase(name, engine, "", dbCreateAppID)
		if err != nil {
			fmt.Printf("Failed to create database: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Database created successfully!\n\n")
		fmt.Printf("ID:             %s\n", db.ID)
		fmt.Printf("Name:           %s\n", db.Name)
		fmt.Printf("Engine:         %s\n", db.Engine)
		fmt.Printf("Status:         %s\n", db.Status)
		if db.ConnectionURL != "" {
			fmt.Printf("Connection URL: %s\n", db.ConnectionURL)
		}
		if db.Host != "" {
			fmt.Printf("Host:           %s\n", db.Host)
		}
		if db.Port != 0 {
			fmt.Printf("Port:           %d\n", db.Port)
		}
		if db.Username != "" {
			fmt.Printf("Username:       %s\n", db.Username)
		}
		if db.Password != "" {
			fmt.Printf("Password:       %s\n", db.Password)
		}
		if db.DatabaseName != "" {
			fmt.Printf("Database:       %s\n", db.DatabaseName)
		}
	},
}

// --- db list ---

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all databases",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		dbs, err := apiClient.ListDatabases()
		if err != nil {
			fmt.Printf("Failed to list databases: %v\n", err)
			os.Exit(1)
		}

		if len(dbs) == 0 {
			fmt.Println("No databases found. Run 'deployer db create <engine>' to create one.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tENGINE\tSTATUS\tCONNECTION URL")
		fmt.Fprintln(w, "--\t----\t------\t------\t--------------")
		for _, db := range dbs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", db.ID, db.Name, db.Engine, db.Status, maskURL(db.ConnectionURL))
		}
		w.Flush()
	},
}

// --- db info ---

var dbInfoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "Show database details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		db, err := apiClient.GetDatabase(args[0])
		if err != nil {
			fmt.Printf("Failed to get database: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("ID:             %s\n", db.ID)
		fmt.Printf("Name:           %s\n", db.Name)
		fmt.Printf("Engine:         %s\n", db.Engine)
		if db.Version != "" {
			fmt.Printf("Version:        %s\n", db.Version)
		}
		fmt.Printf("Status:         %s\n", db.Status)
		if db.AppID != "" {
			fmt.Printf("Linked App:     %s\n", db.AppID)
		}
		if db.ConnectionURL != "" {
			fmt.Printf("Connection URL: %s\n", db.ConnectionURL)
		}
		if db.Host != "" {
			fmt.Printf("Host:           %s\n", db.Host)
		}
		if db.Port != 0 {
			fmt.Printf("Port:           %d\n", db.Port)
		}
		if db.Username != "" {
			fmt.Printf("Username:       %s\n", db.Username)
		}
		if db.Password != "" {
			fmt.Printf("Password:       %s\n", db.Password)
		}
		if db.DatabaseName != "" {
			fmt.Printf("Database:       %s\n", db.DatabaseName)
		}
		if db.CreatedAt != "" {
			fmt.Printf("Created:        %s\n", db.CreatedAt)
		}
	},
}

// --- db delete ---

var dbDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		db, err := apiClient.GetDatabase(args[0])
		if err != nil {
			fmt.Printf("Failed to get database: %v\n", err)
			os.Exit(1)
		}

		if !confirmPrompt(fmt.Sprintf("Are you sure you want to delete database '%s'? This will destroy all data.", db.Name)) {
			fmt.Println("Aborted.")
			return
		}

		if err := apiClient.DeleteDatabase(args[0]); err != nil {
			fmt.Printf("Failed to delete database: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Database '%s' deleted.\n", db.Name)
	},
}

// --- db stop ---

var dbStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		if err := apiClient.StopDatabase(args[0]); err != nil {
			fmt.Printf("Failed to stop database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database stopped.")
	},
}

// --- db start ---

var dbStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		if err := apiClient.StartDatabase(args[0]); err != nil {
			fmt.Printf("Failed to start database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database started.")
	},
}

// --- db link ---

var dbLinkCmd = &cobra.Command{
	Use:   "link <db-id> <app-id>",
	Short: "Link a database to an application",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		if err := apiClient.LinkDatabase(args[0], args[1]); err != nil {
			fmt.Printf("Failed to link database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database linked. DATABASE_URL set on app.")
	},
}

// --- db unlink ---

var dbUnlinkCmd = &cobra.Command{
	Use:   "unlink <db-id>",
	Short: "Unlink a database from its application",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		if err := apiClient.UnlinkDatabase(args[0]); err != nil {
			fmt.Printf("Failed to unlink database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database unlinked.")
	},
}

// --- db backup ---

var dbBackupCmd = &cobra.Command{
	Use:   "backup <id>",
	Short: "Create a backup of a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		backup, err := apiClient.CreateBackup(args[0])
		if err != nil {
			fmt.Printf("Failed to create backup: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Backup created successfully!\n\n")
		fmt.Printf("ID:      %s\n", backup.ID)
		fmt.Printf("Status:  %s\n", backup.Status)
		if backup.Size != "" {
			fmt.Printf("Size:    %s\n", backup.Size)
		}
		if backup.CreatedAt != "" {
			fmt.Printf("Date:    %s\n", backup.CreatedAt)
		}
	},
}

// --- db backups ---

var dbBackupsCmd = &cobra.Command{
	Use:   "backups <id>",
	Short: "List backups for a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		backups, err := apiClient.ListBackups(args[0])
		if err != nil {
			fmt.Printf("Failed to list backups: %v\n", err)
			os.Exit(1)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tDATE\tSIZE\tSTATUS")
		fmt.Fprintln(w, "--\t----\t----\t------")
		for _, b := range backups {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.ID, b.CreatedAt, b.Size, b.Status)
		}
		w.Flush()
	},
}

// --- db restore ---

var dbRestoreCmd = &cobra.Command{
	Use:   "restore <db-id> <backup-id>",
	Short: "Restore a database from a backup",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()

		if !confirmPrompt("Are you sure? This will overwrite current data.") {
			fmt.Println("Aborted.")
			return
		}

		if err := apiClient.RestoreBackup(args[0], args[1]); err != nil {
			fmt.Printf("Failed to restore backup: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database restore initiated.")
	},
}

func init() {
	dbCreateCmd.Flags().StringVar(&dbCreateName, "name", "", "database name (auto-generated if not set)")
	dbCreateCmd.Flags().StringVar(&dbCreateAppID, "app", "", "link to an app by ID")

	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbInfoCmd)
	dbCmd.AddCommand(dbDeleteCmd)
	dbCmd.AddCommand(dbStopCmd)
	dbCmd.AddCommand(dbStartCmd)
	dbCmd.AddCommand(dbLinkCmd)
	dbCmd.AddCommand(dbUnlinkCmd)
	dbCmd.AddCommand(dbBackupCmd)
	dbCmd.AddCommand(dbBackupsCmd)
	dbCmd.AddCommand(dbRestoreCmd)
	rootCmd.AddCommand(dbCmd)
}
