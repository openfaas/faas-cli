## faas-cli
[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas-cli)](https://goreportcard.com/report/github.com/openfaas/faas-cli) [![Build Status](https://travis-ci.org/openfaas/faas-cli.svg?branch=master)](https://travis-ci.org/openfaas/faas-cli)

This is a CLI for use with [OpenFaaS](https://github.com/alexellis/faas) - a serverless functions framework for Docker & Kubernetes.

> Before using this tool please setup OpenFaaS by following instructions over on the main repo.

The CLI can be used to build and deploy functions to [OpenFaaS](https://github.com/alexellis/faas). You can build OpenFaaS functions from a set of supported language templates (such as Node.js, Python, CSharp and Ruby). That means you just write a handler file such as (handler.py/handler.js) and the CLI does the rest to create a Docker image.

Demo: [ASCII cinema](https://asciinema.org/a/121234)

#### TL;DR

[Blog: Coffee with the FaaS-CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)

### Intall the CLI

The easiest way to install the faas-cli is through a curl script or `brew`:

```
$ curl -sSL https://cli.openfaas.com | sudo sh
```

or

```
$ brew install faas-cli
```

> The contributing guide has instructions for building from source

### Run the CLI

The main commands supported by the CLI are:

* `faas-cli build` - builds Docker images from the supported language types
* `faas-cli push` - pushes Docker images into a registry
* `faas-cli deploy` - deploys the functions into a local or remote OpenFaaS gateway
* `faas-cli remove` - removes the functions from a local or remote OpenFaaS gateway

Help for all of the commands supported by the CLI can be found by running:

* `faas-cli help` or `faas-cli [command] --help`

You can chose between using a [programming language template](https://github.com/alexellis/faas-cli/tree/master/template) where you only need to provide a handler file, or a Docker that you can build yourself.

**Templates**

Specify `lang: node/python/csharp/ruby`

* Supports common languages
* Quick and easy - just write one file
* Specify depenencies on Gemfile / requirements.txt or package.json etc

* Customise the provided templates

Perhaps you need to have [`gcc` or another dependency](https://github.com/alexellis/faas-office-sample) in your Python template? That's not a problem.

You can customise the Dockerfile or code for any of the templates. Just create a new directory and copy in the templates folder from this repository. The templates in your current working directory are always used for builds.

**Docker image**

Specify `lang: Dockerfile` if you want the faas-cli to execute a build or `skip_build: true` for pre-supplied images.

* Ultimate versatility and control
* Package anything
* If you are using a stack file add the `skip_build: true` attribute
* Use one of the [samples as a basis](https://github.com/alexellis/faas/tree/master/sample-functions)

### Use a YAML stack file

A YAML stack file groups functions together and also saves on typing.

You can define individual functions or a set of of them within a YAML file. This makes the CLI easier to use and means you can use this file to deploy to your OpenFaaS instance.

Here is an example file using the `samples.yml` file included in the repository.

```yaml
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  url-ping:
    lang: python
    handler: ./sample/url-ping
    image: alexellis2/faas-urlping
```

This url-ping function is defined in the sample/url-ping folder makes use of Python. All we had to do was to write a `handler.py` file and then to list off any Python modules in `requirements.txt`.

* Build the files in the .yml file:

```
$ faas-cli build -f ./samples.yml
```

> `-f` specifies the file or URL to download your YAML file from. The long version of the `-f` flag is: `--yaml`.

You can also download over HTTP/s:

```
$ faas-cli build -f https://raw.githubusercontent.com/alexellis/faas-cli/master/samples.yml
```

Docker along with a Python template will be used to build an image named alexellis2/faas-urlping.

* Deploy your function

Now you can use the following command to deploy your function(s):

```
$ faas-cli deploy -f ./samples.yml
```

### Managing secrets

You can deploy secrets and configuration via environmental variables in-line or via external files.

> Note: external files take priority over in-line environmental variables. This allows you to specify a default and then have overrides within an external file.

Priority:

* environment_file - defined in zero to many external files

```yaml
  environment_file:
    - file1.yml
    - file2.yml
```

If you specify a variable such as "access_key" in more than one `environment_file` file then the last file in the list will take priority.

Environment file format:

```
environment:
  - access_key: key1
  - secret_key: key2
```

* Define environment in-line within the file:

Imagine you needed to define a `http_proxy` variable to operate within a corporate network:

```yaml
functions:
  url-ping:
    lang: python
    handler: ./sample/url-ping
    image: alexellis2/faas-urlping
    environment:
      http_proxy: http://proxy1.corp.com:3128
      no_proxy: http://gateway/
```

### Constraints

Constraints work with Docker Swarm and are useful for pinning functions to certain hosts.

Here is an example of picking only Linux:

```
   constraints:
     - "node.platform.os == linux"
```

Or only Windows:

```
   constraints:
     - "node.platform.os == windows"
```

### YAML reference

The possible entries for functions are documented below:

```yaml
functions:
  deployed_function_name:
    lang: node or python (optional)
    handler: ./path/to/handler (optional)
    image: docker-image-name
    environment:
      env1: value1
      env2: "value2"
   constraints:
     - "com.hdd == ssd"
```

Use environmental variables for setting tokens and configuration.

**Accessing the function with `curl`**

You can initiate a HTTP POST via `curl`:

* with the `-d` flag i.e. `-d "my data here"`
* or with `--data-binary @filename.txt` to send a whole file including newlines
* if you want to pass input from STDIN then use `--data-binary @-`

```
$ curl -d '{"hello": "world"}' http://localhost:8080/function/nodejs-echo
{ nodeVersion: 'v6.9.1', input: '{"hello": "world"}' }

$ curl --data-binary @README.md http://localhost:8080/function/nodejs-echo

$ uname -a | curl http://localhost:8080/function/nodejs-echo--data-binary @-
```

> For further instructions on the manual CLI flags (without using a YAML file) read [manual_cli.md](https://github.com/alexellis/faas-cli/blob/master/MANUAL_CLI.md)


**Bash Auto-completion [experimental]**

An experimental initial Bash auto-completion script for `faas-cli` is available at `contrib/bash/faas-cli`.

Please raise issues with feedback and suggestions on improvements to the auto-completion support.

This may be enabled it as follows.

*Enabling Bash auto-completion on OSX*

Brew install the `bash_completions` package.
```
$ brew install bash-completion
```
Add the following line to your `~/.bash_profile` if not already present.
```
[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion
```
Copy the provided `faas-cli` bash completion script from this repo.
```
cp contrib/bash/faas-cli /usr/local/etc/bash_completion.d/
```

*Enabling Bash auto-completion on Linux*

Refer to your distributions instructions on installing and enabling `bash-completion`, then copy the `faas-cli` completion script from `contrib/bash/` into the appropriate completion directory.

## FaaS-CLI Developers / Contributors

See [contributing guide](https://github.com/alexellis/faas-cli/blob/master/CONTRIBUTING.md).

### License

This project is part of the OpenFaaS project licensed under the MIT License.
