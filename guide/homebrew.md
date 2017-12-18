## How to update the `brew` formula

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
