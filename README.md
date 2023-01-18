## faas-cli

[![Build Status](https://github.com/openfaas/faas-cli/workflows/build/badge.svg?branch=master)](https://github.com/openfaas/faas-cli/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas-cli)](https://goreportcard.com/report/github.com/openfaas/faas-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

faas-cli is the official CLI for [OpenFaaS](https://github.com/openfaas/faas)

Run a demo with `faas-cli --help`

The CLI can be used to build and deploy functions to [OpenFaaS](https://github.com/openfaas/faas). You can build OpenFaaS functions from a set of supported language templates (such as Node.js, Python, CSharp and Ruby). That means you just write a handler file such as (handler.py/handler.js) and the CLI does the rest to create a Docker image.

New user? See how it works: [Morning coffee with the faas-cli](https://blog.alexellis.io/quickstart-openfaas-cli/)
Already an OpenFaaS user? Try [5 tips and tricks for the OpenFaaS CLI](https://www.openfaas.com/blog/five-cli-tips/)

### Get started: Install the CLI

You can install the CLI with a `curl` utility script, `brew` or by downloading the binary from the releases page. Once installed you'll get the `faas-cli` command and `faas` alias.

Utility script with `curl`:

```
$ curl -sSL https://cli.openfaas.com | sudo sh
```

Non-root with curl (requires further actions as advised after downloading):

```
$ curl -sSL https://cli.openfaas.com | sh
```

Via brew:

```
$ brew install faas-cli
```

Note: The `brew` release may not run the latest minor release but is updated regularly.

Via npm (coming soon):

```
$ npm install --global @openfaas/faas-cli
```

Note: See `npm` specific installation instructions and usage in the [npm README.md](https://github.com/openfaas/faas-cli/blob/master/npm/README.md)

#### Windows

To install the faas-cli on Windows go to [Releases](https://github.com/openfaas/faas-cli/releases) and download the latest faas-cli.exe.

Or in PowerShell:

```
$version = (Invoke-WebRequest "https://api.github.com/repos/openfaas/faas-cli/releases/latest" | ConvertFrom-Json)[0].tag_name
(New-Object System.Net.WebClient).DownloadFile("https://github.com/openfaas/faas-cli/releases/download/$version/faas-cli.exe", "faas-cli.exe")
```

#### Build from source

The [contributing guide](CONTRIBUTING.md) has instructions for building from source and for configuring a Golang development environment.

### Run the CLI

The main commands supported by the CLI are:

* `faas-cli new` - creates a new function via a template in the current directory
* `faas-cli login` - stores basic auth credentials for OpenFaaS gateway (supports multiple gateways)
* `faas-cli logout` - removes basic auth credentials for a given gateway

* `faas-cli up` - a combination of `build/push and deploy`

* `faas-cli build` - builds Docker images from the supported language types
* `faas-cli push` - pushes Docker images into a registry
* `faas-cli deploy` - deploys the functions into a local or remote OpenFaaS gateway

* `faas-cli publish` - build and push multi-arch images for CI and release artifacts

* `faas-cli remove` - removes the functions from a local or remote OpenFaaS gateway
* `faas-cli invoke` - invokes the functions and reads from STDIN for the body of the request
* `faas-cli store` - allows browsing and deploying OpenFaaS store functions

* `faas-cli secret` - manage secrets for your functions

* `faas-cli pro auth` - initiates an OAuth2 authorization flow to obtain a token

* `faas-cli registry-login` - generate registry auth file in correct format by providing username and password for docker/ecr/self hosted registry

The default gateway URL of `127.0.0.1:8080` can be overridden in three places including an environmental variable.

* 1st priority `--gateway` flag
* 2nd priority `--yaml` / `-f` flag or `stack.yml` if in current directory
* 3rd priority `OPENFAAS_URL` environmental variable

For Kubernetes users you may want to set this in your `.bash_rc` file:

```
export OPENFAAS_URL=http://127.0.0.1:31112
```

Advanced commands:

* `faas-cli template pull` - pull in templates from a remote git repository [Detailed Documentation](guide/TEMPLATE.md)

The default template URL of `https://github.com/openfaas/templates.git` can be overridden in two places including an environmental variable

* 1st priority CLI input
* 2nd priority `OPENFAAS_TEMPLATE_URL` environmental variable

Help for all of the commands supported by the CLI can be found by running:

* `faas-cli help` or `faas-cli [command] --help`

You can chose between using a [programming language template](https://github.com/openfaas/templates/tree/master/template) where you only need to provide a handler file, or a Docker that you can build yourself.

#### `faas-cli pro auth`

The `auth` command is only licensed for OpenFaaS Pro customers.

Use the `auth` command to obtain a JWT to use as a Bearer token.

##### `code` grant - default

Use this flow to obtain a token for interactive use from your workstation.

The code grant flow uses the PKCE extension.

At this time the `token` cannot be saved or retained in your OpenFaaS config file. You can pass the token using a CLI flag of `--token=$TOKEN`.

Example:

```sh
faas-cli pro auth \
  --auth-url https://tenant0.eu.auth0.com/authorize \
  --token-url https://tenant0.eu.auth0.com/oauth/token \
  --audience http://gw.example.com \
  --client-id "${OAUTH_CLIENT_ID}"
```

##### `client_credentials` grant

Use this flow for machine to machine communication such as when you want to deploy a function to a gateway that uses OAuth2 / OIDC.

Example:

```sh
faas-cli pro auth \
  --grant client_credentials \
  --auth-url https://tenant0.eu.auth0.com/oauth/token \
  --client-id "${OAUTH_CLIENT_ID}" \
  --client-secret "${OAUTH_CLIENT_SECRET}"\
  --audience http://gw.example.com
```

##### Environment variable substitution

The CLI supports the use of `envsubst`-style templates. This means that you can have a single file with multiple configuration options such as for different user accounts, versions or environments.

Here is an example use-case, in your project there is an official and a development Docker Hub username/account. For the CI server images are always pushed to `exampleco`, but in development you may want to push to your own account such as `alexellis2`.

```yaml
functions:
  url-ping:
    lang: python
    handler: ./sample/url-ping
    image: ${DOCKER_USER:-exampleco}/faas-url-ping:0.2
```

Use the default:

```sh
$ faas-cli build
$ DOCKER_USER="" faas-cli build
```

Override with "alexellis2":

```
$ DOCKER_USER="alexellis2" faas-cli build
```

See also: [envsubst package from Drone](https://github.com/drone/envsubst).

#### Build templates

Command: `faas-cli new FUNCTION_NAME --lang python/node/go/ruby/Dockerfile/etc`

In your YAML you can also specify `lang: node/python/go/csharp/ruby`

* Supports common languages
* Quick and easy - just write one file
* Specify dependencies on Gemfile / requirements.txt or package.json etc

* Customise the provided templates

Perhaps you need to have [`gcc` or another dependency](https://github.com/faas-and-furious/faas-office-sample) in your Python template? That's not a problem.

You can customise the Dockerfile or code for any of the templates. Just create a new directory and copy in the templates folder from this repository. The templates in your current working directory are always used for builds.

See also: `faas-cli new --help`

**Third-party community templates**

Templates created and maintained by a third-party can be added to your local system using the `faas-cli template pull` command.

Read more on [community templates here](guide/TEMPLATE.md).

**Templates store**

The template store is a great way to find official, incubator and third-party templates.

Find templates with: `faas-cli template store list`

> Note: You can set your own custom store location with `--url` flag or set `OPENFAAS_TEMPLATE_STORE_URL` environmental variable

To pull templates from the store just write the name of the template you want `faas-cli template store pull go` or the repository and name `faas-cli template store pull openfaas/go`

To get more detail on a template just use the `template store describe` command and pick a template of your choice, example with `go` would look like this `faas-cli template store describe go`

> Note: This feature is still in experimental stage and in the future the CLI verbs might be changed

#### HMAC

It is possible to sign a `faas-cli invoke` request using a sha1 HMAC.  To do this, the name of a header to hold the code during transmission should be specified using the `--sign` flag, and the shared secret used to hash the message should be provided through `--key`. E.g.

```sh
$ echo -n OpenFaaS | faas-cli invoke env --sign X-Hub-Signature --key yoursecret
```

Results in the following header being added:

```
Http_X_Hub_Signature=sha1=2fc4758f8755f57f6e1a59799b56f8a6cf33b13f
```

#### Docker image as a function

Specify `lang: Dockerfile` if you want the faas-cli to execute a build or `skip_build: true` for pre-built images.

* Ultimate versatility and control
* Package anything
* If you are using a stack file add the `skip_build: true` attribute
* Use one of the [samples as a basis](https://github.com/openfaas/faas/tree/master/sample-functions)

Read the blog post/tutorial: [Turn Any CLI into a Function with OpenFaaS](https://blog.alexellis.io/cli-functions-with-openfaas/)

#### `faas-cli registry-login`

This command allows to generate the registry auth file in the correct format in the location `./credentials/config.json`

#### Prepare your Docker registry (if not using AWS ECR)

If you are using Dockerhub you only need to supply your --username and --password-stdin (or --password, but this leaves the password in history).

```bash
faas-cli registry-login --username <your-registry-username> --password-stdin
(then enter your password and use ctrl+d to finish input)
```

You could also have you password in a file, or environment variable and echo/cat this instead of entering interactively
If you are using a different registry (that is not ECR) then also provide a `--server` as well.

#### Prepare your Docker registry (if using AWS ECR)
```
faas-cli registry-login --ecr --region <your-aws-region> --account-id <your-account-id>
```

### Private registries

* For Kubernetes - [see here](https://docs.openfaas.com/deployment/kubernetes/#use-a-private-registry-with-kubernetes)

* For faasd - [see here](https://github.com/openfaas/faasd)

### Use faas-cli in CI environments

If you're running faas-cli in a CI environment like [Github Actions](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables), [CircleCI](https://circleci.com/docs/2.0/env-vars/#built-in-environment-variables), or [Travis](https://docs.travis-ci.com/user/environment-variables/#default-environment-variables), chances are you get the env var `CI` set to true.

If the `CI` env var is set to `true` or `1`, faas-cli change the location of the OpenFaaS config from the default `~/.openfaas/config.yml` to `.openfaas/config.yml` with elevated permissions for the `config.yml` and the shrinkwrapped `build` dir (if there is one).

This is really useful when running faas-cli as a container image. The recommended image type to use in a CI environment is the root variant, tagged with `-root` suffix.
CI environments like Github Actions require you to use Docker images having a root user. Learn more about it [here](https://docs.github.com/en/free-pro-team@latest/actions/creating-actions/dockerfile-support-for-github-actions#user).

### Use a YAML stack file

Read the [YAML reference guide in the OpenFaaS docs](https://docs.openfaas.com/reference/yaml/).

#### Quick guide

A YAML stack file groups functions together and also saves on typing.

You can define individual functions or a set of them within a YAML file. This makes the CLI easier to use and means you can use this file to deploy to your OpenFaaS instance.  By default the faas-cli will attempt to load `stack.yaml` from the current directory.

Here is an example file using the `stack.yml` file included in the repository.

```yaml
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080

functions:
  url-ping:
    lang: python
    handler: ./sample/url-ping
    image: alexellis2/faas-urlping
```

This url-ping function is defined in the sample/url-ping folder makes use of Python. All we had to do was to write a `handler.py` file and then to list off any Python modules in `requirements.txt`.

* Build the files in the .yml file:

```sh
$ faas-cli build
```

> `-f` specifies the file or URL to download your YAML file from. The long version of the `-f` flag is: `--yaml`.

You can also download over HTTP(s):

```sh
$ faas-cli build -f https://raw.githubusercontent.com/openfaas/faas-cli/master/stack.yml
```

Docker along with a Python template will be used to build an image named alexellis2/faas-urlping.

* Deploy your function

Now you can use the following command to deploy your function(s):

```sh
$ faas-cli deploy
```

### Access functions with `curl`

You can initiate a HTTP POST via `curl`:

* with the `-d` flag i.e. `-d "my data here"`
* or with `--data-binary @filename.txt` to send a whole file including newlines
* if you want to pass input from STDIN then use `--data-binary @-`

```sh
$ curl -d '{"hello": "world"}' http://127.0.0.1:8080/function/nodejs-echo
{ nodeVersion: 'v6.9.1', input: '{"hello": "world"}' }

$ curl --data-binary @README.md http://127.0.0.1:8080/function/nodejs-echo

$ uname -a | curl http://127.0.0.1:8080/function/nodejs-echo--data-binary @-
```

> For further instructions on the manual CLI flags (without using a YAML file) read [manual_cli.md](https://github.com/openfaas/faas-cli/blob/master/MANUAL_CLI.md)

### Environment variable overrides

* `OPENFAAS_TEMPLATE_URL` - to set the default URL to pull templates from
* `OPENFAAS_PREFIX` - for use with `faas-cli new` - this can act in place of `--prefix`
* `OPENFAAS_URL` - to override the default gateway URL
* `OPENFAAS_CONFIG` - to override the location of the configuration folder, which contains auth configuration.
* `CI` - to override the location of the configuration folder, when true, the configuration folder is `.openfaas` in the current working directory. This value is ignored if `OPENFAAS_CONFIG` is set.

### Contributing

See [contributing guide](https://github.com/openfaas/faas-cli/blob/master/CONTRIBUTING.md).

### License

Portions of this project are licensed under the OpenFaaS Pro EULA.

The remaining source unless annotated is licensed under the MIT License.
