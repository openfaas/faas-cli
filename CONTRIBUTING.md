## Contributing

### License

This project is licensed under the MIT License.

## Guidelines apply from main OpenFaaS repo

See guide for [FaaS](https://github.com/alexellis/faas/blob/master/CONTRIBUTING.md) here.

## Hacking on the faas-cli

## Installation / pre-requirements

* Docker

Install Docker because it is used to build Docker images if you create new functions.

* FaaS - deployed and live

This CLI can build and deploy templated functions, so it's best if you have FaaS started up on your laptop. Head over to http://docs.get-faas.com/ and get up and running with a sample stack in 60 seconds.

* Golang

> Here's how to install Go in 60 seconds.

* Grab Go 1.7.x from https://golang.org/dl/

Then after installing run this command or place it in your `$HOME/.bash_profile`

```bash
export GOPATH=$HOME/go
```

* Now clone / build `faas-cli`:

```
$ mkdir -p $GOPATH/src/github.com/alexellis/
$ cd $GOPATH/src/github.com/alexellis/
$ git clone https://github.com/alexellis/faas-cli
$ cd faas-cli
$ go get -d -v
$ go build
```

### How to update the `brew` formula

The `brew` formula for the faas-cli is part of the official [homebrew-core](https://github.com/Homebrew/homebrew-core/blob/master/Formula/faas-cli.rb) repo on Github. It needs to be updated for each subsequent release.

#### Simple version bumps

If the only change required is a version bump, ie no new tests, or changes to existing tested functionality or build steps, the `brew bump-formula-pr` command can be used to do everything (i.e. forking, committing, pushing) required to bump the version.

For example (supplying both the new version tag and its associated Git sha-256).

```
brew bump-formula-pr --strict faas-cli --tag=<version> --revision=<sha-256>
```

#### Changes requiring new/update tests/build steps

If a new release alters behaviour tested in the Brew Formula, adds new testable behaviors or alters the build steps then you will need to manually raise a PR with an updated Formula, the guidelines for updating brew describe the process in more detail:

https://github.com/Homebrew/homebrew-core/blob/master/CONTRIBUTING.md

After `brew edit` run the build and test the results:

```
$ brew uninstall --force faas-cli ; \
  brew install --build-from-source faas-cli ; \
  brew test faas-cli ; \
  brew audit --strict faas-cli
```

## Update the utility-script

Please raise a PR for the get.sh file held in this repository. It's used when people install via `curl` and `cli.openfaas.com`. The updated file then has to be redeployed to the hosting server.

