package wallarm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	vulnPrefixMinLength = 2
	vulnPrefixMaxLength = 4
)

func resourceWallarmTenant() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmTenantCreate,
		ReadContext:   resourceWallarmTenantRead,
		UpdateContext: resourceWallarmTenantUpdate,
		DeleteContext: resourceWallarmTenantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmTenantImport,
		},

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
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
			"prevent_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "If true, the tenant will only be disabled on destroy, not deleted. " +
					"Set to false and export WALLARM_ALLOW_CLIENT_DELETE=1 to permanently delete.",
			},
		},
	}
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
// Format: {client_id}
// Example: terraform import wallarm_tenant.my_tenant 110310
func resourceWallarmTenantImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	tenantClientID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid tenant client_id %q: %w", d.Id(), err)
	}

	client := apiClient(m)

	tenant, err := readTenantByID(client, tenantClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenant %d: %w", tenantClientID, err)
	}
	if tenant == nil {
		return nil, fmt.Errorf("tenant with client_id %d not found", tenantClientID)
	}

	d.SetId(strconv.Itoa(tenantClientID))
	d.Set("client_id", tenantClientID)

	return []*schema.ResourceData{d}, nil
}

func resourceWallarmTenantCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	meta := m.(*ProviderMeta)

	name := d.Get("name").(string)

	clientRes, err := client.ClientRead(&wallarm.ClientRead{
		Limit: 1,
		Filter: &wallarm.ClientReadFilter{
			ClientFilter: wallarm.ClientFilter{ID: meta.DefaultClientID},
		},
	})
	if err != nil {
		return diag.FromErr(err)
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
			return diag.FromErr(ImportAsExistsError("wallarm_tenant", "{client_id}"))
		}
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(res.Body.ID))
	d.Set("client_id", res.Body.ID)

	return resourceWallarmTenantRead(context.TODO(), d, m)
}

func resourceWallarmTenantRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

	tenantClientID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid tenant ID %q: %w", d.Id(), err))
	}

	tenant, err := readTenantByID(client, tenantClientID)
	if err != nil {
		return diag.FromErr(err)
	}

	if tenant == nil {
		log.Printf("[WARN] Tenant %d not found in API, removing from state", tenantClientID)
		d.SetId("")
		return nil
	}

	d.Set("client_id", tenantClientID)
	d.Set("name", tenant.Name)
	d.Set("uuid", tenant.UUID)
	d.Set("partner_uuid", tenant.PartnerUUID)
	d.Set("enabled", tenant.Enabled)

	return nil
}

func resourceWallarmTenantUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

	tenantClientID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid tenant ID %q: %w", d.Id(), err))
	}

	if d.HasChange("name") {
		if _, err := client.ClientUpdate(&wallarm.ClientUpdate{
			Filter: &wallarm.ClientFilter{ID: tenantClientID},
			Fields: &wallarm.ClientFields{Name: d.Get("name").(string)},
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceWallarmTenantRead(context.TODO(), d, m)
}

func resourceWallarmTenantDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

	tenantClientID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid tenant ID %q: %w", d.Id(), err))
	}

	// Always disable the tenant first.
	enabled := false
	if _, err := client.ClientUpdate(&wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{ID: tenantClientID},
		Fields: &wallarm.ClientFields{Enabled: &enabled},
	}); err != nil {
		return diag.FromErr(fmt.Errorf("failed to disable tenant %d: %w", tenantClientID, err))
	}

	// Permanently delete only when both safeguards are explicitly removed.
	preventDestroy := d.Get("prevent_destroy").(bool)
	allowDelete := os.Getenv("WALLARM_ALLOW_CLIENT_DELETE") != ""

	if preventDestroy || !allowDelete {
		log.Printf("[WARN] Tenant %d has been disabled but not deleted. "+
			"To permanently delete, set prevent_destroy = false and export WALLARM_ALLOW_CLIENT_DELETE=1", tenantClientID)
		return nil
	}

	if _, err := client.ClientDelete(&wallarm.ClientDelete{
		Filter: &wallarm.ClientFilter{ID: tenantClientID},
	}); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete tenant %d: %w", tenantClientID, err))
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
