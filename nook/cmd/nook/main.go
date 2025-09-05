//go:build !test

// Code coverage for main is ignored for now. TODO: Add integration tests for main entrypoint.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jbweber/homelab/nook/internal/api"
	"github.com/jbweber/homelab/nook/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "nook",
		Short: "Nook is a metadata service for cloud-init",
		Long:  `Nook provides metadata endpoints for cloud-init and allows management of machines, networks, and SSH keys.`,
	}

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the nook web service",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.NewConfig()
			cfg.DBPath, _ = cmd.Flags().GetString("db-path")
			cfg.Port, _ = cmd.Flags().GetString("port")
			runServer(cfg)
		},
	}
	serverCmd.Flags().String("db-path", "~/nook/data/nook.db", "Path to the database file")
	serverCmd.Flags().String("port", "8080", "Port to run the server on")

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add resources to the nook service",
	}

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete resources from the nook service",
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
	if err := addMachineCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	if err := addMachineCmd.MarkFlagRequired("hostname"); err != nil {
		log.Fatal(err)
	}
	if err := addMachineCmd.MarkFlagRequired("ipv4"); err != nil {
		log.Fatal(err)
	}

	var addNetworkCmd = &cobra.Command{
		Use:   "network",
		Short: "Add a network",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			addNetwork(name)
		},
	}
	addNetworkCmd.Flags().String("name", "", "Network name (required)")
	if err := addNetworkCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}

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
	if err := addSSHKeyCmd.MarkFlagRequired("machine-id"); err != nil {
		log.Fatal(err)
	}
	if err := addSSHKeyCmd.MarkFlagRequired("key-text"); err != nil {
		log.Fatal(err)
	}

	var deleteMachineCmd = &cobra.Command{
		Use:   "machine",
		Short: "Delete a machine",
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetInt64("id")
			deleteMachine(id)
		},
	}
	deleteMachineCmd.Flags().Int64("id", 0, "Machine ID (required)")
	if err := deleteMachineCmd.MarkFlagRequired("id"); err != nil {
		log.Fatal(err)
	}

	var deleteNetworkCmd = &cobra.Command{
		Use:   "network",
		Short: "Delete a network",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			deleteNetwork(name)
		},
	}
	deleteNetworkCmd.Flags().String("name", "", "Network name (required)")
	if err := deleteNetworkCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}

	var deleteSSHKeyCmd = &cobra.Command{
		Use:   "ssh-key",
		Short: "Delete an SSH key",
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetInt64("id")
			deleteSSHKey(id)
		},
	}
	deleteSSHKeyCmd.Flags().Int64("id", 0, "SSH key ID (required)")
	if err := deleteSSHKeyCmd.MarkFlagRequired("id"); err != nil {
		log.Fatal(err)
	}

	addCmd.AddCommand(addMachineCmd)
	addCmd.AddCommand(addNetworkCmd)
	addCmd.AddCommand(addSSHKeyCmd)
	deleteCmd.AddCommand(deleteMachineCmd)
	deleteCmd.AddCommand(deleteNetworkCmd)
	deleteCmd.AddCommand(deleteSSHKeyCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cfg *config.Config) {
	// Initialize database
	db, err := cfg.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

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

	fmt.Printf("Starting Nook web service on :%s...\n", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, r)
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to add SSH key: %s", resp.Status)
	}
	fmt.Println("SSH key added successfully")
}

func deleteMachine(id int64) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/api/v0/machines/%d", id), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to delete machine: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Failed to delete machine: %s", resp.Status)
	}
	fmt.Println("Machine deleted successfully")
}

func deleteNetwork(name string) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/api/v0/networks/%s", name), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to delete network: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Failed to delete network: %s", resp.Status)
	}
	fmt.Println("Network deleted successfully")
}

func deleteSSHKey(id int64) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/api/v0/ssh-keys/%d", id), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to delete SSH key: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Failed to delete SSH key: %s", resp.Status)
	}
	fmt.Println("SSH key deleted successfully")
}

