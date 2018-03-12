### Manual CLI options

In addition to YAML file support, you can use the CLI to build and deploy individual functions as follows:

#### Worked example with Node.js

So if you want to write in another language, just prepare a Dockerfile and build an image manually, like in the [FaaS samples](https://github.com/openfaas/faas/tree/master/sample-functions).

**Build a FaaS function in NodeJS from a template:**

This will generate a Docker image for a Node.js function using the code in `/samples/info`.

* The `faas-cli build` command can accept a `--lang` option of `python`, `node`, `ruby`, `csharp`, `python3`, `go`, or `dockerfile`.

```
   $ faas-cli build \
      --image=alexellis2/node_info \
      --lang=node \
      --name=node_info \
      --handler=./sample/node_info

Building: alexellis2/node_info with Docker. Please wait..
...
Image: alexellis2/node_info built.
```

You can customise the code by editing the handler.js file and changing the `--handler` parameter. You can also edit the packages.json file, which will be used during the build to make sure all your dependencies are available at runtime.

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

Now we can deploy the image as a named function called `node_info`.

```
$ faas-cli deploy \
   --image=alexellis2/node_info \
   --name=node_info

200 OK

URL: http://127.0.0.1:8080/function/node_info
```

> This tool can be used to deploy any Docker image as a FaaS function, as long as it includes the watchdog binary as the `CMD` or `ENTRYPOINT` of the image.

*Deploy remotely*

You can deploy to a remote FaaS instance as along as you push the image to the Docker Hub, or another accessible Docker registry. Specify your remote gateway with the following flag: `--gateway=http://remote-site.com:8080`
