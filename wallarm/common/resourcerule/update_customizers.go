package resourcerule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

// Reusable UpdateCustomizers — pass to Update() to wire per-resource mutable
// fields into HintUpdateV3Params. Each customizer reads one schema field via
// d.Get and writes the corresponding pointer on the params struct.

func WithMode(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Mode = GetPointerWithTypeCastingOrDefault[string](d, "mode")
	return nil
}

func WithAttackType(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.AttackType = GetPointerWithTypeCastingOrDefault[string](d, "attack_type")
	return nil
}

func WithStamp(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Stamp = GetPointerWithTypeCastingOrDefault[int](d, "stamp")
	return nil
}

func WithRegex(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Regex = GetPointerWithTypeCastingOrDefault[string](d, "regex")
	return nil
}

func WithLoginRegex(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.LoginRegex = GetPointerWithTypeCastingOrDefault[string](d, "login_regex")
	return nil
}

func WithCaseSensitive(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.CaseSensitive = GetPointerWithTypeCastingOrDefault[bool](d, "case_sensitive")
	return nil
}

func WithCredStuffType(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.CredStuffType = GetPointerWithTypeCastingOrDefault[string](d, "cred_stuff_type")
	return nil
}

func WithSize(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Size = GetPointerWithTypeCastingOrDefault[int](d, "size")
	return nil
}

func WithMaxDepth(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.MaxDepth = GetPointerWithTypeCastingOrDefault[int](d, "max_depth")
	return nil
}

func WithMaxValueSizeKb(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.MaxValueSizeKb = GetPointerWithTypeCastingOrDefault[int](d, "max_value_size_kb")
	return nil
}

func WithMaxDocSizeKb(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.MaxDocSizeKb = GetPointerWithTypeCastingOrDefault[int](d, "max_doc_size_kb")
	return nil
}

func WithMaxDocPerBatch(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.MaxDocPerBatch = GetPointerWithTypeCastingOrDefault[int](d, "max_doc_per_batch")
	return nil
}

func WithIntrospection(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Introspection = GetPointerWithTypeCastingOrDefault[bool](d, "introspection")
	return nil
}

func WithDebugEnabled(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.DebugEnabled = GetPointerWithTypeCastingOrDefault[bool](d, "debug_enabled")
	return nil
}

func WithOverlimitTime(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.OverlimitTime = GetPointerWithTypeCastingOrDefault[int](d, "overlimit_time")
	return nil
}

func WithParser(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Parser = GetPointerWithTypeCastingOrDefault[string](d, "parser")
	return nil
}

func WithState(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.State = GetPointerWithTypeCastingOrDefault[string](d, "state")
	return nil
}

func WithDelay(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Delay = GetPointerWithTypeCastingOrDefault[int](d, "delay")
	return nil
}

func WithBurst(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Burst = GetPointerWithTypeCastingOrDefault[int](d, "burst")
	return nil
}

func WithRate(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Rate = GetPointerWithTypeCastingOrDefault[int](d, "rate")
	return nil
}

func WithRspStatus(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.RspStatus = GetPointerWithTypeCastingOrDefault[int](d, "rsp_status")
	return nil
}

func WithTimeUnit(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.TimeUnit = GetPointerWithTypeCastingOrDefault[string](d, "time_unit")
	return nil
}

func WithName(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.Name = GetPointerWithTypeCastingOrDefault[string](d, "name")
	return nil
}

func WithValues(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	out := ConvertToStringSlice(d.Get("values").([]interface{}))
	p.Values = &out
	return nil
}

func WithFileType(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	p.FileType = GetPointerWithTypeCastingOrDefault[string](d, "file_type")
	return nil
}

func WithThreshold(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	t, err := ThresholdToAPI(d.Get("threshold").([]interface{}))
	if err != nil {
		return err
	}
	p.Threshold = t
	return nil
}

func WithReaction(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	r, err := ReactionToAPI(d.Get("reaction").([]interface{}))
	if err != nil {
		return err
	}
	p.Reaction = r
	return nil
}

func WithEnumeratedParameters(d *schema.ResourceData, p *wallarm.HintUpdateV3Params) error {
	e, err := EnumeratedParametersToAPI(d.Get("enumerated_parameters").([]interface{}))
	if err != nil {
		return err
	}
	p.EnumeratedParameters = e
	return nil
}
