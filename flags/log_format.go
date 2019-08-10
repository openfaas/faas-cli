package flags

import (
	"fmt"
	"strings"
)

// LogFormat determines the output format of the log stream
type LogFormat string

const PlainLogFormat LogFormat = "plain"
const KeyValueLogFormat LogFormat = "keyvalue"
const JSONLogFormat LogFormat = "json"

// Type implements pflag.Value
func (l *LogFormat) Type() string {
	return "logformat"
}

// String implements Stringer
func (l *LogFormat) String() string {
	if l == nil {
		return ""
	}
	return string(*l)
}

// Set implements pflag.Value
func (l *LogFormat) Set(value string) error {
	switch strings.ToLower(value) {
	case "plain", "keyvalue", "json":
		*l = LogFormat(value)
	default:
		return fmt.Errorf("unknown log format: '%s'", value)
	}
	return nil
}
