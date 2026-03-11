package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var followLogs bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View application logs",
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth()
		appID := requireAppID()

		if followLogs {
			streamLogs(appID)
		} else {
			fetchLogs(appID)
		}
	},
}

func fetchLogs(appID string) {
	logs, err := apiClient.GetLogs(appID)
	if err != nil {
		fmt.Printf("Failed to fetch logs: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(logs)
}

func streamLogs(appID string) {
	wsURL := apiClient.LogsWebSocketURL(appID)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Printf("Failed to connect to log stream: %v\n", err)
		fmt.Println("Falling back to non-streaming logs...")
		fetchLogs(appID)
		return
	}
	defer conn.Close()

	// Handle interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			fmt.Print(string(message))
		}
	}()

	select {
	case <-done:
	case <-interrupt:
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")
	rootCmd.AddCommand(logsCmd)
}
