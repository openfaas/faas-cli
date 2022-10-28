package util

import (
	"fmt"
	"strings"
)

func ParseMap(envvars []string, keyName string) (map[string]string, error) {
	result := make(map[string]string)
	for _, envvar := range envvars {
		s := strings.SplitN(strings.TrimSpace(envvar), "=", 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("label format is not correct, needs key=value")
		}
		envvarName := s[0]
		envvarValue := s[1]

		if !(len(envvarName) > 0) {
			return nil, fmt.Errorf("empty %s name: [%s]", keyName, envvar)
		}
		if !(len(envvarValue) > 0) {
			return nil, fmt.Errorf("empty %s value: [%s]", keyName, envvar)
		}

		result[envvarName] = envvarValue
	}
	return result, nil
}

// util.MergeMap merges two maps, with the overlay taking precedence.
// The return value allocates a new map.
func MergeMap(base map[string]string, overlay map[string]string) map[string]string {
	merged := make(map[string]string)

	for k, v := range base {
		merged[k] = v
	}
	for k, v := range overlay {
		merged[k] = v
	}

	return merged
}

func MergeSlice(values []string, overlay []string) []string {
	results := []string{}
	added := make(map[string]bool)
	for _, value := range overlay {
		results = append(results, value)
		added[value] = true
	}

	for _, value := range values {
		if exists := added[value]; !exists {
			results = append(results, value)
		}
	}

	return results
}
