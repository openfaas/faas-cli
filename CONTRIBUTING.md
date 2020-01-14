## Contributing

### License

This project is licensed under the MIT License.

## Guidelines apply from main OpenFaaS repo

See guide for [FaaS](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md) here.

## Unit testing with Golang

Please follow style guide on [this blog post](https://blog.alexellis.io/golang-writing-unit-tests/) from [The Go Programming Language](https://www.amazon.co.uk/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)

# Hacking on the faas-cli

## Installation / pre-requirements

* Docker

Install Docker because it is used to build Docker images if you create new functions.

* OpenFaaS - deployed and live

This CLI can build and deploy templated functions, so it's best if you have FaaS started up on your laptop. Head over to http://github.com/openfaas/faas/ and get up and running with a sample stack in 60 seconds.

* Golang

> Here's how to install Go in 60 seconds.

* Download Go 1.13 from https://golang.org/dl/

Then after installing run this command or place it in your `$HOME/.bash_profile`

```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

* Now clone / build `faas-cli`:

```bash
Migrate all imports/references to OpenFaaS org
$ mkdir -p $GOPATH/src/github.com/openfaas/
$ cd $GOPATH/src/github.com/openfaas/
$ git clone https://github.com/openfaas/faas-cli
$ cd faas-cli
$ go build
```

* Build multi-arch binaries

To build the release binaries type in:

```
./build_redist.sh
```

This creates the faas-cli for Mac, Windows, Linux x64, Linux ARMHF and Linux ARM64.

* Get the vendoring tool called `dep`

```
$ go get -u github.com/golang/dep/cmd/dep
```

Use the tool if you add new dependencies or want to update the existing ones.

> See also: [dep docs](https://github.com/golang/dep)

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

### How to update the `scoop` manifest

The `scoop` manifest for the faas-cli is part of the official [scoop](https://github.com/lukesampson/scoop/blob/master/bucket/faas-cli.json) repo on Github. It needs to be updated for each subsequent release.

#### Simple version bumps

```
git clone https://github.com/lukesampson/scoop
cd scoop
./bin/checkver.ps1 faas-cli -u
```

Test the updated manifest
```
scoop install .\bucket\faas-cli.json
```

Create a new branch and commit the manifest `faas-cli.json`, then create a PR to update the manifest in Scoop repository

## Update the utility-script

The `get.sh` file is served from the [cli.openfaas.com](https://github.com/openfaas/cli.openfaas.com) repository. 

It's used when people install via `curl` and `cli.openfaas.com`. The updated file then has to be redeployed to the hosting server.

Please raise a PR for changes there.

## Developer DCO (re-iteration from referenced CONTRIBUTING guide)

### Sign your work

> Note: all of the commits in your PR/Patch must be signed-off.

The sign-off is a simple line at the end of the explanation for a patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.

* Please sign your commits with `git commit -s` so that commits are traceable.

## Making a new release of the CLI

### Create a GitHub release

1. Through GitHub releases page create a new release and increment the version number.
2. Mark the release as pre-release to prevent the download script picking up the version
3. Wait until the Travis build is completed (which will add binaries to the page if successful)

Finally if the binaries were added successfully you should un-mark the "pre-release" checkbox, the CLI will now be available from our download utility script.

See above for notes on Brew. At present the brew team are auto-releasing PRs to their database when we make releases.

Community packages:

* ~~Arch Linux PKGBUILD (see [rawkode](https://github.com/rawkode)) [unmaintained]~~
* ~~Chocolately (see [pkeuter](https://github.com/pkeuter) via [au-packages](https://github.com/openfaas-incubator/au-packages)) [unmaintained]~~

### Update CHANGELOG

Get the changelog tool (requires Ruby)

```
$ sudo gem install github_changelog_generator
```

Generate a personal access token in GitHub and use it to update the CHANGELOG.md file:

```
$ export CHANGELOG_GITHUB_TOKEN=TOKEN_VALUE
$ github_changelog_generator
```
