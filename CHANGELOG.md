# Change Log

## [Unreleased](https://github.com/openfaas/faas-cli/tree/HEAD)

[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.31...HEAD)

**Closed issues:**

- `faas-cli deploy` needs to reset constraints array after each function deploy in stack YAML [\#218](https://github.com/openfaas/faas-cli/issues/218)

## [0.4.31](https://github.com/openfaas/faas-cli/tree/0.4.31) (2017-11-11)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.3b...0.4.31)

**Implemented enhancements:**

- Proposal: Set default gateway value for executable [\#196](https://github.com/openfaas/faas-cli/issues/196)
- Version logo colour clashes with Powershell blue background [\#149](https://github.com/openfaas/faas-cli/issues/149)
- Proposal: Default to a 'well-known' named stack file [\#71](https://github.com/openfaas/faas-cli/issues/71)

**Fixed bugs:**

- `faas-cli new --help` shows a --name flag  [\#151](https://github.com/openfaas/faas-cli/issues/151)

**Closed issues:**

- Separate version from command version [\#208](https://github.com/openfaas/faas-cli/issues/208)
- Versioning [\#204](https://github.com/openfaas/faas-cli/issues/204)
- Add `--all` option to `faas-cli rm` to quickly remove all functions [\#203](https://github.com/openfaas/faas-cli/issues/203)
- Use strict parsing for --gateway flag [\#185](https://github.com/openfaas/faas-cli/issues/185)
- Proposal Enable basic auth for CLI [\#178](https://github.com/openfaas/faas-cli/issues/178)
- python3 template proposal [\#163](https://github.com/openfaas/faas-cli/issues/163)
- Create and manage .gitignore file for function templates [\#127](https://github.com/openfaas/faas-cli/issues/127)
- Add tests for a dummy endpoint [\#45](https://github.com/openfaas/faas-cli/issues/45)

**Merged pull requests:**

- Don't re-use constraints for each function [\#219](https://github.com/openfaas/faas-cli/pull/219) ([alexellis](https://github.com/alexellis))
- Command Version Separation [\#209](https://github.com/openfaas/faas-cli/pull/209) ([Lewiscowles1986](https://github.com/Lewiscowles1986))
- build\_test: Update the Copyright [\#206](https://github.com/openfaas/faas-cli/pull/206) ([nenadilic84](https://github.com/nenadilic84))
- Scan for license compliance during build [\#205](https://github.com/openfaas/faas-cli/pull/205) ([alexellis](https://github.com/alexellis))

## [0.4.3b](https://github.com/openfaas/faas-cli/tree/0.4.3b) (2017-11-04)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.30...0.4.3b)

## [0.4.30](https://github.com/openfaas/faas-cli/tree/0.4.30) (2017-11-04)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.3a...0.4.30)

## [0.4.3a](https://github.com/openfaas/faas-cli/tree/0.4.3a) (2017-11-03)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.3...0.4.3a)

## [0.4.3](https://github.com/openfaas/faas-cli/tree/0.4.3) (2017-11-03)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.20...0.4.3)

**Closed issues:**

- derek test this [\#193](https://github.com/openfaas/faas-cli/issues/193)
- Dependency resolution within template Dockerfiles [\#191](https://github.com/openfaas/faas-cli/issues/191)

**Merged pull requests:**

- Restore Stefan as author of commits in Git [\#202](https://github.com/openfaas/faas-cli/pull/202) ([alexellis](https://github.com/alexellis))
- Update tests for invalid --gateway URL to use got/want in error message [\#199](https://github.com/openfaas/faas-cli/pull/199) ([ericstoekl](https://github.com/ericstoekl))
- Add error handling for calls to http.NewRequest\(\) in package proxy [\#198](https://github.com/openfaas/faas-cli/pull/198) ([ericstoekl](https://github.com/ericstoekl))
- Enable non-zero-exit in case of an error while running "build" [\#195](https://github.com/openfaas/faas-cli/pull/195) ([nenadilic84](https://github.com/nenadilic84))
- Update deletion error handling [\#192](https://github.com/openfaas/faas-cli/pull/192) ([alexellis](https://github.com/alexellis))
- \[Testing\] Support Basic Auth for multiple gateways [\#182](https://github.com/openfaas/faas-cli/pull/182) ([stefanprodan](https://github.com/stefanprodan))

## [0.4.20](https://github.com/openfaas/faas-cli/tree/0.4.20) (2017-10-27)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.19c...0.4.20)

**Fixed bugs:**

- The CLI is setting fprocess to `node index.js` on deploy by default, ignoring the value set in images [\#169](https://github.com/openfaas/faas-cli/issues/169)

**Closed issues:**

- faas-cli build to select architecture automatically [\#176](https://github.com/openfaas/faas-cli/issues/176)
- curl / faas-cli etc hangs with IPv6 entry for localhost [\#164](https://github.com/openfaas/faas-cli/issues/164)

## [0.4.19c](https://github.com/openfaas/faas-cli/tree/0.4.19c) (2017-10-26)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.19b...0.4.19c)

## [0.4.19b](https://github.com/openfaas/faas-cli/tree/0.4.19b) (2017-10-26)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.19...0.4.19b)

**Merged pull requests:**

- Fix permissions for Node.js [\#188](https://github.com/openfaas/faas-cli/pull/188) ([alexellis](https://github.com/alexellis))

## [0.4.19](https://github.com/openfaas/faas-cli/tree/0.4.19) (2017-10-23)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18d...0.4.19)

**Fixed bugs:**

- Homebrew test block fails with 0.4.18c [\#175](https://github.com/openfaas/faas-cli/issues/175)

**Merged pull requests:**

- Add label support [\#184](https://github.com/openfaas/faas-cli/pull/184) ([alexellis](https://github.com/alexellis))
- Mile high commit - adds --query flag to CLI, but not to YAML [\#173](https://github.com/openfaas/faas-cli/pull/173) ([alexellis](https://github.com/alexellis))
- Add timeouts for HTTP clients [\#165](https://github.com/openfaas/faas-cli/pull/165) ([alexellis](https://github.com/alexellis))
- \[WIP\] non-root user for NodeJS template [\#83](https://github.com/openfaas/faas-cli/pull/83) ([austinfrey](https://github.com/austinfrey))

## [0.4.18d](https://github.com/openfaas/faas-cli/tree/0.4.18d) (2017-10-16)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18c...0.4.18d)

**Merged pull requests:**

- Add test coverage to gateway URL overriding [\#177](https://github.com/openfaas/faas-cli/pull/177) ([alexellis](https://github.com/alexellis))

## [0.4.18c](https://github.com/openfaas/faas-cli/tree/0.4.18c) (2017-10-16)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18b...0.4.18c)

**Merged pull requests:**

- Fixes --gateway override which was ignored when using YAML [\#174](https://github.com/openfaas/faas-cli/pull/174) ([alexellis](https://github.com/alexellis))
- Remove language default and fix command flags [\#170](https://github.com/openfaas/faas-cli/pull/170) ([johnmccabe](https://github.com/johnmccabe))

## [0.4.18b](https://github.com/openfaas/faas-cli/tree/0.4.18b) (2017-10-12)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18a...0.4.18b)

## [0.4.18a](https://github.com/openfaas/faas-cli/tree/0.4.18a) (2017-10-12)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18...0.4.18a)

**Closed issues:**

- faas-cli deploy  -f sample.yml failing to start [\#156](https://github.com/openfaas/faas-cli/issues/156)
- faas-cli: not deploying to my pi swarm [\#148](https://github.com/openfaas/faas-cli/issues/148)
- Golang Template Dockerfile tests Vendor Package [\#146](https://github.com/openfaas/faas-cli/issues/146)
- Fail "build" if Docker CLI throws an error [\#138](https://github.com/openfaas/faas-cli/issues/138)
- Proposal: Add template for golang functions [\#79](https://github.com/openfaas/faas-cli/issues/79)

**Merged pull requests:**

- Add named secrets [\#161](https://github.com/openfaas/faas-cli/pull/161) ([alexellis](https://github.com/alexellis))
- Golang ARMHF template [\#160](https://github.com/openfaas/faas-cli/pull/160) ([alexellis](https://github.com/alexellis))
- Add missing deps to vendor.conf [\#158](https://github.com/openfaas/faas-cli/pull/158) ([johnmccabe](https://github.com/johnmccabe))
- Fix environment file example [\#157](https://github.com/openfaas/faas-cli/pull/157) ([developius](https://github.com/developius))
- Replace console.log with process.stdout.write for Node templates [\#153](https://github.com/openfaas/faas-cli/pull/153) ([developius](https://github.com/developius))
- Update asciinema demo [\#152](https://github.com/openfaas/faas-cli/pull/152) ([developius](https://github.com/developius))
- Use green logo on windows [\#150](https://github.com/openfaas/faas-cli/pull/150) ([johnmccabe](https://github.com/johnmccabe))
- Amend golang template Dockerfile so that go test ignores vendor fixes \#146 [\#147](https://github.com/openfaas/faas-cli/pull/147) ([rgee0](https://github.com/rgee0))

## [0.4.18](https://github.com/openfaas/faas-cli/tree/0.4.18) (2017-10-06)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18-alpha1...0.4.18)

## [0.4.18-alpha1](https://github.com/openfaas/faas-cli/tree/0.4.18-alpha1) (2017-10-06)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.18-alpha...0.4.18-alpha1)

**Closed issues:**

- Error deploying faas-cli [\#144](https://github.com/openfaas/faas-cli/issues/144)

**Merged pull requests:**

- Add Go template capability for creating functions [\#145](https://github.com/openfaas/faas-cli/pull/145) ([nicholasjackson](https://github.com/nicholasjackson))
- Default to stack.yml when no --yaml flag given [\#126](https://github.com/openfaas/faas-cli/pull/126) ([nicholasjackson](https://github.com/nicholasjackson))

## [0.4.18-alpha](https://github.com/openfaas/faas-cli/tree/0.4.18-alpha) (2017-10-05)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.17...0.4.18-alpha)

**Implemented enhancements:**

- Print error message when --filter or --regex returns no matches [\#124](https://github.com/openfaas/faas-cli/issues/124)

**Closed issues:**

- How do I monitor resource utilization? [\#140](https://github.com/openfaas/faas-cli/issues/140)
- Hacktoberfest: Alter verb invoke to take positional argument [\#134](https://github.com/openfaas/faas-cli/issues/134)
- Add colour to "faas-cli new" command [\#129](https://github.com/openfaas/faas-cli/issues/129)
- Use multi-stage build and dotnet publish for C\# [\#53](https://github.com/openfaas/faas-cli/issues/53)

**Merged pull requests:**

- Error checking for cli exec-commands [\#139](https://github.com/openfaas/faas-cli/pull/139) ([shorsher](https://github.com/shorsher))
- Update .gitignore when generating templates [\#137](https://github.com/openfaas/faas-cli/pull/137) ([viveksyngh](https://github.com/viveksyngh))
- Alter-invoke-verb [\#135](https://github.com/openfaas/faas-cli/pull/135) ([gardlt](https://github.com/gardlt))

## [0.4.17](https://github.com/openfaas/faas-cli/tree/0.4.17) (2017-10-01)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.16...0.4.17)

**Closed issues:**

- Hacktoberfest - move from --name parameter to positional arg [\#133](https://github.com/openfaas/faas-cli/issues/133)

**Merged pull requests:**

- Change to positional name for `new` function [\#132](https://github.com/openfaas/faas-cli/pull/132) ([alexellis](https://github.com/alexellis))

## [0.4.16](https://github.com/openfaas/faas-cli/tree/0.4.16) (2017-09-29)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.16-alpha...0.4.16)

**Closed issues:**

- Proposal - Include `ls` alias for faas-cli list [\#122](https://github.com/openfaas/faas-cli/issues/122)
- Canonical repository URL does not match Go import [\#116](https://github.com/openfaas/faas-cli/issues/116)
- faas-cli push made it so that I couldn't update my container [\#114](https://github.com/openfaas/faas-cli/issues/114)

**Merged pull requests:**

- Push in parallel [\#142](https://github.com/openfaas/faas-cli/pull/142) ([alexellis](https://github.com/alexellis))
- Print error message when no matches are found by --filter or --regex flag [\#125](https://github.com/openfaas/faas-cli/pull/125) ([ericstoekl](https://github.com/ericstoekl))
- Use PUT verb for function updates and update messages [\#123](https://github.com/openfaas/faas-cli/pull/123) ([johnmccabe](https://github.com/johnmccabe))

## [0.4.16-alpha](https://github.com/openfaas/faas-cli/tree/0.4.16-alpha) (2017-09-25)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.15a...0.4.16-alpha)

**Merged pull requests:**

- Updated contributing.md with new openfaas path [\#119](https://github.com/openfaas/faas-cli/pull/119) ([tripdubroot](https://github.com/tripdubroot))
- Add `ls` alias for `faas-cli list` command [\#118](https://github.com/openfaas/faas-cli/pull/118) ([ericstoekl](https://github.com/ericstoekl))
- Migrate all imports/references to OpenFaaS org [\#117](https://github.com/openfaas/faas-cli/pull/117) ([johnmccabe](https://github.com/johnmccabe))

## [0.4.15a](https://github.com/openfaas/faas-cli/tree/0.4.15a) (2017-09-24)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.15...0.4.15a)

**Merged pull requests:**

- Add update flag for existing deployments [\#111](https://github.com/openfaas/faas-cli/pull/111) ([alexellis](https://github.com/alexellis))

## [0.4.15](https://github.com/openfaas/faas-cli/tree/0.4.15) (2017-09-23)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.15-alpha1...0.4.15)

**Merged pull requests:**

- Allow environment\_file in YAML files [\#110](https://github.com/openfaas/faas-cli/pull/110) ([alexellis](https://github.com/alexellis))

## [0.4.15-alpha1](https://github.com/openfaas/faas-cli/tree/0.4.15-alpha1) (2017-09-22)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.15-alpha...0.4.15-alpha1)

## [0.4.15-alpha](https://github.com/openfaas/faas-cli/tree/0.4.15-alpha) (2017-09-22)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.14...0.4.15-alpha)

**Implemented enhancements:**

- Proposal - filter functions in stacks [\#60](https://github.com/openfaas/faas-cli/issues/60)

**Closed issues:**

- Proposal: Use constant to declare default gateway [\#101](https://github.com/openfaas/faas-cli/issues/101)
- Cli: open ./template/python: no such file or directory error [\#99](https://github.com/openfaas/faas-cli/issues/99)
- CLI can't successfully deploy to 'custom' named stack [\#96](https://github.com/openfaas/faas-cli/issues/96)
- Proposal - automate adding cli binaries to Github release on tag [\#67](https://github.com/openfaas/faas-cli/issues/67)
- Support overriding function's Dockerfile [\#15](https://github.com/openfaas/faas-cli/issues/15)

**Merged pull requests:**

- Parallel build functionality [\#113](https://github.com/openfaas/faas-cli/pull/113) ([alexellis](https://github.com/alexellis))
- Tests with mocked server & tests for commands and proxy directories [\#112](https://github.com/openfaas/faas-cli/pull/112) ([itscaro](https://github.com/itscaro))
- Support ARM64 for Packet [\#109](https://github.com/openfaas/faas-cli/pull/109) ([alexellis](https://github.com/alexellis))
- Add tag into version command and shorten SHA1 [\#104](https://github.com/openfaas/faas-cli/pull/104) ([itscaro](https://github.com/itscaro))

## [0.4.14](https://github.com/openfaas/faas-cli/tree/0.4.14) (2017-09-15)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.14-alpha...0.4.14)

**Closed issues:**

- Proposal: Improve version command [\#105](https://github.com/openfaas/faas-cli/issues/105)
- Proposal: Do a gofmt verification before build [\#95](https://github.com/openfaas/faas-cli/issues/95)
- Add Clojure as supported language [\#89](https://github.com/openfaas/faas-cli/issues/89)

**Merged pull requests:**

- Use constant for default gateway strings [\#102](https://github.com/openfaas/faas-cli/pull/102) ([itscaro](https://github.com/itscaro))
- Allow to override the network name from CLI [\#100](https://github.com/openfaas/faas-cli/pull/100) ([itscaro](https://github.com/itscaro))
- Make gofmt fail builds [\#93](https://github.com/openfaas/faas-cli/pull/93) ([itscaro](https://github.com/itscaro))

## [0.4.14-alpha](https://github.com/openfaas/faas-cli/tree/0.4.14-alpha) (2017-09-12)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.13-alpha...0.4.14-alpha)

**Fixed bugs:**

- build action template download/extraction fails on Windows [\#94](https://github.com/openfaas/faas-cli/issues/94)

**Closed issues:**

- Calling function created with `csharp` template results in error [\#90](https://github.com/openfaas/faas-cli/issues/90)
- publish function api to kong [\#88](https://github.com/openfaas/faas-cli/issues/88)

## [0.4.13-alpha](https://github.com/openfaas/faas-cli/tree/0.4.13-alpha) (2017-09-08)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.12...0.4.13-alpha)

**Implemented enhancements:**

- Proposal: Add 'invoke' command to the CLI [\#74](https://github.com/openfaas/faas-cli/issues/74)
- Proposal: Add init/new/create commands to the CLI [\#70](https://github.com/openfaas/faas-cli/issues/70)

**Merged pull requests:**

- \[WIP\] Use multi-stage build for CSharp [\#82](https://github.com/openfaas/faas-cli/pull/82) ([rorpage](https://github.com/rorpage))

## [0.4.12](https://github.com/openfaas/faas-cli/tree/0.4.12) (2017-09-05)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.11...0.4.12)

**Implemented enhancements:**

- Proposal: Add 'list' command to the CLI [\#73](https://github.com/openfaas/faas-cli/issues/73)

**Merged pull requests:**

- Added basic filter functionality with regex search in stack.ParseYAML [\#78](https://github.com/openfaas/faas-cli/pull/78) ([ericstoekl](https://github.com/ericstoekl))
- Assorted minor fixes while testing new faas-cli subcommands [\#77](https://github.com/openfaas/faas-cli/pull/77) ([johnmccabe](https://github.com/johnmccabe))

## [0.4.11](https://github.com/openfaas/faas-cli/tree/0.4.11) (2017-09-04)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.10...0.4.11)

**Merged pull requests:**

- `new` cmd [\#75](https://github.com/openfaas/faas-cli/pull/75) ([alexellis](https://github.com/alexellis))

## [0.4.10](https://github.com/openfaas/faas-cli/tree/0.4.10) (2017-09-04)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.9...0.4.10)

**Fixed bugs:**

- GitCommit in commands package not being updated at build time [\#62](https://github.com/openfaas/faas-cli/issues/62)
- Issue - function sub-folders are not copied recursively [\#59](https://github.com/openfaas/faas-cli/issues/59)

**Closed issues:**

- Proposal: auto-completion / Cobra CLI [\#17](https://github.com/openfaas/faas-cli/issues/17)

**Merged pull requests:**

- Build and deploy all platforms to Github on tag - fixes \#67 [\#68](https://github.com/openfaas/faas-cli/pull/68) ([johnmccabe](https://github.com/johnmccabe))
- Remove master.zip file after building function with faas-cli build [\#66](https://github.com/openfaas/faas-cli/pull/66) ([ericstoekl](https://github.com/ericstoekl))
- Fix importpath in Go linker -X ldflag - fixes \#62 [\#63](https://github.com/openfaas/faas-cli/pull/63) ([johnmccabe](https://github.com/johnmccabe))
- Added recusion to handler copy + debugPrint fn [\#61](https://github.com/openfaas/faas-cli/pull/61) ([rgee0](https://github.com/rgee0))

## [0.4.9](https://github.com/openfaas/faas-cli/tree/0.4.9) (2017-08-31)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.9-beta...0.4.9)

**Closed issues:**

- Update fwatchdog to 0.6.1 and update command help message [\#56](https://github.com/openfaas/faas-cli/issues/56)

## [0.4.9-beta](https://github.com/openfaas/faas-cli/tree/0.4.9-beta) (2017-08-27)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.9-alpha...0.4.9-beta)

## [0.4.9-alpha](https://github.com/openfaas/faas-cli/tree/0.4.9-alpha) (2017-08-27)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.8...0.4.9-alpha)

**Closed issues:**

- Should have - "Custom" template type [\#29](https://github.com/openfaas/faas-cli/issues/29)

**Merged pull requests:**

- Merge header notice from MIT to newly refactored files [\#58](https://github.com/openfaas/faas-cli/pull/58) ([alexellis](https://github.com/alexellis))
- Update fwatchdog to 0.6.1 and correct help message for "action" option [\#57](https://github.com/openfaas/faas-cli/pull/57) ([itscaro](https://github.com/itscaro))

## [0.4.8](https://github.com/openfaas/faas-cli/tree/0.4.8) (2017-08-27)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.7...0.4.8)

**Closed issues:**

- on faas-cli -action deploy: dial tcp 127.0.0.1:8080: getsockopt: connection refused [\#49](https://github.com/openfaas/faas-cli/issues/49)
- Docs using ./faas-cli instead of faas-cli [\#48](https://github.com/openfaas/faas-cli/issues/48)
- Refactor: app.go into separate packages/files [\#43](https://github.com/openfaas/faas-cli/issues/43)
- Enhancement - CLI should support function deletion [\#42](https://github.com/openfaas/faas-cli/issues/42)
- Bump Golang version to 1.8.x [\#40](https://github.com/openfaas/faas-cli/issues/40)

**Merged pull requests:**

- Allow Dockerfile language [\#52](https://github.com/openfaas/faas-cli/pull/52) ([alexellis](https://github.com/alexellis))
- Update to Golang 1.8.3 [\#51](https://github.com/openfaas/faas-cli/pull/51) ([alexellis](https://github.com/alexellis))
- Migrate CLI to Cobra and add initial bash completion : fixes \#17 [\#50](https://github.com/openfaas/faas-cli/pull/50) ([johnmccabe](https://github.com/johnmccabe))
- Refactor app.go into packages/files [\#44](https://github.com/openfaas/faas-cli/pull/44) ([alexellis](https://github.com/alexellis))
- Enable -delete as an action [\#41](https://github.com/openfaas/faas-cli/pull/41) ([alexellis](https://github.com/alexellis))
- Migrate dependency management from glide to vndr [\#39](https://github.com/openfaas/faas-cli/pull/39) ([johnmccabe](https://github.com/johnmccabe))
- Added syntax highlighting [\#38](https://github.com/openfaas/faas-cli/pull/38) ([morsik](https://github.com/morsik))
- Update get.sh to determine latest full release from Github [\#37](https://github.com/openfaas/faas-cli/pull/37) ([johnmccabe](https://github.com/johnmccabe))
- Add bump-formula-pr to Brew update steps [\#36](https://github.com/openfaas/faas-cli/pull/36) ([johnmccabe](https://github.com/johnmccabe))

## [0.4.7](https://github.com/openfaas/faas-cli/tree/0.4.7) (2017-08-16)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.5-b...0.4.7)

**Merged pull requests:**

- Add ARM dockerfiles [\#35](https://github.com/openfaas/faas-cli/pull/35) ([alexellis](https://github.com/alexellis))
- Enable CSharp template with .NET 2.0 standard [\#34](https://github.com/openfaas/faas-cli/pull/34) ([alexellis](https://github.com/alexellis))

## [0.4.5-b](https://github.com/openfaas/faas-cli/tree/0.4.5-b) (2017-08-10)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.6...0.4.5-b)

## [0.4.6](https://github.com/openfaas/faas-cli/tree/0.4.6) (2017-08-10)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4.5...0.4.6)

**Merged pull requests:**

- Restore pulling templates behaviour [\#31](https://github.com/openfaas/faas-cli/pull/31) ([alexellis](https://github.com/alexellis))
- Add Ruby template [\#26](https://github.com/openfaas/faas-cli/pull/26) ([alexellis](https://github.com/alexellis))

## [0.4.5](https://github.com/openfaas/faas-cli/tree/0.4.5) (2017-08-08)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.4...0.4.5)

**Closed issues:**

- Templating should use the --squash parameter for docker build [\#23](https://github.com/openfaas/faas-cli/issues/23)
- -action build should copy sub-folders into template [\#18](https://github.com/openfaas/faas-cli/issues/18)
- cli hanging on deploy command [\#16](https://github.com/openfaas/faas-cli/issues/16)
- faas-cli --help should give some fully worked examples [\#14](https://github.com/openfaas/faas-cli/issues/14)
- Idea - Utility script for installation from binary [\#11](https://github.com/openfaas/faas-cli/issues/11)

**Merged pull requests:**

- Update samples to have DNS-sanitized names [\#25](https://github.com/openfaas/faas-cli/pull/25) ([alexellis](https://github.com/alexellis))
- Added squash flag & associated fixes [\#24](https://github.com/openfaas/faas-cli/pull/24) ([rgee0](https://github.com/rgee0))
- Added copyTree function and replaced the template copyFiles calls [\#21](https://github.com/openfaas/faas-cli/pull/21) ([rgee0](https://github.com/rgee0))
- Adding instructions to install faas-cli [\#20](https://github.com/openfaas/faas-cli/pull/20) ([jmkhael](https://github.com/jmkhael))

## [0.4](https://github.com/openfaas/faas-cli/tree/0.4) (2017-06-23)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.3...0.4)

## [0.3](https://github.com/openfaas/faas-cli/tree/0.3) (2017-06-23)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.2...0.3)

## [0.2](https://github.com/openfaas/faas-cli/tree/0.2) (2017-06-01)
[Full Changelog](https://github.com/openfaas/faas-cli/compare/0.1-alpha...0.2)

**Merged pull requests:**

- Add -skipBuild flag for non-built images [\#10](https://github.com/openfaas/faas-cli/pull/10) ([alexellis](https://github.com/alexellis))
- Add 'push' action and -f for URLs [\#9](https://github.com/openfaas/faas-cli/pull/9) ([alexellis](https://github.com/alexellis))
- Merge samples master [\#7](https://github.com/openfaas/faas-cli/pull/7) ([alexellis](https://github.com/alexellis))
- Rename samples [\#6](https://github.com/openfaas/faas-cli/pull/6) ([alexellis](https://github.com/alexellis))
- Support YAML [\#5](https://github.com/openfaas/faas-cli/pull/5) ([alexellis](https://github.com/alexellis))
- Move from exec/bash to Go native copy/mkdir commands. [\#4](https://github.com/openfaas/faas-cli/pull/4) ([alexellis](https://github.com/alexellis))
- Support Python as a target language, Node as default. [\#1](https://github.com/openfaas/faas-cli/pull/1) ([alexellis](https://github.com/alexellis))

## [0.1-alpha](https://github.com/openfaas/faas-cli/tree/0.1-alpha) (2017-04-14)


\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*