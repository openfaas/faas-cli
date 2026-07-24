package commands

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/openfaas/go-sdk/stack"
)

func Test_buildDockerRun_limits(t *testing.T) {
	cases := []struct {
		name     string
		limits   *stack.FunctionResources
		wantArgs []string
		omitArgs []string
	}{
		{
			name:     "no limits",
			limits:   nil,
			omitArgs: []string{"--memory-reservation=", "--cpus="},
		},
		{
			name:     "kubernetes quantities are converted for docker",
			limits:   &stack.FunctionResources{Memory: "256Mi", CPU: "100m"},
			wantArgs: []string{"--memory-reservation=268435456", "--cpus=0.1"},
		},
		{
			name:     "decimal memory and whole cpus",
			limits:   &stack.FunctionResources{Memory: "512M", CPU: "2"},
			wantArgs: []string{"--memory-reservation=512000000", "--cpus=2"},
		},
		{
			name:     "memory under docker's minimum is raised",
			limits:   &stack.FunctionResources{Memory: "5Mi"},
			wantArgs: []string{"--memory-reservation=6291456"},
		},
		{
			name:     "only memory set",
			limits:   &stack.FunctionResources{Memory: "1Gi"},
			wantArgs: []string{"--memory-reservation=1073741824"},
			omitArgs: []string{"--cpus="},
		},
		{
			name:     "only cpu set",
			limits:   &stack.FunctionResources{CPU: "1500m"},
			wantArgs: []string{"--cpus=1.5"},
			omitArgs: []string{"--memory-reservation="},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tagFormat = 0

			fnc := stack.Function{
				Name:     "fn",
				Image:    "alexellis/fn:latest",
				FProcess: "./handler",
				Limits:   tc.limits,
			}

			cmd, err := buildDockerRun(context.Background(), "fn", fnc, runOptions{port: 8080})
			if err != nil {
				t.Fatalf("want no error, but got: %s", err)
			}

			for _, want := range tc.wantArgs {
				if !slices.Contains(cmd.Args, want) {
					t.Fatalf("want %q in: %v", want, cmd.Args)
				}
			}

			for _, prefix := range tc.omitArgs {
				for _, arg := range cmd.Args {
					if strings.HasPrefix(arg, prefix) {
						t.Fatalf("want no %q argument, but got %q", prefix, arg)
					}
				}
			}
		})
	}
}

func Test_buildDockerRun_invalidLimits(t *testing.T) {
	cases := []struct {
		name   string
		limits *stack.FunctionResources
	}{
		{name: "unknown memory suffix", limits: &stack.FunctionResources{Memory: "256MB"}},
		{name: "memory too large for docker", limits: &stack.FunctionResources{Memory: "8Ei"}},
		{name: "memory is not a number", limits: &stack.FunctionResources{Memory: "NaN"}},
		{name: "cpu is not a number", limits: &stack.FunctionResources{CPU: "Inf"}},
		{name: "negative cpu", limits: &stack.FunctionResources{CPU: "-1"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tagFormat = 0

			fnc := stack.Function{
				Name:     "fn",
				Image:    "alexellis/fn:latest",
				FProcess: "./handler",
				Limits:   tc.limits,
			}

			if _, err := buildDockerRun(context.Background(), "fn", fnc, runOptions{port: 8080}); err == nil {
				t.Fatalf("want an error for %v, but got none", tc.limits)
			}
		})
	}
}
