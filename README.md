## faas-cli

This CLI can be used to build and deploy functions to FaaS.

> Functions as a Service is a serverless framework for Docker: [Star on Github](https://github.com/alexellis/faas)

### Running the tool

The tool can be used to create a Docker image to be deployed on FaaS through a template meaning you only have to write a single handler file. The templates currently supported are:

* NodeJS (via handler.js)

So if you want to write in another language, just prepare a Dockerfile and build an image manually, like in the [FaaS samples](https://github.com/alexellis/faas/tree/master/sample-functions).

**Build a FaaS function in NodeJS from a template:**

This will generate a Docker image for a Node.js function using the code in `/samples/info`.

```
$ ./faas-cli -action=build \ 
   -image=alexellis2/hello-function \
   -name=hello-function \
   -handler=./samples/info

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

The CLI will thenn build a Docker image containing the FaaS watchdog and a bootstrap file to invoke your NodeJS function.

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

**Accessing the function with `curl`**

You can initiate a HTTP POST via `curl`:

* with the `-d` flag i.e. `-d "my data here"` 
* or with `--data-binary @filename.txt` to send a whole file including newlines

```
$ curl -d '{"hello": "world"}' http://localhost:8080/function/hello-function
```

**Installation (require Go 1.7 or later)**

```
$ cd $GOPATH
$ mkdir -p src/github.com/alexellis/
$ git clone https://github.com/alexellis/faas-cli
$ cd faas-cli
$ go get -d -v

$ go install
```
