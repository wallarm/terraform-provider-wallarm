package common

const (
	Path   = "path"
	Iequal = "iequal"
	Header = "header"
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
