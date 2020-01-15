## faas-cli
[![Build Status](https://travis-ci.com/openfaas/faas-cli.svg?branch=master)](https://travis-ci.com/openfaas/faas-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas-cli)](https://goreportcard.com/report/github.com/openfaas/faas-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

This is a CLI for use with [OpenFaaS](https://github.com/openfaas/faas) - a serverless functions framework for Docker & Kubernetes.

> Before using this tool please setup OpenFaaS by following instructions over on the main repo.

The CLI can be used to build and deploy functions to [OpenFaaS](https://github.com/openfaas/faas). You can build OpenFaaS functions from a set of supported language templates (such as Node.js, Python, CSharp and Ruby). That means you just write a handler file such as (handler.py/handler.js) and the CLI does the rest to create a Docker image.

Demo: [ASCII cinema](https://asciinema.org/a/141284)

### TL;DR - Introductory tutorial

[Blog: Coffee with the FaaS-CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)

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

* `faas-cli remove` - removes the functions from a local or remote OpenFaaS gateway
* `faas-cli invoke` - invokes the functions and reads from STDIN for the body of the request
* `faas-cli store` - allows browsing and deploying OpenFaaS store functions

* `faas-cli secret` - manage secrets for your functions

* `faas-cli auth` - (alpha) initiates an OAuth2 authorization flow to obtain a cookie

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

#### `faas-cli auth`

The `auth` command is currently available for alpha testing. Use the `auth` command to obtain a JWT to use as a Bearer token.

Two flow-types are supported in the CLI.

##### `code` grant - default

Use this flow to obtain a token.

At this time the `token` cannot be saved or retained in your OpenFaaS config file. You can pass the token using a CLI flag of `--token=$TOKEN`.

Example:

```sh
faas-cli auth \
  --auth-url https://tenant0.eu.auth0.com/authorize \
  --audience http://gw.example.com \
  --client-id "${OAUTH_CLIENT_ID}"
```

##### `client_credentials` grant

Use this flow for machine to machine communication such as when you want to deploy a function to a gateway that uses OAuth2 / OIDC.

Example:

```sh
faas-cli auth \
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

### Private registries

* For Kubernetes

Create a named image pull secret and add the secret name to the `secrets` section of your YAML file or your deployment arguments with `--secret`.

Alternatively you can assign a secret to the node to allow it to pull from your private registry. In this case you do not need to assign the secret to your function.

* For Docker Swarm

For Docker Swarm use the `--send-registry-auth` flag or its shorthand `-a` which will look up your registry credentials in your local credentials store and then transmit them over the wire to the deploy command on the API Gateway. Make sure HTTPS/TLS is enabled before attempting this.

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

### OpenFaaS Cloud (extensions)

[OpenFaaS Cloud](https://github.com/openfaas/openfaas-cloud) provides a GitOps experience for functions on Kubernetes.

Commands:

* `seal`

You can use the CLI to seal a secret for usage on public Git repo. The pre-requisite is that you have installed [SealedSecrets](https://github.com/bitnami-labs/sealed-secrets) and exported your public key from your cluster as `pub-cert.pem`.

Install `kubeseal` using `faas-cli` or the [SealedSecrets docs](https://github.com/bitnami-labs/sealed-secrets):

```sh
$ faas-cli cloud seal --download
```

You can also download a specific version:

```sh
$ faas-cli cloud seal --download --download-version v0.8.0
```

Now grab your pub-cert.pem file from your cluster, or use the official [OpenFaaS Cloud certificate](https://github.com/openfaas/cloud-functions/blob/master/pub-cert.pem).

```sh
$ kubeseal --fetch-cert --controller-name ofc-sealedsecrets-sealed-secrets > pub-cert.pem
```

Then seal a secret using the OpenFaaS CLI:

```
$ faas-cli cloud seal --name alexellis-github \
  --literal hmac-secret=1234 --cert=pub-cert.pem
```

You can then place the `secrets.yml` file in any public Git repo without others being able to read the contents.

When SealedSecrets is installed by ofc-bootstrap

The [scripts/export-sealed-secret-pubcert.sh](https://github.com/openfaas-incubator/ofc-bootstrap/blob/master/scripts/export-sealed-secret-pubcert.sh) does everything automatically.


### Environment variable overrides

* `OPENFAAS_TEMPLATE_URL` - to set the default URL to pull templates from
* `OPENFAAS_PREFIX` - for use with `faas-cli new` - this can act in place of `--prefix`
* `OPENFAAS_URL` - to override the default gateway URL

### FaaS-CLI Developers / Contributors

See [contributing guide](https://github.com/openfaas/faas-cli/blob/master/CONTRIBUTING.md).

### License

This project is part of OpenFaaS and is licensed under the MIT License.
