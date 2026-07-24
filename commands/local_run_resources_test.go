package commands

import "testing"

func Test_toDockerMemory(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{name: "binary mebibytes", value: "256Mi", want: "268435456"},
		{name: "binary gibibytes", value: "1Gi", want: "1073741824"},
		{name: "binary kibibytes", value: "64Ki", want: "65536"},
		{name: "decimal megabytes", value: "512M", want: "512000000"},
		{name: "decimal gigabytes", value: "2G", want: "2000000000"},
		{name: "decimal kilobytes", value: "128k", want: "128000"},
		{name: "no suffix is bytes", value: "1048576", want: "1048576"},
		{name: "fractional quantity", value: "1.5Gi", want: "1610612736"},
		{name: "surrounding whitespace", value: " 256Mi ", want: "268435456"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toDockerMemory(tc.value)
			if err != nil {
				t.Fatalf("want no error, but got: %s", err)
			}

			if got != tc.want {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
		})
	}
}

func Test_toDockerMemory_invalid(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "not a number", value: "someMi"},
		{name: "unknown suffix", value: "256MB"},
		{name: "negative", value: "-256Mi"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := toDockerMemory(tc.value); err == nil {
				t.Fatalf("want an error for %q, but got none", tc.value)
			}
		})
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
		{name: "surrounding whitespace", value: " 100m ", want: "0.1"},
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
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "not a number", value: "somem"},
		{name: "negative", value: "-100m"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := toDockerCPU(tc.value); err == nil {
				t.Fatalf("want an error for %q, but got none", tc.value)
			}
		})
	}
}
