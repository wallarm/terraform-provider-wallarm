package wallarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/wallarm/wallarm-go"
)

var allowedAttackTypes = map[string]bool{
	"sqli":           true,
	"nosqli":         true,
	"rce":            true,
	"ssi":            true,
	"ssti":           true,
	"ldapi":          true,
	"mail_injection": true,
	"ssrf":           true,
	"ptrav":          true,
	"xxe":            true,
	"scanner":        true,
	"xss":            true,
	"redir":          true,
	"crlf":           true,
}

func resourceWallarmFalsePositive() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmFalsePositiveCreate,
		Read:   resourceWallarmFalsePositiveRead,
		Delete: resourceWallarmFalsePositiveDelete,
		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"request_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"days_back": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      89,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 365),
				Description:  "How many days back to search for the request (default 89). Limits the time window of the hit query.",
			},
			"attack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attack_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"action_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"stamp": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"point": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// hitAPIResponse is the row-based response from /v1/objects/hit.
type hitAPIResponse struct {
	Status int       `json:"status"`
	Body   []hitBody `json:"body"`
}

// hitBody represents a single hit returned by the API.
type hitBody struct {
	Type      string        `json:"type"`
	Stamps    []int         `json:"stamps"`
	Point     []interface{} `json:"point"`
	Domain    string        `json:"domain"`
	Path      string        `json:"path"`
	AttackID  []string      `json:"attackid"`
	RequestID string        `json:"request_id"`
}

func resourceWallarmFalsePositiveCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	requestID := d.Get("request_id").(string)
	daysBack := d.Get("days_back").(int)

	// Compute time window for the hit search.
	now := time.Now().UTC()
	startOfWindow := now.AddDate(0, 0, -daysBack).Truncate(24 * time.Hour)
	endOfWindow := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	start := int(startOfWindow.Unix())
	end := int(endOfWindow.Unix())

	// 1. Look up the hit(s) matching this request_id.
	hitReqBody := map[string]interface{}{
		"filter": map[string]interface{}{
			"clientid":   []int{clientID},
			"request_id": []string{requestID},
			"time":       [][]int{{start, end}},
		},
		"limit": 100,
	}

	hitRespBytes, err := doAPIPost(APIURL+"/v1/objects/hit", hitReqBody)
	if err != nil {
		return fmt.Errorf("error looking up hit for request_id %q: %w", requestID, err)
	}

	var hitResp hitAPIResponse
	if err := json.Unmarshal(hitRespBytes, &hitResp); err != nil {
		return fmt.Errorf("error parsing hit response: %w", err)
	}

	if len(hitResp.Body) == 0 {
		return fmt.Errorf("no hit found for request_id %q in the last %d days", requestID, daysBack)
	}

	// Use the first hit for metadata. All hits sharing a request_id come from the same request.
	firstHit := hitResp.Body[0]
	attackType := firstHit.Type
	domain := firstHit.Domain
	attackPath := firstHit.Path

	// 2. Validate the attack type supports false-positive suppression via disable_stamp.
	if !allowedAttackTypes[attackType] {
		return fmt.Errorf("attack type %q is not supported for false positive suppression (must be one of: sqli, nosqli, rce, ssi, ssti, ldapi, mail_injection, ssrf, ptrav, xxe, scanner, xss, redir, crlf)", attackType)
	}

	// Format the attackid from the compound array.
	attackID := ""
	if len(firstHit.AttackID) >= 2 {
		attackID = firstHit.AttackID[0] + ":" + firstHit.AttackID[1]
	}

	d.Set("attack_id", attackID)     //nolint:errcheck
	d.Set("attack_type", attackType) //nolint:errcheck
	d.Set("domain", domain)          //nolint:errcheck
	d.Set("path", attackPath)        //nolint:errcheck

	// 3. Collect unique (pointJSON, stamp) pairs across all hits for this request.
	// map[pointJSON]map[stamp]bool
	uniquePairs := make(map[string]map[int]bool)

	for _, hit := range hitResp.Body {
		if len(hit.Point) == 0 || len(hit.Stamps) == 0 {
			continue
		}

		pointJSON, err := json.Marshal(hit.Point)
		if err != nil {
			log.Printf("[WARN] wallarm_rule_false_positive: could not marshal point %v: %v", hit.Point, err)
			continue
		}
		pointKey := string(pointJSON)

		if uniquePairs[pointKey] == nil {
			uniquePairs[pointKey] = make(map[int]bool)
		}
		for _, stamp := range hit.Stamps {
			uniquePairs[pointKey][stamp] = true
		}
	}

	if len(uniquePairs) == 0 {
		return fmt.Errorf("hit for request_id %q has no stamps — cannot create disable_stamp rules (attack type may not use stamp-based detection)", requestID)
	}

	// 4. Create one disable_stamp hint per unique (point, stamp) pair.
	var createdRules []interface{}

	for pointKey, stamps := range uniquePairs {
		var pointElems []interface{}
		if err := json.Unmarshal([]byte(pointKey), &pointElems); err != nil {
			return fmt.Errorf("error unmarshalling point %q: %w", pointKey, err)
		}

		for stamp := range stamps {
			wm := &wallarm.ActionCreate{
				Type:     "disable_stamp",
				Clientid: clientID,
				Stamp:    stamp,
				Point:    wallarm.TwoDimensionalSlice{pointElems},
				Action: &[]wallarm.ActionDetails{
					{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: domain},
					{Type: "iequal", Point: []interface{}{"uri"}, Value: attackPath},
				},
				Comment:             "Managed by Terraform (false positive suppression)",
				Validated:           false,
				VariativityDisabled: true,
			}

			actionResp, err := client.HintCreate(wm)
			if err != nil {
				return fmt.Errorf("error creating disable_stamp rule for stamp %d, point %q: %w", stamp, pointKey, err)
			}

			createdRules = append(createdRules, map[string]interface{}{
				"rule_id":   actionResp.Body.ID,
				"action_id": actionResp.Body.ActionID,
				"stamp":     stamp,
				"point":     pointKey,
			})
		}
	}

	if err := d.Set("created_rules", createdRules); err != nil {
		return fmt.Errorf("error setting created_rules: %w", err)
	}

	d.SetId(fmt.Sprintf("%d/%s", clientID, requestID))

	return resourceWallarmFalsePositiveRead(d, m)
}

func resourceWallarmFalsePositiveRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	rawRules, ok := d.GetOk("created_rules")
	if !ok {
		return nil
	}

	var existingRules []interface{}

	for _, rawRule := range rawRules.([]interface{}) {
		rule := rawRule.(map[string]interface{})
		ruleID := rule["rule_id"].(int)

		_, err := findRule(client, clientID, ruleID)
		if err != nil {
			if _, notFound := err.(*ruleNotFoundError); notFound {
				log.Printf("[WARN] wallarm_rule_false_positive: rule %d not found, removing from state", ruleID)
				continue
			}
			return err
		}
		existingRules = append(existingRules, rule)
	}

	if len(existingRules) == 0 {
		d.SetId("")
		return nil
	}

	return d.Set("created_rules", existingRules)
}

func resourceWallarmFalsePositiveDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	rawRules, ok := d.GetOk("created_rules")
	if !ok {
		return nil
	}

	for _, rawRule := range rawRules.([]interface{}) {
		rule := rawRule.(map[string]interface{})
		ruleID := rule["rule_id"].(int)

		h := &wallarm.HintDelete{
			Filter: &wallarm.HintDeleteFilter{
				Clientid: []int{clientID},
				ID:       ruleID,
			},
		}

		if err := client.HintDelete(h); err != nil {
			notFound, _ := isNotFoundError(err)
			if !notFound {
				return fmt.Errorf("error deleting disable_stamp rule %d: %w", ruleID, err)
			}
			log.Printf("[WARN] wallarm_rule_false_positive: rule %d already deleted", ruleID)
		}
	}

	return nil
}

// doAPIPost makes an authenticated POST request using the package-level HTTPClient and APIHeaders.
func doAPIPost(url string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request to %s: %w", url, err)
	}

	for key, values := range APIHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing HTTP request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, url, string(respBody))
	}

	return respBody, nil
}
