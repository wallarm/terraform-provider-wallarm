package wallarm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	vulnPrefixMinLength = 2
	vulnPrefixMaxLength = 4
)

func resourceWallarmTenant() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmTenantCreate,
		Read:   resourceWallarmTenantRead,
		Delete: resourceWallarmTenantDelete,
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Client ID to perform changes",
				ForceNew:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"partner_uuid": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWallarmTenantCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

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

	partnerUUID := clientRes.Body[0].PartnerUUID
	if partnerUUID == "" {
		partnerUUID = d.Get("partner_uuid").(string)
	}

	params := wallarm.ClientCreate{
		Name:        name,
		VulnPrefix:  generateVulnPrefix(name),
		PartnerUUID: &partnerUUID,
	}

	res, err := client.ClientCreate(&params)
	if err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			existingID := fmt.Sprintf("%d/%s", clientID, params.Name)
			return ImportAsExistsError("wallarm_tenant", existingID)
		}
		return err
	}

	d.Set("client_id", clientID)

	tenantID := res.Body.ID
	if err := d.Set("tenant_id", tenantID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, params.Name, tenantID)
	d.SetId(resID)

	return resourceWallarmTenantRead(d, m)
}

func resourceWallarmTenantRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	tenantID := d.Get("tenant_id").(int)

	res, err := client.ClientRead(&wallarm.ClientRead{
		Limit: 1,
		Filter: &wallarm.ClientReadFilter{
			ClientFilter: wallarm.ClientFilter{ID: tenantID},
		},
	})
	if err != nil {
		return err
	}

	if len(res.Body) == 0 {
		body, err := json.Marshal(res)
		if err != nil {
			return err
		}
		log.Printf("[WARN] Tenant hasn't been found in API. Body: %s", body)

		d.SetId("")
		return nil
	}

	if err := d.Set("name", res.Body[0].Name); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmTenantDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	tenantID := d.Get("tenant_id").(int)

	if _, err := client.ClientUpdate(&wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{ID: tenantID},
		Fields: &wallarm.ClientFields{Enabled: false},
	}); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func generateVulnPrefix(name string) string {
	prefix := strings.ToUpper(name)

	if len(prefix) <= vulnPrefixMaxLength {
		return prefix
	}

	reg, _ := regexp.Compile("[^A-Z]+")
	prefix = reg.ReplaceAllString(prefix, "")

	reg, _ = regexp.Compile("[AEIOU]")
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
