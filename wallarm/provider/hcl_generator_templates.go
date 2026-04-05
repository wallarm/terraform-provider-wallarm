package wallarm

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
	"github.com/zclconf/go-cty/cty"
)

// ActionCondition represents a single action condition from data.wallarm_hits.
type ActionCondition struct {
	Type  string
	Point []string // e.g. ["header", "HOST"] or ["path", "0"] or ["action_name", "login"]
	Value string
}

// StaticRuleConfig holds config for generating a single static resource block.
type StaticRuleConfig struct {
	ClientID   int
	Comment    string
	Point      [][]string
	Actions    []ActionCondition
	Stamp      int    // for disable_stamp
	AttackType string // for disable_attack_type
}

// generateStaticDisableStamp writes a single wallarm_rule_disable_stamp resource block.
func generateStaticDisableStamp(f *hclwrite.File, name string, cfg StaticRuleConfig) {
	block := f.Body().AppendNewBlock("resource", []string{"wallarm_rule_disable_stamp", name})
	body := block.Body()

	body.SetAttributeValue("client_id", cty.NumberIntVal(int64(cfg.ClientID)))
	body.SetAttributeValue("comment", cty.StringVal(cfg.Comment))
	body.SetAttributeValue("variativity_disabled", cty.True)
	body.SetAttributeValue("stamp", cty.NumberIntVal(int64(cfg.Stamp)))
	body.AppendNewline()
	writePointAttribute(body, cfg.Point)
	body.AppendNewline()
	writeActionBlocks(body, cfg.Actions)

	f.Body().AppendNewline()
}

// generateStaticDisableAttackType writes a single wallarm_rule_disable_attack_type resource block.
func generateStaticDisableAttackType(f *hclwrite.File, name string, cfg StaticRuleConfig) {
	block := f.Body().AppendNewBlock("resource", []string{"wallarm_rule_disable_attack_type", name})
	body := block.Body()

	body.SetAttributeValue("client_id", cty.NumberIntVal(int64(cfg.ClientID)))
	body.SetAttributeValue("comment", cty.StringVal(cfg.Comment))
	body.SetAttributeValue("variativity_disabled", cty.True)
	body.SetAttributeValue("attack_type", cty.StringVal(cfg.AttackType))
	body.AppendNewline()
	writePointAttribute(body, cfg.Point)
	body.AppendNewline()
	writeActionBlocks(body, cfg.Actions)

	f.Body().AppendNewline()
}

// writeMovedBlock appends a moved { from = ... to = ... } block.
// from: wallarm_rule_disable_stamp.fp["key"]
// to:   wallarm_rule_disable_stamp.fp_reqid_key
func writeMovedBlock(f *hclwrite.File, resourceType, fromName, forEachKey, toName string) {
	block := f.Body().AppendNewBlock("moved", nil)
	body := block.Body()
	body.SetAttributeRaw("from", hclMovedRef(resourceType, fromName, forEachKey))
	body.SetAttributeRaw("to", hclSimpleRef(resourceType, toName))
	f.Body().AppendNewline()
}

// hclMovedRef produces tokens for: wallarm_rule_disable_stamp.fp["key"]
func hclMovedRef(resourceType, name, key string) hclwrite.Tokens {
	ref := fmt.Sprintf("%s.%s[\"%s\"]", resourceType, name, key)
	return hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(ref)}, // TokenIdent
	}
}

// hclSimpleRef produces tokens for: wallarm_rule_disable_stamp.fp_reqid_key
func hclSimpleRef(resourceType, name string) hclwrite.Tokens {
	ref := fmt.Sprintf("%s.%s", resourceType, name)
	return hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(ref)}, // TokenIdent
	}
}

// writeActionBlocks appends action {} blocks to a body from action conditions.
// Uses resourcerule.ActionDetailToSchemaItem for correct point-value type handling.
func writeActionBlocks(body *hclwrite.Body, conditions []ActionCondition) {
	for _, c := range conditions {
		// Convert to ActionDetails so we can reuse the shared conversion logic.
		point := make([]interface{}, len(c.Point))
		for i, p := range c.Point {
			point[i] = p
		}

		var value interface{} = c.Value
		if c.Value == "" {
			value = nil
		}

		schemaItem := resourcerule.ActionDetailToSchemaItem(wallarm.ActionDetails{
			Type:  c.Type,
			Point: point,
			Value: value,
		})

		actionBlock := body.AppendNewBlock("action", nil)
		ab := actionBlock.Body()

		condType := schemaItem["type"].(string)
		condValue := schemaItem["value"].(string)

		if condType != "" {
			ab.SetAttributeValue("type", cty.StringVal(condType))
		}
		if condValue != "" {
			ab.SetAttributeValue("value", cty.StringVal(condValue))
		}

		pointMap := make(map[string]cty.Value)
		for k, v := range schemaItem["point"].(map[string]interface{}) {
			pointMap[k] = cty.StringVal(fmt.Sprintf("%v", v))
		}
		ab.SetAttributeValue("point", cty.ObjectVal(pointMap))
	}
}

// writePointAttribute writes the point = [[...], [...]] attribute.
func writePointAttribute(body *hclwrite.Body, point [][]string) {
	vals := make([]cty.Value, 0, len(point))
	for _, inner := range point {
		innerVals := make([]cty.Value, 0, len(inner))
		for _, s := range inner {
			innerVals = append(innerVals, cty.StringVal(s))
		}
		if len(innerVals) == 0 {
			vals = append(vals, cty.ListValEmpty(cty.String))
		} else {
			vals = append(vals, cty.ListVal(innerVals))
		}
	}
	if len(vals) == 0 {
		body.SetAttributeValue("point", cty.ListValEmpty(cty.List(cty.String)))
	} else {
		body.SetAttributeValue("point", cty.TupleVal(vals))
	}
}
