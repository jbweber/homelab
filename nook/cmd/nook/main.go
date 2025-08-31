package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jbweber/homelab/nook/internal/datastore"
)

func main() {
	// Initialize the datastore
	dbPath := "nook.db"
	ds, err := datastore.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer ds.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Nook web service is running!")
	})

	fmt.Println("Starting Nook web service on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
