// ci-tenant-create creates a temporary tenant for acceptance tests.
// Outputs the tenant client_id to stdout for use in CI pipelines.
//
// Usage: go run ./scripts/ci-tenant-create/
//
// Required env vars:
//
//	WALLARM_API_TOKEN      — API token with Global Administrator permissions
//	WALLARM_API_HOST       — API endpoint (e.g., https://api.wallarm.com)
//	WALLARM_PARTNER_UUID   — partner UUID for the test tenant
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	wallarm "github.com/wallarm/wallarm-go"
)

func main() {
	token := os.Getenv("WALLARM_API_TOKEN")
	host := os.Getenv("WALLARM_API_HOST")
	partnerUUID := os.Getenv("WALLARM_PARTNER_UUID")
	if token == "" || host == "" || partnerUUID == "" {
		log.Fatal("WALLARM_API_TOKEN, WALLARM_API_HOST, and WALLARM_PARTNER_UUID must be set")
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

	name := fmt.Sprintf("ci-test-%s", time.Now().UTC().Format("20060102-150405"))

	res, err := api.ClientCreate(&wallarm.ClientCreate{
		Name:        name,
		PartnerUUID: partnerUUID,
	})
	if err != nil {
		log.Fatalf("failed to create tenant: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Created tenant %q (client_id=%d)\n", name, res.Body.ID)
	// Output only the client_id — CI pipeline captures this.
	fmt.Print(res.Body.ID)
}
