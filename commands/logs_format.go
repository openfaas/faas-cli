package commands

import (
	"encoding/json"
	"strings"

	"github.com/openfaas/faas-cli/flags"
	"github.com/openfaas/faas-provider/logs"
)

// LogFormatter is a function that converts a log message to a string based on the supplied options
type LogFormatter func(msg logs.Message, timeFormat string, includeName, includeInstance bool) string

// GetLogFormatter maps a formatter name to a LogFormatter method
func GetLogFormatter(name string) LogFormatter {
	switch name {
	case string(flags.JSONLogFormat):
		return JSONFormatMessage
	case string(flags.KeyValueLogFormat):
		return KeyValueFormatMessage
	default:
		return PlainFormatMessage
	}
}

// JSONFormatMessage is a JSON formatting for log messages, the options are ignored and the entire log
// message json serialized
func JSONFormatMessage(msg logs.Message, timeFormat string, includeName, includeInstance bool) string {
	// error really can't happen here because of how simple the msg object is
	b, _ := json.Marshal(msg)
	return string(b)
}

// KeyValueFormatMessage returns the message in the format "timestamp=<> name=<> instance=<> message=<>"
func KeyValueFormatMessage(msg logs.Message, timeFormat string, includeName, includeInstance bool) string {
	var b strings.Builder

	// note that WriteString's error is always nil and safe to ignore here
	if timeFormat != "" {
		b.WriteString("timestamp=\"")
		b.WriteString(msg.Timestamp.Format(timeFormat))
		b.WriteString("\" ")
	}

	if includeName {
		b.WriteString("name=\"")
		b.WriteString(msg.Name)
		b.WriteString("\" ")
	}

	if includeInstance {
		b.WriteString("instance=\"")
		b.WriteString(msg.Instance)
		b.WriteString("\" ")
	}

	b.WriteString("text=\"")
	b.WriteString(strings.TrimRight(msg.Text, "\n"))
	b.WriteString("\" ")

	return b.String()
}

// PlainFormatMessage formats a log message as "<timestamp> <name> (<instance>) <text>"
func PlainFormatMessage(msg logs.Message, timeFormat string, includeName, includeInstance bool) string {
	var b strings.Builder

	// note that WriteString's error is always nil and safe to ignore here
	if timeFormat != "" {
		b.WriteString(msg.Timestamp.Format(timeFormat))
		b.WriteString(" ")
	}

	if includeName {
		b.WriteString(msg.Name)
		b.WriteString(" ")
	}

	if includeInstance {
		b.WriteString("(")
		b.WriteString(msg.Instance)
		b.WriteString(")")
		b.WriteString(" ")
	}

	b.WriteString(msg.Text)

	return strings.TrimRight(b.String(), "\n")
}
