## faas-cli

[![Build Status](https://travis-ci.org/alexellis/faas-cli.svg?branch=master)](https://travis-ci.org/alexellis/faas-cli)

This experimental CLI can be used to and deploy functions to FaaS or to build Node.js or Python functions from a templates meaning you just write a handler file (handler.py/handler.js). Read on for examples and usage.

> Functions as a Service is a serverless framework for Docker: [Read more on docs.get-faas.com](http://docs.get-faas.com/)

Website: www.openfaas.com

### Installing the tool

The easiest way to install the faas-cli is by doing:

```
$ curl -sSL https://cli.openfaas.com | sudo sh
```

Note that the tool is also available on brew. The last section also documents how to build it from source.

### Running the tool

The tool can be used to create a Docker image to be deployed on FaaS through a template meaning you only have to write a single handler file. The templates currently supported are: node and python, however you can create a FaaS function out of any process.

#### YAML files for ease of use

You can define individual functions or a set of of them within a YAML file. This makes the CLI easier to use and means you can use this file to deploy to your FaaS instance.

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
