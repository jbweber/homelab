//go:build !test

// Code coverage for main is ignored for now. TODO: Add integration tests for main entrypoint.
package main

import (
	"log"

	"github.com/jbweber/homelab/nook/cmd/nook/internal/client"
	"github.com/jbweber/homelab/nook/cmd/nook/internal/commands"
	"github.com/jbweber/homelab/nook/cmd/nook/internal/database"
	"github.com/jbweber/homelab/nook/cmd/nook/internal/server"
	"github.com/spf13/cobra"
)

func main() {
	// Initialize client for API operations
	clientConfig := client.NewConfig("http://localhost:8080")
	commandsConfig := commands.NewConfig(clientConfig)

	rootCmd := commandsConfig.NewRootCmd()

	// Override the server command to handle it specially
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the nook web service",
		Run: func(cmd *cobra.Command, args []string) {
			dbPath, _ := cmd.Flags().GetString("db-path")
			port, _ := cmd.Flags().GetString("port")
			runServer(dbPath, port)
		},
	}
	serverCmd.Flags().String("db-path", "~/nook/data/nook.db", "Path to the database file")
	serverCmd.Flags().String("port", "8080", "Port to run the server on")

	// Replace the server command
	for i, cmd := range rootCmd.Commands() {
		if cmd.Use == "server" {
			rootCmd.Commands()[i] = serverCmd
			break
		}
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(dbPath, port string) {
	// Initialize database
	dbConfig := database.NewConfig(dbPath)
	db, err := dbConfig.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start server
	serverConfig := server.NewConfig(port, db)
	if err := serverConfig.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
