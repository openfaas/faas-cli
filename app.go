package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"net/http"

	"encoding/json"

	"bytes"
	"os"

	"github.com/alexellis/faas/gateway/requests"
)

func main() {
	// var handler string
	var handler string
	var image string

	var action string
	var functionName string
	var gateway string
	var fprocess string
	var language string
    var replace bool

	flag.StringVar(&handler, "handler", "", "handler for function, i.e. handler.js")
	flag.StringVar(&image, "image", "", "Docker image name to build")
	flag.StringVar(&action, "action", "", "either build or deploy")
	flag.StringVar(&functionName, "name", "", "give the name of your deployed function")
	flag.StringVar(&gateway, "gateway", "http://localhost:8080", "gateway URI - i.e. http://localhost:8080")
	flag.StringVar(&fprocess, "fprocess", "", "fprocess to be run by the watchdog")
	flag.StringVar(&language, "lang", "node", "programming language template, default is: node")
    flag.BoolVar(&replace, "replace", true, "replace any existing function")

	flag.Parse()

	if len(action) == 0 {
		fmt.Println("give either -action= build or deploy")
		return
	}

	if action == "build" {
		if len(image) == 0 {
			fmt.Println("Give a valid -image name for your Docker image.")
			return
		}
		if len(handler) == 0 {
			fmt.Println("Please give the full path to your function's handler.")
			return
		}
		if len(functionName) == 0 {
			fmt.Println("Please give the deployed -name of your function")
			return
		}

		tempPath := createBuildTemplate(functionName, handler, language)

		fmt.Printf("Building: %s with Docker. Please wait..\n", image)

		builder := strings.Split(fmt.Sprintf("docker build -t %s .", image), " ")
		if len(os.Getenv("http_proxy")) > 0 || len(os.Getenv("http_proxy")) > 0 {
			builder = strings.Split(fmt.Sprintf("docker build --build-arg http_proxy=%s --build-arg https_proxy=%s -t %s .", os.Getenv("http_proxy"), os.Getenv("https_proxy"), image), " ")
		}

		fmt.Println(strings.Join(builder, " "))
		targetCmd := exec.Command(builder[0], builder[1:]...)
		targetCmd.Dir = tempPath
		cmdOutput, cmdErr := targetCmd.CombinedOutput()

		if cmdErr != nil {
			fmt.Printf("Error: %s\n" + cmdErr.Error())
		}

		fmt.Println(string(cmdOutput))

		fmt.Printf("Image: %s built.\n", image)
	} else if action == "deploy" {
		if len(image) == 0 {
			fmt.Println("Give an image name to be deployed.")
			return
		}
		if len(functionName) == 0 {
			fmt.Println("Give a -name for your function as it will be deployed on FaaS")
			return
		}

		// Need to alter Gateway to allow nil/empty string as fprocess, to avoid this repetition.
		fprocessTemplate := "node index.js"
		if len(fprocess) > 0 {
			fprocessTemplate = fprocess
		}
		if language == "python" {
			fprocessTemplate = "python index.py"
		}

        // TODO: Extract to function
        if replace {
          deleteFunction(gateway, functionName) 
        }

		req := requests.CreateFunctionRequest{
			EnvProcess: fprocessTemplate,
			Image:      image,
			Network:    "func_functions",
			Service:    functionName,
		}

		reqBytes, _ := json.Marshal(&req)
		reader := bytes.NewReader(reqBytes)
        res, err := http.Post(gateway+"/system/functions", "application/json", reader)
		if err != nil {
			fmt.Println("Is FaaS deployed? Do you need to specify the -gateway flag?")
			fmt.Println(err)
			return
		}
		fmt.Println(res.Status)
		deployedUrl := fmt.Sprintf("URL: %s/function/%s\n", gateway, functionName)
		fmt.Println(deployedUrl)

	}
}

func deleteFunction(gateway string, functionName string) {
     delReq := requests.DeleteFunctionRequest{ FunctionName: functionName }
    		reqBytes, _ := json.Marshal(&delReq)
    		reader := bytes.NewReader(reqBytes)
	
            c := http.Client{}
            req, _ := http.NewRequest("DELETE", gateway + "/system/functions", reader)
            req.Header.Set("Content-Type", "application/json")
            delRes, delErr:= c.Do(req)

            if(delErr!=nil) {
                fmt.Println(delErr.Error())
            }
            switch delRes.StatusCode {
                case 200:
                    fmt.Println("Removing old service.")
                case 404:
                    fmt.Println("No existing service to remove")

            }
}


// createBuildTemplate creates temporary build folder to perform a Docker build with Node template
func createBuildTemplate(functionName string, handler string, language string) string {
	tempPath := fmt.Sprintf("./build/%s/", functionName)
	fmt.Printf("Creating temporary folder: %s\n", tempPath)
	exec.Command("mkdir", "-p", tempPath).Run()

	if language == "node" {
		exec.Command("cp", "./template/node/index.js", tempPath).Run()
		exec.Command("cp", "./template/node/Dockerfile", tempPath).Run()
		exec.Command("cp", "./template/node/package.json", tempPath).Run()
	} else if language == "python" {
		exec.Command("cp", "./template/python/index.py", tempPath).Run()
		exec.Command("cp", "./template/python/Dockerfile", tempPath).Run()
		exec.Command("cp", "./template/python/requirements.txt", tempPath).Run()
	}

	exec.Command("mkdir", "-p", tempPath+"/function").Run()
	exec.Command("cp", "-rf", handler+"/", tempPath+"/function/").Run()
	return tempPath
}
