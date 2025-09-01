//go:build !test

// Code coverage for main is ignored for now. TODO: Add integration tests for main entrypoint.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jbweber/homelab/nook/internal/api"
	"github.com/jbweber/homelab/nook/internal/migrations"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "nook",
		Short: "Nook is a metadata service for cloud-init",
		Long:  `Nook provides metadata endpoints for cloud-init and allows management of machines, networks, and SSH keys.`,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add resources to the nook service",
	}

	var addMachineCmd = &cobra.Command{
		Use:   "machine",
		Short: "Add a machine",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			hostname, _ := cmd.Flags().GetString("hostname")
			ipv4, _ := cmd.Flags().GetString("ipv4")
			addMachine(name, hostname, ipv4)
		},
	}
	addMachineCmd.Flags().String("name", "", "Machine name (required)")
	addMachineCmd.Flags().String("hostname", "", "Machine hostname (required)")
	addMachineCmd.Flags().String("ipv4", "", "Machine IPv4 address (required)")
	addMachineCmd.MarkFlagRequired("name")
	addMachineCmd.MarkFlagRequired("hostname")
	addMachineCmd.MarkFlagRequired("ipv4")

	var addNetworkCmd = &cobra.Command{
		Use:   "network",
		Short: "Add a network",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			addNetwork(name)
		},
	}
	addNetworkCmd.Flags().String("name", "", "Network name (required)")
	addNetworkCmd.MarkFlagRequired("name")

	var addSSHKeyCmd = &cobra.Command{
		Use:   "ssh-key",
		Short: "Add an SSH key",
		Run: func(cmd *cobra.Command, args []string) {
			machineID, _ := cmd.Flags().GetInt64("machine-id")
			keyText, _ := cmd.Flags().GetString("key-text")
			addSSHKey(machineID, keyText)
		},
	}
	addSSHKeyCmd.Flags().Int64("machine-id", 0, "Machine ID (required)")
	addSSHKeyCmd.Flags().String("key-text", "", "SSH key text (required)")
	addSSHKeyCmd.MarkFlagRequired("machine-id")
	addSSHKeyCmd.MarkFlagRequired("key-text")

	addCmd.AddCommand(addMachineCmd)
	addCmd.AddCommand(addNetworkCmd)
	addCmd.AddCommand(addSSHKeyCmd)
	rootCmd.AddCommand(addCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer() {
	// Initialize database
	db, err := initializeDatabase("nook.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register API routes
	api := api.NewAPI(db)
	api.RegisterRoutes(r)

	// Health check endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "Nook web service is running!"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	fmt.Println("Starting Nook web service on :8080...")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func addMachine(name, hostname, ipv4 string) {
	req := map[string]string{
		"name":     name,
		"hostname": hostname,
		"ipv4":     ipv4,
	}
	data, _ := json.Marshal(req)
	resp, err := http.Post("http://localhost:8080/api/v0/machines", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Failed to add machine: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to add machine: %s", resp.Status)
	}
	fmt.Println("Machine added successfully")
}

func addNetwork(name string) {
	req := map[string]string{
		"name": name,
	}
	data, _ := json.Marshal(req)
	resp, err := http.Post("http://localhost:8080/api/v0/networks", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Failed to add network: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to add network: %s", resp.Status)
	}
	fmt.Println("Network added successfully")
}

func addSSHKey(machineID int64, keyText string) {
	req := map[string]interface{}{
		"machine_id": machineID,
		"key_text":   keyText,
	}
	data, _ := json.Marshal(req)
	resp, err := http.Post("http://localhost:8080/api/v0/ssh-keys", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Failed to add SSH key: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to add SSH key: %s", resp.Status)
	}
	fmt.Println("SSH key added successfully")
}

// initializeDatabase creates a new SQLite database and runs migrations.
func initializeDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

// runMigrations runs all database migrations
func runMigrations(db *sql.DB) error {
	migrator := migrations.NewMigrator(db)

	// Add all migrations
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	// Run migrations
	if err := migrator.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
