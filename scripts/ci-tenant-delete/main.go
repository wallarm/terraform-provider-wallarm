// ci-tenant-delete disables and permanently deletes a test tenant.
// Used as a CI cleanup step — runs even if tests fail.
//
// Usage: go run scripts/ci-tenant-delete.go <client_id>
//
// Required env vars:
//
//	WALLARM_API_TOKEN  — API token with Global Administrator permissions
//	WALLARM_API_HOST   — API endpoint (e.g., https://api.wallarm.com)
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	wallarm "github.com/wallarm/wallarm-go"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: ci-tenant-delete <client_id>")
	}

	clientID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("invalid client_id %q: %v", os.Args[1], err)
	}

	token := os.Getenv("WALLARM_API_TOKEN")
	host := os.Getenv("WALLARM_API_HOST")
	if token == "" || host == "" {
		log.Fatal("WALLARM_API_TOKEN and WALLARM_API_HOST must be set")
	}

	headers := make(http.Header)
	headers.Add("X-WallarmAPI-Token", token)

	api, err := wallarm.New(
		wallarm.UsingBaseURL(host),
		wallarm.Headers(headers),
	)
	if err != nil {
		log.Fatalf("failed to create API client: %v", err)
	}

	// Step 1: Disable the tenant.
	enabled := false
	if _, err := api.ClientUpdate(&wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{ID: clientID},
		Fields: &wallarm.ClientFields{Enabled: &enabled},
	}); err != nil {
		log.Printf("[WARN] failed to disable tenant %d: %v (continuing to delete)", clientID, err)
	} else {
		fmt.Fprintf(os.Stderr, "Tenant %d disabled\n", clientID)
	}

	// Step 2: Permanently delete.
	if _, err := api.ClientDelete(&wallarm.ClientDelete{
		Filter: &wallarm.ClientFilter{ID: clientID},
	}); err != nil {
		log.Fatalf("failed to delete tenant %d: %v", clientID, err)
	}

	fmt.Fprintf(os.Stderr, "Tenant %d deleted\n", clientID)
}
