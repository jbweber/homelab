package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jbweber/homelab/nook/internal/api"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

func main() {
	// Initialize datastore
	ds, err := datastore.New("nook.db")
	if err != nil {
		log.Fatalf("Failed to initialize datastore: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register API routes
	api := api.NewAPI(ds)
	api.RegisterRoutes(r)

	// Health check endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Nook web service is running!")
	})

	fmt.Println("Starting Nook web service on :8080...")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
