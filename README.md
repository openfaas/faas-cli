## faas-cli

[![Build Status](https://travis-ci.org/alexellis/faas-cli.svg?branch=master)](https://travis-ci.org/alexellis/faas-cli)

This experimental CLI can be used to and deploy functions to FaaS or to build Node.js or Python functions from a templates meaning you just write a handler file (handler.py/handler.js). Read on for examples and usage.

> Functions as a Service is a serverless framework for Docker: [Read more on docs.get-faas.com](http://docs.get-faas.com/)

### Running the tool

The tool can be used to create a Docker image to be deployed on FaaS through a template meaning you only have to write a single handler file. The templates currently supported are:

There are two tester scripts included in the repository test_python.sh and test_node.sh - check them out or read on for a detailled explanation.

#### Worked example with Node.js

So if you want to write in another language, just prepare a Dockerfile and build an image manually, like in the [FaaS samples](https://github.com/alexellis/faas/tree/master/sample-functions).

**Build a FaaS function in NodeJS from a template:**

This will generate a Docker image for a Node.js function using the code in `/samples/info`.

* The `faas-cli` can accept a `-lang` option of `python` or `node` and is `node` by default.

```
   $ ./faas-cli -action=build \
      -image=alexellis2/hello-function \
      -name=hello-function \
      -handler=./sample/info

Building: alexellis2/hello-cli with Docker. Please wait..
...
Image: alexellis2/hello-cli built.
```

You can customise the code by editing the handler.js file and changing the `-handler` parameter. You can also edit the packages.json file, which will be used during the build to make sure all your dependencies are available at runtime.

For example:

```
"use strict"

module.exports = (context, callback) => {
    console.log("echo - " + context);
    
    callback(undefined, {status: "done"});
}
```

The CLI will then build a Docker image containing the FaaS watchdog and a bootstrap file to invoke your NodeJS function.

**Deploy the Docker image as a FaaS function:**

Now we can deploy the image as a named function called `hello-function`.

```
$ ./faas-cli -action=deploy \
   -image=alexellis2/hello-function \
   -name=hello-function

200 OK

URL: http://localhost:8080/function/hello-function
```

> This tool can be used to deploy any Docker image as a FaaS function, as long as it includes the watchdog binary as the `CMD` or `ENTRYPOINT` of the image.

*Deploy remotely*

You can deploy to a remote FaaS instance as along as you push the image to the Docker Hub, or another accessible Docker registry. Specify your remote gateway with the following flag: `-gateway=http://remote-site.com:8080`

**Accessing the function with `curl`**

You can initiate a HTTP POST via `curl`:

* with the `-d` flag i.e. `-d "my data here"` 
* or with `--data-binary @filename.txt` to send a whole file including newlines
* if you want to pass input from STDIN then use `--data-binary @-`

```
$ curl -d '{"hello": "world"}' http://localhost:8080/function/hello-function
{ nodeVersion: 'v6.9.1', input: '{"hello": "world"}' }

$ curl --data-binary @README.md http://localhost:8080/function/hello-function

$ uname -a | curl http://localhost:8080/function/hello-function --data-binary @-
```

### License and contributing

This project is part of the FaaS project licensed under the MIT License.

For more details see the [Contributing guide](https://github.com/alexellis/faas-cli/blob/master/CONTRIBUTING.md).


### Installation / pre-requirements

* Docker

Install Docker because it is used to build Docker images if you create new functions.

* FaaS - deployed and live

This CLI can build and deploy templated functions, so it's best if you have FaaS started up on your laptop. Head over to http://docs.get-faas.com/ and get up and running with a sample stack in 60 seconds.

* Golang

> Here's how to install Go in 60 seconds.

* Grab Go 1.7.x from https://golang.org/dl/

Then after installing run this command or place it in your `$HOME/.bash_profile`

```
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
