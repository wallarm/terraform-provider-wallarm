package resourcerule

import "fmt"

const (
	Path   = "path"
	Iequal = "iequal"
	Header = "header"
)

// EnumeratedParameters.Mode values.
const (
	modeExact  = "exact"
	modeRegexp = "regexp"
)

type ReadOption string

const (
	ReadOptionWithPoint                ReadOption = "with_point"
	ReadOptionWithAction               ReadOption = "without_action"
	ReadOptionWithRegexID              ReadOption = "with_regex_id"
	ReadOptionWithMode                 ReadOption = "with_mode"
	ReadOptionWithName                 ReadOption = "with_name"
	ReadOptionWithValues               ReadOption = "with_values"
	ReadOptionWithThreshold            ReadOption = "with_threshold"
	ReadOptionWithReaction             ReadOption = "with_reaction"
	ReadOptionWithEnumeratedParameters ReadOption = "with_enumerated_parameters"
	ReadOptionWithArbitraryConditions  ReadOption = "with_arbitrary_conditions"
)

type CreateOption string

const (
	CreateOptionWithAction CreateOption = "with_action"
)

// ConvertToStringSlice converts []interface{} to []string, skipping nils.
func ConvertToStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		if v == nil {
			continue
		}
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprintf("%v", v)
		}
		result = append(result, s)
	}
	return result
}
