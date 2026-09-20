package format

import "errors"

const (
	Flag = "format"
)

type Enum int8

const (
	Cli        Enum = 1
	Json       Enum = 2
	JsonPretty Enum = 3
)

const (
	CliString        = "cli"
	JsonString       = "json"
	JsonPrettyString = "json-pretty"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *Enum) String() string {
	switch *e {
	case Cli:
		return CliString
	case Json:
		return JsonString
	case JsonPretty:
		return JsonPrettyString
	}
	return ""
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *Enum) Set(v string) error {
	switch v {
	case CliString:
		*e = Cli
	case JsonString:
		*e = Json
	case JsonPrettyString:
		*e = JsonPretty
	default:
		return errors.New(`must be one of "` + CliString + `", "` + JsonString + `", "` + JsonPrettyString + `"`)
	}
	return nil
}

// Type is only used in help text
func (e *Enum) Type() string { return "string" }
