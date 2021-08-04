package wallarm

import (
	"errors"
	"fmt"
	"sort"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmScanner() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmScannerCreate,
		Read:   resourceWallarmScannerRead,
		Update: resourceWallarmScannerUpdate,
		Delete: resourceWallarmScannerDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Client ID to perform changes",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},

			"element": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"resource_id": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func resourceWallarmScannerCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	elementInterface := d.Get("element").([]interface{})
	var element []string
	for _, v := range elementInterface {
		element = append(element, v.(string))
	}

	resources := make(map[string]int)
	var existingIDs []string
	for _, elem := range element {
		scannerBody := wallarm.ScannerCreate{
			Query:    elem,
			Clientid: clientID,
		}

		r, err := client.ScannerCreate(&scannerBody)
		if err != nil {
			if errors.Is(err, wallarm.ErrExistingResource) {
				existingIDs = append(existingIDs, elem)
				continue
			}
			return err
		}

		for _, obj := range r.Body.Objects {
			_, err := validation.IsIPAddress(obj.IP, "")
			if len(err) == 0 {
				resources[obj.IP] = obj.ID
			} else {
				resources[obj.Domain] = obj.ID
			}
		}
	}

	if len(existingIDs) != 0 {
		existingID := fmt.Sprintf("%d/%s", clientID, existingIDs)
		return ImportAsExistsError("wallarm_scanner", existingID)
	}

	if err := d.Set("resource_id", resources); err != nil {
		return err
	}

	if d.Get("disabled").(bool) {
		for k, resID := range resources {
			_, err := validation.IsIPAddress(k, "")
			var resType string
			if len(err) == 0 {
				resType = "ip"
			} else {
				resType = "domain"
			}
			scanUpdate := wallarm.ScannerUpdate{
				Disabled: d.Get("disabled").(bool),
				Clientid: clientID,
			}
			if err := client.ScannerUpdate(&scanUpdate, resType, resID); err != nil {
				return err
			}
		}
	}

	d.Set("client_id", clientID)

	resID := fmt.Sprintf("%d/%s", clientID, element)
	d.SetId(resID)

	return nil
}

func resourceWallarmScannerRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceWallarmScannerUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	elementInterface := d.Get("element").([]interface{})
	var element []string
	for _, v := range elementInterface {
		element = append(element, v.(string))
	}

	resID := fmt.Sprintf("%d/%s", clientID, element)
	d.SetId(resID)

	d.Set("client_id", clientID)

	switch resourceIDs := d.Get("resource_id").(type) {

	case map[string]interface{}:

		var stateElements []string
		for k := range resourceIDs {
			stateElements = append(stateElements, k)
		}

		sort.Strings(element)
		sort.Strings(stateElements)
		addElements := diffStringSlice(element, stateElements)
		delElements := diffStringSlice(stateElements, element)

		resIDreversed := make(map[int]string)
		for k, v := range resourceIDs {
			resIDreversed[v.(int)] = k
		}

		// Selective deletion according to affiliation.
		delElem, err := scopeDeletion(client, clientID, resourceIDs, delElements)
		if err != nil {
			return err
		}

		resources := make(map[string]int)
		var existingIDs []string
		for _, elem := range addElements {
			scannerBody := wallarm.ScannerCreate{
				Query:    elem,
				Clientid: clientID,
			}

			r, err := client.ScannerCreate(&scannerBody)
			if err != nil {
				if errors.Is(err, wallarm.ErrExistingResource) {
					existingIDs = append(existingIDs, elem)
					continue
				}
				return err
			}

			for _, obj := range r.Body.Objects {
				_, err := validation.IsIPAddress(obj.IP, "")
				if len(err) == 0 {
					resources[obj.IP] = obj.ID
				} else {
					resources[obj.Domain] = obj.ID
				}
			}
		}

		if len(existingIDs) != 0 {
			existingID := fmt.Sprintf("%d/%s", clientID, existingIDs)
			return ImportAsExistsError("wallarm_scanner", existingID)
		}

		resourcesUpdated := make(map[string]int)
		for k, v := range resIDreversed {
			if wallarm.Contains(delElem, k) {
				continue
			} else {
				resourcesUpdated[v] = k
			}
		}

		resourcesID := appendMap(resourcesUpdated, resources)

		if err := d.Set("resource_id", resourcesID); err != nil {
			return err
		}

		if d.HasChange("disabled") {
			for k, resID := range resourcesID {
				var resType string
				_, err := validation.IsIPAddress(k, "")
				if len(err) == 0 {
					resType = "ip"
				} else {
					resType = "domain"
				}
				scanUpdate := wallarm.ScannerUpdate{
					Disabled: d.Get("disabled").(bool),
					Clientid: clientID,
				}
				if err := client.ScannerUpdate(&scanUpdate, resType, resID); err != nil {
					return err
				}
			}
		}

	default:
		resourceWallarmScannerCreate(d, m)
	}

	return nil
}

func resourceWallarmScannerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

	switch resourceIDs := d.Get("resource_id").(type) {

	case map[string]interface{}:
		var stateElements []string
		for k := range resourceIDs {
			stateElements = append(stateElements, k)
		}
		_, err := scopeDeletion(client, clientID, resourceIDs, stateElements)
		if err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func scopeDeletion(client wallarm.API, clientID int, resourceIDs map[string]interface{}, delElements []string) ([]int, error) {
	var deleteIP []int
	var deleteDomain []int

	if len(delElements) != 0 {
		for _, v := range delElements {
			_, err := validation.IsIPAddress(v, "")
			if len(err) == 0 {
				deleteIP = append(deleteIP, resourceIDs[v].(int))
			} else {
				deleteDomain = append(deleteDomain, resourceIDs[v].(int))
			}
		}

		if len(deleteIP) != 0 {
			bulk := wallarm.ScannerDeleteBulk{
				Filter: &wallarm.ScannerFilter{
					ScannerCreate: &wallarm.ScannerCreate{
						Query:    "",
						Clientid: clientID},
					ID: deleteIP,
				},
			}
			var bulks []wallarm.ScannerDeleteBulk
			bulks = append(bulks, bulk)
			scannerBody := wallarm.ScannerDelete{
				Bulk: &bulks,
			}

			if err := client.ScannerDelete(&scannerBody, "ip"); err != nil {
				return nil, err
			}
		}

		if len(deleteDomain) != 0 {
			bulk := wallarm.ScannerDeleteBulk{
				Filter: &wallarm.ScannerFilter{
					ScannerCreate: &wallarm.ScannerCreate{
						Query:    "",
						Clientid: clientID},
					ID: deleteDomain,
				},
			}
			var bulks []wallarm.ScannerDeleteBulk
			bulks = append(bulks, bulk)
			scannerBody := wallarm.ScannerDelete{
				Bulk: &bulks,
			}

			if err := client.ScannerDelete(&scannerBody, "domain"); err != nil {
				return nil, err
			}
		}
	}
	delElem := append(deleteIP, deleteDomain...)
	return delElem, nil
}
