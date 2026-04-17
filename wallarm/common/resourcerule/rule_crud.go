package resourcerule

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/mapper/apitotf"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/mapper/tftoapi"
	"github.com/wallarm/wallarm-go"
)

// schemaHasKey checks whether the resource schema includes the given attribute
// by examining the cty type of the raw state. Returns true if undetermined.
func schemaHasKey(d *schema.ResourceData, key string) (exists bool) {
	exists = true // default to true if we can't determine
	defer func() { recover() }()
	return d.GetRawState().Type().HasAttribute(key)
}

// setIfExists calls d.Set for the given key only if the key is present in the
// resource schema. This is needed because ResourceRuleWallarmRead is shared
// across many resources, each of which defines only a subset of the fields.
// In SDK v2 d.Set() logs an [ERROR] and panics (in test mode) for keys not in
// the schema; checking beforehand avoids both.
func setIfExists(d *schema.ResourceData, key string, value interface{}) {
	if !schemaHasKey(d, key) {
		return
	}
	if err := d.Set(key, value); err != nil {
		log.Printf("[ERROR] error setting %s: %s", key, err)
	}
}

// nolint
func ResourceRuleWallarmRead(d *schema.ResourceData, clientID int, cli wallarm.API, opts ...common.ReadOption) error {
	var ruleID = d.Get("rule_id").(int)
	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
		},
	}
	actionHints, err := cli.HintRead(hint)
	if err != nil {
		return err
	}

	var updatedRule *wallarm.ActionBody
	for _, rule := range *actionHints.Body {
		if ruleID == rule.ID {
			updatedRule = &rule
			break
		}
	}

	if updatedRule == nil {
		d.SetId("")
		return nil
	}

	// Fields from commonResourceRuleFields — always present in every resource
	d.Set("rule_id", updatedRule.ID)
	d.Set("client_id", clientID)
	d.Set("active", updatedRule.Active)
	d.Set("title", updatedRule.Title)
	d.Set("mitigation", updatedRule.Mitigation)
	d.Set("set", updatedRule.Set)
	d.Set("variativity_disabled", updatedRule.VariativityDisabled)
	d.Set("comment", updatedRule.Comment)

	// Resource-specific fields — use setIfExists because each resource
	// only defines a subset of these in its schema. In SDK v2, d.Set
	// panics for keys not in the schema.
	setIfExists(d, "point", WrapPointElements(updatedRule.Point))
	setIfExists(d, "threshold", apitotf.Threshold(updatedRule.Threshold))
	setIfExists(d, "reaction", apitotf.Reaction(updatedRule.Reaction))
	setIfExists(d, "mode", updatedRule.Mode)
	setIfExists(d, "enumerated_parameters", apitotf.EnumeratedParameters(updatedRule.EnumeratedParameters))
	setIfExists(d, "advanced_conditions", apitotf.AdvancedConditions(updatedRule.AdvancedConditions))
	setIfExists(d, "arbitrary_conditions", apitotf.ArbitraryConditions(updatedRule.ArbitraryConditions))
	setIfExists(d, "counter", updatedRule.Counter)
	setIfExists(d, "size", updatedRule.Size)
	setIfExists(d, "size_unit", updatedRule.SizeUnit)
	setIfExists(d, "debug_enabled", updatedRule.DebugEnabled)
	setIfExists(d, "introspection", updatedRule.Introspection)
	setIfExists(d, "max_depth", updatedRule.MaxDepth)
	setIfExists(d, "max_value_size_kb", updatedRule.MaxValueSizeKb)
	setIfExists(d, "max_doc_size_kb", updatedRule.MaxDocSizeKb)
	setIfExists(d, "max_alias_size_kb", updatedRule.MaxAliasesSizeKb)
	setIfExists(d, "max_doc_per_batch", updatedRule.MaxDocPerBatch)
	setIfExists(d, "attack_type", updatedRule.AttackType)
	setIfExists(d, "stamp", updatedRule.Stamp)
	setIfExists(d, "file_type", updatedRule.FileType)
	setIfExists(d, "name", updatedRule.Name)
	setIfExists(d, "burst", updatedRule.Burst)
	setIfExists(d, "delay", updatedRule.Delay)
	setIfExists(d, "rate", updatedRule.Rate)
	setIfExists(d, "rsp_status", updatedRule.RspStatus)
	setIfExists(d, "time_unit", updatedRule.TimeUnit)
	setIfExists(d, "parser", updatedRule.Parser)
	setIfExists(d, "state", updatedRule.State)
	setIfExists(d, "overlimit_time", updatedRule.OverlimitTime)
	setIfExists(d, "values", updatedRule.Values)
	setIfExists(d, "regex", updatedRule.Regex)
	setIfExists(d, "regex_id", updatedRule.RegexID)

	actionsSet := schema.Set{F: HashActionDetails}
	for _, a := range updatedRule.Action {
		acts, err := ActionDetailsToMap(a)
		if err != nil {
			return fmt.Errorf("failed to map action details: %w", err)
		}
		TransformAPIActionToSchema(acts)
		actionsSet.Add(acts)
	}
	setIfExists(d, "action", &actionsSet)

	return nil
}

