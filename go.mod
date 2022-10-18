module github.com/openfaas/faas-cli

go 1.17

// replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190717161051-705d9623b7c1

// replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

require (
	github.com/alexellis/arkade v0.0.0-20221017065732-2e42683df1e6
	github.com/alexellis/go-execute v0.5.0
	github.com/alexellis/hmac v1.3.0
	github.com/docker/docker v20.10.17+incompatible
	github.com/drone/envsubst v1.0.3
	github.com/google/go-cmp v0.5.9
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0
	github.com/openfaas/faas-provider v0.19.1
	github.com/openfaas/faas/gateway v0.0.0-20221013075423-32b828b25e1c
	github.com/pkg/errors v0.9.1
	github.com/ryanuber/go-glob v1.0.0
	github.com/spf13/cobra v1.6.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/cheggaaa/pb/v3 v3.1.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20220823224334-20c2bfdbfe24 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gotest.tools/v3 v3.0.3 // indirect
)
