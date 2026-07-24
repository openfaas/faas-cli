package commands

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Kubernetes memory suffixes, mapped to their multiplier in bytes. Both the
// binary (Ki, Mi, ...) and decimal (k, M, ...) forms are valid in a function's
// limits/requests, but Docker understands neither.
//
// The binary forms are listed first so that "1Ei" matches "Ei" and not "i",
// and so that the single letter forms are only reached once the two letter
// ones have been ruled out.
var memorySuffixes = []struct {
	suffix     string
	multiplier int64
}{
	{"Ki", 1 << 10},
	{"Mi", 1 << 20},
	{"Gi", 1 << 30},
	{"Ti", 1 << 40},
	{"Pi", 1 << 50},
	{"Ei", 1 << 60},
	{"k", 1e3},
	{"M", 1e6},
	{"G", 1e9},
	{"T", 1e12},
	{"P", 1e15},
	{"E", 1e18},
}

// memorySuffixHelp lists the suffixes accepted in a memory quantity, for use
// in error messages. "256MB" and "256mb" are common typos for "256M"/"256Mi".
const memorySuffixHelp = `Ki, Mi, Gi, Ti, Pi, Ei, k, M, G, T, P or E`

// dockerMinMemoryReservation is the lowest value Docker accepts for
// --memory-reservation. Anything between one byte and this floor is refused
// by the daemon with "Minimum memory reservation allowed is 6MB".
const dockerMinMemoryReservation int64 = 6 * 1024 * 1024

// toDockerMemoryReservation converts a Kubernetes memory quantity such as
// "256Mi" or "1G" into a value for Docker's --memory-reservation flag, which is
// a plain byte count. Docker's own suffixes (b, k, m, g) do not overlap cleanly
// with Kubernetes', so bytes avoid any ambiguity.
//
// Docker refuses any reservation below 6MB, so a smaller limit is raised to the
// minimum and the caller is given a warning to print. A limit of zero means
// "unset" and is left alone.
func toDockerMemoryReservation(value string) (reservation string, warning string, err error) {
	bytes, err := parseMemoryBytes(value)
	if err != nil {
		return "", "", err
	}

	if bytes > 0 && bytes < dockerMinMemoryReservation {
		warning = fmt.Sprintf("Warning: memory limit of %q is under Docker's minimum reservation of 6MB, using 6MB instead", strings.TrimSpace(value))
		bytes = dockerMinMemoryReservation
	}

	return strconv.FormatInt(bytes, 10), warning, nil
}

// parseMemoryBytes resolves a Kubernetes memory quantity to a count of bytes.
func parseMemoryBytes(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("memory value is empty")
	}

	digits := trimmed
	var multiplier int64 = 1

	for _, s := range memorySuffixes {
		if rest, found := strings.CutSuffix(trimmed, s.suffix); found {
			digits = rest
			multiplier = s.multiplier
			break
		}
	}

	digits = strings.TrimSpace(digits)

	// Whole numbers are multiplied as integers so that a quantity too large
	// for float64 to hold exactly is still converted precisely.
	if quantity, err := strconv.ParseInt(digits, 10, 64); err == nil {
		if quantity < 0 {
			return 0, negativeMemoryError(value)
		}

		if quantity > math.MaxInt64/multiplier {
			return 0, oversizedMemoryError(value)
		}

		return quantity * multiplier, nil
	}

	quantity, err := strconv.ParseFloat(digits, 64)
	if err != nil {
		return 0, invalidMemoryError(value)
	}

	// ParseFloat accepts "NaN", "Inf" and "+Inf", none of which are a size,
	// and a NaN slips past the check for a negative quantity below.
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return 0, invalidMemoryError(value)
	}

	if quantity < 0 {
		return 0, negativeMemoryError(value)
	}

	// float64 cannot represent math.MaxInt64, the nearest value it holds is
	// one larger, so anything at or above it overflows the conversion.
	bytes := math.Round(quantity * float64(multiplier))
	if bytes >= float64(math.MaxInt64) {
		return 0, oversizedMemoryError(value)
	}

	return int64(bytes), nil
}

func invalidMemoryError(value string) error {
	if suffix := strings.TrimSpace(value); strings.HasSuffix(suffix, "m") {
		return fmt.Errorf("invalid memory value %q: milli-units are only valid for cpu, give a number of bytes with an optional suffix of: %s", value, memorySuffixHelp)
	}

	return fmt.Errorf("invalid memory value %q: give a number of bytes with an optional suffix of: %s", value, memorySuffixHelp)
}

func negativeMemoryError(value string) error {
	return fmt.Errorf("invalid memory value %q: must not be negative", value)
}

func oversizedMemoryError(value string) error {
	// Typed as int64 so the constant does not take its default type of int,
	// which would overflow when building for a 32-bit target.
	return fmt.Errorf("invalid memory value %q: must be under %d bytes", value, int64(math.MaxInt64))
}

// toDockerCPU converts a Kubernetes CPU quantity such as "100m" (millicores)
// into the decimal number of CPUs that Docker's --cpus flag accepts. Values
// without a suffix are already whole or fractional CPUs and pass through.
func toDockerCPU(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("cpu value is empty")
	}

	digits, milli := strings.CutSuffix(trimmed, "m")

	cpus, err := strconv.ParseFloat(strings.TrimSpace(digits), 64)
	if err != nil {
		return "", invalidCPUError(value)
	}

	// ParseFloat accepts "NaN", "Inf" and "+Inf", none of which are a number
	// of CPUs, and a NaN slips past the check for a negative quantity below.
	if math.IsNaN(cpus) || math.IsInf(cpus, 0) {
		return "", invalidCPUError(value)
	}

	if milli {
		cpus = cpus / 1000
	}

	if cpus < 0 {
		return "", fmt.Errorf("invalid cpu value %q: must not be negative", value)
	}

	return strconv.FormatFloat(cpus, 'f', -1, 64), nil
}

func invalidCPUError(value string) error {
	return fmt.Errorf("invalid cpu value %q: give a number of cpus such as \"1\" or \"0.5\", or millicores such as \"500m\"", value)
}
