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
	github.com/gotestyourself/gotestyourself v1.4.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0
	github.com/openfaas/faas v0.0.0-20201210155854-272ae94b506c
	github.com/openfaas/faas-provider v0.16.1
	github.com/pkg/errors v0.8.1
	github.com/ryanuber/go-glob v1.0.0
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v1.4.0 // indirect
)
