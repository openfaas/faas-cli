module github.com/openfaas/faas-cli

go 1.13

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190717161051-705d9623b7c1

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/alexellis/go-execute v0.0.0-20191207085904-961405ea7544
	github.com/alexellis/hmac v0.0.0-20180624210714-d5d71edd7bc7
	github.com/docker/docker v1.13.1
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/drone/envsubst v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0
	github.com/openfaas/faas v0.0.0-20191128202628-4d4ecc6bbf98
	github.com/openfaas/faas-provider v0.0.0-20191005090653-478f741b64cb
	github.com/pkg/errors v0.8.1
	github.com/ryanuber/go-glob v1.0.0
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	go.uber.org/goleak v1.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.7
	gotest.tools v2.2.0+incompatible // indirect
)
