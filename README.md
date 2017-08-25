## faas-cli

[![Build Status](https://travis-ci.org/alexellis/faas-cli.svg?branch=master)](https://travis-ci.org/alexellis/faas-cli)

This is a CLI for use with [OpenFaaS](https://github.com/alexellis/faas) - a serverless functions framework for Docker & Kubernetes.

The CLI can be used to build and deploy functions to [OpenFaaS](https://github.com/alexellis/faas). You can build OpenFaaS functions from a set of supported language templates (such as Node.js, Python, CSharp and Ruby). That means you just write a handler file such as (handler.py/handler.js) and the CLI does the rest to create a Docker image.

Demo: [ASCII cinema](https://asciinema.org/a/121234)

### Installing the tool

The easiest way to install the faas-cli is through a curl script or `brew`:

```
$ curl -sSL https://cli.openfaas.com | sudo sh
```

or

```
$ brew install faas-cli
```

> The contributing guide has instructions for building from source

### Run the tool

The main actions for the tool are:

* `-action build` - builds Docker images from the supported language types
* `-action push` - pushes Docker images into a registry
* `-action deploy` - deploys the functions into a local or remote OpenFaaS gateway

You can chose between using a [programming language template](https://github.com/alexellis/faas-cli/tree/master/template) where you only need to provide a handler file, or a Docker that you can build yourself.

**Templates**

* Supports common languages
* Quick and easy - just write one file
* Specify depenencies on Gemfile / requirements.txt or package.json etc

* Customise the provided templates

Perhaps you need to have [`gcc` or another dependency](https://github.com/alexellis/faas-office-sample) in your Python template? That's not a problem.

You can customise the Dockerfile or code for any of the templates. Just create a new directory and copy in the templates folder from this repository. The templates in your current working directory are always used for builds.

**Docker image**

* Ultimate versatility and control
* Package anything
* If you are using a stack file add the `skip_build: true` attribute
* Use one of the [samples as a basis](https://github.com/alexellis/faas/tree/master/sample-functions)

#### YAML files for ease of use

You can define individual functions or a set of of them within a YAML file. This makes the CLI easier to use and means you can use this file to deploy to your OpenFaaS instance.

Here is an example file using the `samples.yml` file included in the repository.

```yaml
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  url-ping:
    lang: python
    handler: ./sample/url_ping
    image: alexellis2/faas-urlping
```

This url_ping function is defined in the samples/url__ping folder makes use of Python. All we had to do was to write a `handler.py` file and then to list off any Python modules in `requirements.txt`.

* Build the files in the .yml file:

```
$ faas-cli -action build -f ./samples.yml
```

> `-f` specifies the file or URL to download your YAML file from. The long version of the `-f` flag is: `-yaml`.

You can also download over HTTP/s:

```
$ faas-cli -action build -f https://raw.githubusercontent.com/alexellis/faas-cli/master/samples.yml
```

Docker along with a Python template will be used to build an image named alexellis2/faas-urlping.

* Deploy your function

Now you can use the following command to deploy your function(s):

```
$ faas-cli -action deploy -f ./samples.yml
```

* Possible entries for functions are documented below:

```yaml
functions:
  deployed_function_name:
    lang: node or python (optional)
    handler: ./path/to/handler (optional)
    image: docker-image-name
    environment:
      env1: value1
      env2: "value2"
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

## FaaS-CLI Developers / Contributors

See [contributing guide](https://github.com/alexellis/faas-cli/blob/master/CONTRIBUTING.md).

### License

This project is part of the OpenFaaS project licensed under the MIT License.
