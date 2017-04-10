## faas-cli

This CLI can be used to build and deploy functions to FaaS.

**Installation (require Go 1.7 or later)**

```
$ cd $GOPATH
$ mkdir -p src/github.com/alexellis/
$ git clone https://github.com/alexellis/faas-cli
$ cd faas-cli
$ go get -d -v

$ go install
```

**Running the tool:**

This will generate a Docker image for Node.js and deploy it as `hellofunction`

```
$ faas-cli -action=build -image=alexellis2/hello-function
Building: alexellis2/hello-cli with Docker. Please wait..
Image: alexellis2/hello-cli built.

$ faas-cli -action=deploy -image=alexellis2/hello-function -name=hellofunction
200 OK
URL: http://localhost:8080/function/hello1
```

This will use the handler.js file found in the template/node folder to build a Docker image containing the FaaS watchdog.


**Accessing the function:**

```
curl -d "input" http://localhost:8080/function/hello1
```


