package wallarm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	vulnPrefixMinLength = 2
	vulnPrefixMaxLength = 4
)

func resourceWallarmTenant() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmTenantCreate,
		Read:   resourceWallarmTenantRead,
		Update: resourceWallarmTenantUpdate,
		Delete: resourceWallarmTenantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmTenantImport,
		},

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "The parent Client ID",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Tenant name",
			},
			"partner_uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Partner UUID. Inherited from parent client if not specified.",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tenant's unique identifier (UUID).",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the tenant is enabled.",
			},
		},
	}
}

// parseTenantID parses the resource ID in format {clientID}/{uuidPrefix}.
func parseTenantID(id string) (int, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("invalid resource ID %q, expected format: {clientID}/{uuidPrefix}", id)
	}

	var clientID int
	if _, err := fmt.Sscanf(parts[0], "%d", &clientID); err != nil {
		return 0, "", fmt.Errorf("invalid client_id %q: %w", parts[0], err)
	}

	return clientID, parts[1], nil
}

// uuidShort returns the first segment of a UUID (before the first hyphen).
// e.g. "1463a8c4-ef77-4eb9-87ca-8a281dfdfceb" → "1463a8c4"
func uuidShort(uuid string) string {
	if i := strings.IndexByte(uuid, '-'); i > 0 {
		return uuid[:i]
	}
	return uuid
}

// readTenantByID fetches a single client by its numeric ID and returns it.
func readTenantByID(client wallarm.API, tenantClientID int) (*wallarm.ClientInfoBody, error) {
	res, err := client.ClientRead(&wallarm.ClientRead{
		Limit: 1,
		Filter: &wallarm.ClientReadFilter{
			ClientFilter: wallarm.ClientFilter{ID: tenantClientID},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Body) == 0 {
		return nil, nil
	}
	return &res.Body[0], nil
}

// resourceWallarmTenantImport handles terraform import.
// Format: {clientID}/{uuidPrefix}
// The uuidPrefix is the first segment of the tenant UUID (before the first hyphen).
// Example: terraform import wallarm_tenant.my_tenant 8649/1463a8c4
func resourceWallarmTenantImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientID, uuidPrefix, err := parseTenantID(d.Id())
	if err != nil {
		return nil, err
	}

	client := m.(wallarm.API)

	// List accessible clients once to resolve UUID prefix → tenant client ID.
	res, err := client.ClientRead(&wallarm.ClientRead{
		Limit:  1000,
		Offset: 0,
		Filter: &wallarm.ClientReadFilter{
			Enabled: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	var matched []wallarm.ClientInfoBody
	for _, c := range res.Body {
		if strings.HasPrefix(c.UUID, uuidPrefix) {
			matched = append(matched, c)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no tenant found with UUID prefix %q", uuidPrefix)
	}
	if len(matched) > 1 {
		return nil, fmt.Errorf("UUID prefix %q is ambiguous — matches %d tenants, use more characters", uuidPrefix, len(matched))
	}

	// Store the resolved tenant client ID and uuid prefix as the canonical ID.
	d.Set("client_id", clientID)
	d.SetId(fmt.Sprintf("%d/%s", matched[0].ID, uuidShort(matched[0].UUID)))

	return []*schema.ResourceData{d}, nil
}

func resourceWallarmTenantCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	name := d.Get("name").(string)

	clientRes, err := client.ClientRead(&wallarm.ClientRead{
		Limit: 1,
		Filter: &wallarm.ClientReadFilter{
			ClientFilter: wallarm.ClientFilter{ID: clientID},
		},
	})
	if err != nil {
		return err
	}

	partnerUUID := d.Get("partner_uuid").(string)
	if partnerUUID == "" {
		partnerUUID = clientRes.Body[0].PartnerUUID
	}

	params := wallarm.ClientCreate{
		Name:        name,
		VulnPrefix:  generateVulnPrefix(name),
		PartnerUUID: partnerUUID,
	}

	res, err := client.ClientCreate(&params)
	if err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			return ImportAsExistsError("wallarm_tenant", "{clientID}/{uuidPrefix}")
		}
		return err
	}

	d.Set("client_id", clientID)
	d.SetId(fmt.Sprintf("%d/%s", res.Body.ID, uuidShort(res.Body.UUID)))

	return resourceWallarmTenantRead(d, m)
}

func resourceWallarmTenantRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)

	tenantClientID, _, err := parseTenantID(d.Id())
	if err != nil {
		return err
	}

	// Fetch single client by its ID — no listing needed.
	tenant, err := readTenantByID(client, tenantClientID)
	if err != nil {
		return err
	}

	if tenant == nil {
		log.Printf("[WARN] Tenant %d not found in API, removing from state", tenantClientID)
		d.SetId("")
		return nil
	}

	d.Set("name", tenant.Name)
	d.Set("uuid", tenant.UUID)
	d.Set("partner_uuid", tenant.PartnerUUID)
	d.Set("enabled", tenant.Enabled)

	return nil
}

func resourceWallarmTenantUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)

	tenantClientID, _, err := parseTenantID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		if _, err := client.ClientUpdate(&wallarm.ClientUpdate{
			Filter: &wallarm.ClientFilter{ID: tenantClientID},
			Fields: &wallarm.ClientFields{Name: d.Get("name").(string)},
		}); err != nil {
			return err
		}
	}

	return resourceWallarmTenantRead(d, m)
}

func resourceWallarmTenantDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)

	tenantClientID, _, err := parseTenantID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.ClientUpdate(&wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{ID: tenantClientID},
		Fields: &wallarm.ClientFields{Enabled: false},
	}); err != nil {
		return err
	}

	return nil
}

func generateVulnPrefix(name string) string {
	prefix := strings.ToUpper(name)

	if len(prefix) <= vulnPrefixMaxLength {
		return prefix
	}

	reg := regexp.MustCompile("[^A-Z]+")
	prefix = reg.ReplaceAllString(prefix, "")

	reg = regexp.MustCompile("[AEIOU]")
	prefix = reg.ReplaceAllString(prefix, "")

	prefix = removeConsecutiveDuplicates(prefix)

	if len(prefix) > vulnPrefixMaxLength {
		prefix = prefix[:vulnPrefixMaxLength]
	}

	if len(prefix) < vulnPrefixMinLength {
		prefix = strings.ToUpper(name)
		prefix = reg.ReplaceAllString(prefix, "")

		if len(prefix) > vulnPrefixMaxLength {
			prefix = prefix[:vulnPrefixMaxLength]
		}
	}

	return prefix
}

func removeConsecutiveDuplicates(s string) string {
	var result []rune
	var lastRune rune

	for _, r := range s {
		if r != lastRune {
			result = append(result, r)
			lastRune = r
		}
	}

	return string(result)
}
