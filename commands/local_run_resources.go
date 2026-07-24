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
var memorySuffixes = []struct {
	suffix     string
	multiplier float64
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

// toDockerMemory converts a Kubernetes memory quantity such as "256Mi" or "1G"
// into a plain byte count, which is what Docker's --memory and
// --memory-reservation flags accept. Docker's own suffixes (b, k, m, g) do not
// overlap cleanly with Kubernetes', so bytes avoid any ambiguity.
func toDockerMemory(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("memory value is empty")
	}

	for _, s := range memorySuffixes {
		digits, found := strings.CutSuffix(trimmed, s.suffix)
		if !found {
			continue
		}

		quantity, err := strconv.ParseFloat(strings.TrimSpace(digits), 64)
		if err != nil {
			return "", fmt.Errorf("invalid memory value %q: %w", value, err)
		}

		if quantity < 0 {
			return "", fmt.Errorf("invalid memory value %q: must not be negative", value)
		}

		return strconv.FormatInt(int64(math.Round(quantity*s.multiplier)), 10), nil
	}

	// No suffix, so the value is already a count of bytes.
	bytes, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return "", fmt.Errorf("invalid memory value %q: %w", value, err)
	}

	if bytes < 0 {
		return "", fmt.Errorf("invalid memory value %q: must not be negative", value)
	}

	return strconv.FormatInt(int64(math.Round(bytes)), 10), nil
}

// toDockerCPU converts a Kubernetes CPU quantity such as "100m" (millicores)
// into the decimal number of CPUs that Docker's --cpus flag accepts. Values
// without a suffix are already whole or fractional CPUs and pass through.
func toDockerCPU(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("cpu value is empty")
	}

	cpus, err := strconv.ParseFloat(strings.TrimSuffix(trimmed, "m"), 64)
	if err != nil {
		return "", fmt.Errorf("invalid cpu value %q: %w", value, err)
	}

	if strings.HasSuffix(trimmed, "m") {
		cpus = cpus / 1000
	}

	if cpus < 0 {
		return "", fmt.Errorf("invalid cpu value %q: must not be negative", value)
	}

	return strconv.FormatFloat(cpus, 'f', -1, 64), nil
}
