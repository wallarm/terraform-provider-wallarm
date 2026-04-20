package wallarm

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

const eventTypeSIEM = "siem"

// isNotFoundError checks if the error is a Wallarm API 404 response.
func isNotFoundError(err error) bool {
	var apiErr *wallarm.APIError
	return stderrors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// importIntegration parses a 3-part integration import ID
// ("{client_id}/{type}/{integration_id}"), validates the type segment, and
// populates client_id, integration_id on the ResourceData.
func importIntegration(integrationType string) schema.StateContextFunc {
	return func(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
		parts := strings.SplitN(d.Id(), "/", 4)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{client_id}/%s/{integration_id}\"", d.Id(), integrationType)
		}
		clientID, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid client_id: %w", err)
		}
		if parts[1] != integrationType {
			return nil, fmt.Errorf("invalid type segment %q, expected %q", parts[1], integrationType)
		}
		integrationID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid integration_id: %w", err)
		}
		d.Set("client_id", clientID)
		d.Set("integration_id", integrationID)
		d.SetId(fmt.Sprintf("%d/%s/%d", clientID, integrationType, integrationID))
		return []*schema.ResourceData{d}, nil
	}
}

// validateWithHeadersOnlySiem returns a CustomizeDiffFunc that ensures
// with_headers is only set to true on events of type "siem".
func validateWithHeadersOnlySiem() schema.CustomizeDiffFunc {
	return customdiff.All(
		func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			events, ok := d.GetOk("event")
			if !ok {
				return nil
			}
			for _, e := range events.(*schema.Set).List() {
				m := e.(map[string]interface{})
				eventType, _ := m["event_type"].(string)
				withHeaders, _ := m["with_headers"].(bool)
				if withHeaders && eventType != eventTypeSIEM {
					return fmt.Errorf("with_headers can only be set for the 'siem' event type, got event_type=%q", eventType)
				}
			}
			return nil
		},
	)
}

// TODO: add test — empty set per resource type returns defaults, populated events with hit→siem mapping, with_headers on siem
func expandWallarmEventToIntEvents(d interface{}, resourceType string) *[]wallarm.IntegrationEvents {
	cfg := d.(*schema.Set).List()
	events := []wallarm.IntegrationEvents{}
	if len(cfg) == 0 || cfg[0] == nil {
		var defaultEvents []map[string]interface{}
		switch resourceType {
		case "email":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "aasm_report",
					"active":     false},
				{
					"event_type": "api_discovery_hourly_changes_report",
					"active":     false},
				{
					"event_type": "api_discovery_daily_changes_report",
					"active":     false},
				{
					"event_type": "report_daily",
					"active":     false},
				{
					"event_type": "report_weekly",
					"active":     false},
				{
					"event_type": "report_monthly",
					"active":     false},
			}
		case "data_dog", "insight_connect", "splunk", "sumo_logic", "web_hooks":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": eventTypeSIEM,
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "number_of_requests_per_hour",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
					"active":     false},
				{
					"event_type": "system",
					"active":     false},
			}
		case "telegram":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
					"active":     false},
				{
					"event_type": "report_daily",
					"active":     false},
				{
					"event_type": "report_weekly",
					"active":     false},
				{
					"event_type": "report_monthly",
					"active":     false},
			}
		case "ms_teams", "opsgenie", "pager_duty", "slack":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
					"active":     false},
			}
		default:
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
					"active":     false},
			}
		}
		for _, v := range defaultEvents {
			event := wallarm.IntegrationEvents{
				Event:  v["event_type"].(string),
				Active: v["active"].(bool),
			}
			events = append(events, event)
		}
		return &events
	}

	for _, conf := range cfg {

		m := conf.(map[string]interface{})
		e := wallarm.IntegrationEvents{}
		event, ok := m["event_type"]
		if ok {
			if event.(string) == "hit" {
				e.Event = eventTypeSIEM
			} else {
				e.Event = event.(string)
			}
		}

		active, ok := m["active"]
		if ok {
			e.Active = active.(bool)
		}
		// with_headers is only applicable to the siem event type
		if e.Event == eventTypeSIEM {
			if wh, ok := m["with_headers"]; ok {
				whBool := wh.(bool)
				e.WithHeaders = &whBool
			}
		}
		events = append(events, e)
	}
	return &events
}
