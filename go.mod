module github.com/openfaas/faas-cli

go 1.15

// replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190717161051-705d9623b7c1

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/alexellis/go-execute v0.0.0-20191207085904-961405ea7544
	github.com/alexellis/hmac v0.0.0-20180624210714-d5d71edd7bc7
	github.com/docker/docker v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/drone/envsubst v1.0.2
	github.com/google/go-cmp v0.4.0
	github.com/gotestyourself/gotestyourself v1.4.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0
	github.com/openfaas/faas-provider v0.18.5
	github.com/openfaas/faas/gateway v0.0.0-20210311210633-a6dbb4cd0285
	github.com/pkg/errors v0.9.1
	github.com/ryanuber/go-glob v1.0.0
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v1.4.0 // indirect
)