// nolint
func ResourceRuleWallarmCreate(
	ctx context.Context,
	d *schema.ResourceData,
	cli wallarm.API,
	clientID int,
	ruleType, attackType string,
	readMethod func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics,
	m interface{},
) diag.Diagnostics {
	actionsFromState := d.Get("action").(*schema.Set)
	action, err := ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(errors.WithMessage(err, "on ExpandSetToActionDetailsList"))
	}

	enumeratedParametersFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "enumerated_parameters")
	enumeratedParameters, err := tftoapi.EnumeratedParameters(enumeratedParametersFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	reactionFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "reaction")
	reaction, err := tftoapi.Reaction(reactionFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	thresholdFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "threshold")
	threshold, err := tftoapi.Threshold(thresholdFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	advancedConditionsFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "advanced_conditions")
	advancedConditions, err := tftoapi.AdvancedConditions(advancedConditionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	arbitraryConditionsFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "arbitrary_conditions")
	arbitraryConditions, err := tftoapi.ArbitraryConditionsReq(arbitraryConditionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	pointFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "point")
	points, err := ExpandPointsToTwoDimensionalArray(pointFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	wm := &wallarm.ActionCreate{
		Type:                 ruleType,
		Clientid:             clientID,
		Action:               &action,
		Point:                points,
		Validated:            false,
		Comment:              GetValueWithTypeCastingOrDefault[string](d, "comment"),
		VariativityDisabled:  true,
		Set:                  GetValueWithTypeCastingOrDefault[string](d, "set"),
		Active:               GetValueWithTypeCastingOrDefault[bool](d, "active"),
		Title:                GetValueWithTypeCastingOrDefault[string](d, "title"),
		AttackType:           attackType,
		Reaction:             reaction,
		Threshold:            threshold,
		EnumeratedParameters: enumeratedParameters,
		AdvancedConditions:   advancedConditions,
		ArbitraryConditions:  arbitraryConditions,
		Mode:                 GetValueWithTypeCastingOrDefault[string](d, "mode"),
		MaxDepth:             GetValueWithTypeCastingOrDefault[int](d, "max_depth"),
		MaxValueSizeKb:       GetValueWithTypeCastingOrDefault[int](d, "max_value_size_kb"),
		MaxDocSizeKb:         GetValueWithTypeCastingOrDefault[int](d, "max_doc_size_kb"),
		MaxAliasesSizeKb:     GetValueWithTypeCastingOrDefault[int](d, "max_alias_size_kb"),
		MaxDocPerBatch:       GetValueWithTypeCastingOrDefault[int](d, "max_doc_per_batch"),
		Introspection:        GetPointerWithTypeCastingOrDefault[bool](d, "introspection"),
		DebugEnabled:         GetPointerWithTypeCastingOrDefault[bool](d, "debug_enabled"),
		Size:                 GetValueWithTypeCastingOrDefault[int](d, "size"),
		SizeUnit:             GetValueWithTypeCastingOrDefault[string](d, "size_unit"),
	}

	actionResp, err := cli.HintCreate(wm)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return readMethod(ctx, d, m)
}

func GetValueWithTypeCastingOrDefault[T any](d *schema.ResourceData, name string) T {
	var defaultValue T
	resourceValue := d.Get(name)
	if resourceValue == nil {
		return defaultValue
	}
	v, ok := resourceValue.(T)
	if !ok {
		return defaultValue
	}
	return v
}

func GetPointerWithTypeCastingOrDefault[T any](d *schema.ResourceData, name string) *T {
	resourceValue := d.Get(name)
	if resourceValue == nil {
		return nil
	}
	v, ok := d.Get(name).(T)
	if !ok {
		return nil
	}
	return &v
}
