package commands

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func Test_parseMemoryBytes(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  int64
	}{
		{name: "binary mebibytes", value: "256Mi", want: 268435456},
		{name: "binary gibibytes", value: "1Gi", want: 1073741824},
		{name: "binary kibibytes", value: "64Ki", want: 65536},
		{name: "binary tebibytes", value: "1Ti", want: 1 << 40},
		{name: "binary pebibytes", value: "1Pi", want: 1 << 50},
		{name: "binary exbibytes", value: "1Ei", want: 1 << 60},
		{name: "decimal megabytes", value: "512M", want: 512000000},
		{name: "decimal gigabytes", value: "2G", want: 2000000000},
		{name: "decimal kilobytes", value: "128k", want: 128000},
		{name: "decimal terabytes", value: "1T", want: 1e12},
		{name: "decimal petabytes", value: "1P", want: 1e15},
		{name: "decimal exabytes", value: "1E", want: 1e18},
		{name: "no suffix is bytes", value: "1048576", want: 1048576},
		{name: "zero is unset", value: "0", want: 0},
		{name: "fractional quantity", value: "1.5Gi", want: 1610612736},
		{name: "surrounding whitespace", value: " 256Mi ", want: 268435456},
		{name: "inner whitespace", value: "256 Mi", want: 268435456},
		{name: "lowercase exponent is not a suffix", value: "1e3", want: 1000},
		{name: "uppercase exponent is not a suffix", value: "1E3", want: 1000},
		{name: "largest exbibytes that fit", value: "7Ei", want: 7 << 60},
		{name: "largest byte count that fits", value: "9223372036854775807", want: math.MaxInt64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseMemoryBytes(tc.value)
			if err != nil {
				t.Fatalf("want no error, but got: %s", err)
			}

			if got != tc.want {
				t.Fatalf("want %d, but got %d", tc.want, got)
			}
		})
	}
}

func Test_parseMemoryBytes_invalid(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		contains string
	}{
		{name: "empty", value: "", contains: "empty"},
		{name: "whitespace only", value: "   ", contains: "empty"},
		{name: "not a number", value: "someMi", contains: "Ki, Mi, Gi"},
		{name: "unknown suffix", value: "256MB", contains: "Ki, Mi, Gi"},
		{name: "lowercase suffix", value: "256mi", contains: "Ki, Mi, Gi"},
		{name: "millicores are for cpu", value: "100m", contains: "milli-units are only valid for cpu"},
		{name: "negative", value: "-256Mi", contains: "must not be negative"},
		{name: "negative bytes", value: "-1", contains: "must not be negative"},
		{name: "negative fraction", value: "-0.5Gi", contains: "must not be negative"},
		{name: "not a number", value: "NaN", contains: "give a number of bytes"},
		{name: "infinity", value: "Inf", contains: "give a number of bytes"},
		{name: "signed infinity", value: "+Inf", contains: "give a number of bytes"},
		{name: "lowercase infinity", value: "inf", contains: "give a number of bytes"},
		{name: "overflows int64 in binary units", value: "8Ei", contains: "must be under"},
		{name: "overflows int64 in decimal units", value: "10E", contains: "must be under"},
		{name: "overflows int64 as a fraction", value: "9223372036854775808.5", contains: "must be under"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseMemoryBytes(tc.value)
			if err == nil {
				t.Fatalf("want an error for %q, but got none", tc.value)
			}

			if !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("want error containing %q, but got: %s", tc.contains, err)
			}

			// The parser's own text is of no use to someone editing a
			// stack.yaml, the message has to name what is accepted instead.
			if strings.Contains(err.Error(), "strconv") {
				t.Fatalf("want no parser detail in the error, but got: %s", err)
			}
		})
	}
}

func Test_toDockerMemoryReservation(t *testing.T) {
	minimum := strconv.FormatInt(dockerMinMemoryReservation, 10)

	cases := []struct {
		name        string
		value       string
		want        string
		wantWarning bool
	}{
		{name: "above the minimum", value: "256Mi", want: "268435456"},
		{name: "exactly the minimum", value: "6Mi", want: minimum},
		{name: "zero means unset", value: "0", want: "0"},
		{name: "decimal megabyte is under the minimum", value: "1M", want: minimum, wantWarning: true},
		{name: "mebibytes under the minimum", value: "5Mi", want: minimum, wantWarning: true},
		{name: "one byte is under the minimum", value: "1", want: minimum, wantWarning: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, warning, err := toDockerMemoryReservation(tc.value)
			if err != nil {
				t.Fatalf("want no error, but got: %s", err)
			}

			if got != tc.want {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}

			if tc.wantWarning {
				if warning == "" {
					t.Fatalf("want a warning for %q, but got none", tc.value)
				}

				if !strings.Contains(warning, "6MB") {
					t.Fatalf("want the warning to name the minimum, but got: %s", warning)
				}
			} else if warning != "" {
				t.Fatalf("want no warning for %q, but got: %s", tc.value, warning)
			}
		})
	}
}

func Test_toDockerMemoryReservation_invalid(t *testing.T) {
	if _, _, err := toDockerMemoryReservation("256MB"); err == nil {
		t.Fatalf("want an error for an invalid quantity, but got none")
	}
}

func Test_toDockerCPU(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{name: "millicores", value: "100m", want: "0.1"},
		{name: "millicores under one tenth", value: "50m", want: "0.05"},
		{name: "millicores over one cpu", value: "1500m", want: "1.5"},
		{name: "whole cpus", value: "2", want: "2"},
		{name: "fractional cpus", value: "0.5", want: "0.5"},
		{name: "zero is unset", value: "0", want: "0"},
		{name: "surrounding whitespace", value: " 100m ", want: "0.1"},
		{name: "inner whitespace", value: "100 m", want: "0.1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toDockerCPU(tc.value)
			if err != nil {
				t.Fatalf("want no error, but got: %s", err)
			}

			if got != tc.want {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
		})
	}
}

func Test_toDockerCPU_invalid(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		contains string
	}{
		{name: "empty", value: "", contains: "empty"},
		{name: "not a number", value: "somem", contains: "give a number of cpus"},
		{name: "suffix only", value: "m", contains: `invalid cpu value "m"`},
		{name: "repeated suffix", value: "100mm", contains: "give a number of cpus"},
		{name: "negative", value: "-100m", contains: "must not be negative"},
		{name: "negative cpus", value: "-1", contains: "must not be negative"},
		{name: "not a number", value: "NaN", contains: "give a number of cpus"},
		{name: "infinity", value: "Inf", contains: "give a number of cpus"},
		{name: "signed infinity", value: "+Inf", contains: "give a number of cpus"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := toDockerCPU(tc.value)
			if err == nil {
				t.Fatalf("want an error for %q, but got none", tc.value)
			}

			if !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("want error containing %q, but got: %s", tc.contains, err)
			}

			if strings.Contains(err.Error(), "strconv") {
				t.Fatalf("want no parser detail in the error, but got: %s", err)
			}
		})
	}
}
