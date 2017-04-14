## faas-cli

This CLI can be used to build and deploy functions to FaaS.

> Functions as a Service is a serverless framework for Docker: [Star on Github](https://github.com/alexellis/faas)

### Running the tool

**Build a Docker image:**

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

Put your code in the handler.js file, for example:

```
"use strict"

module.exports = (context, callback) => {
    console.log("echo - " + context);
    
    callback(undefined, {status: "done"});
}
```

The CLI will build Docker image containing the FaaS watchdog to handle communication between the gateway and Node.js.

**Deploy the Docker image as a FaaS function:**

Now we can deploy the image as a named function called `hello-function`.

```
$ ./faas-cli -action=deploy \
   -image=alexellis2/hello-function \
   -name=hello-function

200 OK

URL: http://localhost:8080/function/hello-function
```

**Accessing the function:**

You can pass input with the `-d` flag or `--data-binary @filename.txt` to send a whole file including newlines.

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
