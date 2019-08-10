package flags

import (
	"strings"
	"time"
)

// TimeFormat is a timestamp format string that also accepts the following RFC names as shortcuts
//
//  ANSIC       = "Mon Jan _2 15:04:05 2006"
// 	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
// 	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
// 	RFC822      = "02 Jan 06 15:04 MST"
// 	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
// 	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
// 	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
// 	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
// 	RFC3339     = "2006-01-02T15:04:05Z07:00"
// 	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
//
// Any string is accepted
type TimeFormat string

// Type implements pflag.Value
func (l *TimeFormat) Type() string {
	return "timeformat"
}

// String implements Stringer
func (l *TimeFormat) String() string {
	if l == nil {
		return ""
	}
	return string(*l)
}

// Set implements pflag.Value
func (l *TimeFormat) Set(value string) error {
	switch strings.ToLower(value) {
	case "ansic":
		*l = TimeFormat(time.ANSIC)
	case "unixdate":
		*l = TimeFormat(time.UnixDate)
	case "rubydate":
		*l = TimeFormat(time.RubyDate)
	case "rfc822":
		*l = TimeFormat(time.RFC822)
	case "rfc822z":
		*l = TimeFormat(time.RFC822Z)
	case "rfc850":
		*l = TimeFormat(time.RFC850)
	case "rfc1123":
		*l = TimeFormat(time.RFC1123)
	case "rfc1123z":
		*l = TimeFormat(time.RFC1123Z)
	case "rfc3339":
		*l = TimeFormat(time.RFC3339)
	case "rfc3339nano":
		*l = TimeFormat(time.RFC3339Nano)
	default:
		*l = TimeFormat(value)
	}
	return nil
}
