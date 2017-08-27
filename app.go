// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"fmt"
	"log"

	"os"

	"github.com/alexellis/faas-cli/proxy"
	"github.com/alexellis/faas-cli/stack"

	"github.com/alexellis/faas-cli/builder"
)

const defaultNetwork = "func_functions"

// GitCommit injected at build-time
var GitCommit string

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
	var nocache bool
	var yamlFile string
	var yamlFileShort string
	var version bool
	var squash bool

	flag.StringVar(&handler, "handler", "", "handler for function, i.e. handler.js")
	flag.StringVar(&image, "image", "", "Docker image name to build")
	flag.StringVar(&action, "action", "", "Available actions: build, deploy, push, delete")
	flag.StringVar(&functionName, "name", "", "give the name of your deployed function")
	flag.StringVar(&gateway, "gateway", "http://localhost:8080", "gateway URI - i.e. http://localhost:8080")
	flag.StringVar(&fprocess, "fprocess", "", "fprocess to be run by the watchdog")
	flag.StringVar(&language, "lang", "node", "programming language template, default is: node")
	flag.BoolVar(&replace, "replace", true, "replace any existing function")
	flag.BoolVar(&nocache, "no-cache", false, "do not use Docker's build cache")

	flag.StringVar(&yamlFile, "yaml", "", "use a yaml file for a set of functions")
	flag.StringVar(&yamlFileShort, "f", "", "use a yaml file for a set of functions (same as -yaml)")
	flag.BoolVar(&version, "version", false, "show version and quit")
	flag.BoolVar(&squash, "squash", false, "use Docker's squash flag for potentially smaller images (currently experimental)")

	flag.Parse()

	if version {
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	// support short-argument -f
	if len(yamlFile) == 0 && len(yamlFileShort) > 0 {
		yamlFile = yamlFileShort
	}

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAML(yamlFile)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if len(action) == 0 {
		fmt.Println("give either -action= build or deploy")
		return
	}

	switch action {
	case "build":

		if pullErr := pullTemplates(); pullErr != nil {
			log.Fatalln("Could not pull templates for FaaS.", pullErr)
		}

		if len(services.Functions) > 0 {
			for k, function := range services.Functions {
				if function.SkipBuild {
					fmt.Printf("Skipping build of: %s.\n", function.Name)
				} else {
					function.Name = k
					// fmt.Println(k, function)
					fmt.Printf("Building: %s.\n", function.Name)
					builder.BuildImage(function.Image, function.Handler, function.Name, function.Language, nocache, squash)
				}
			}
		} else {
			if len(image) == 0 {
				fmt.Println("Please provide a valid -image name for your Docker image.")
				return
			}
			if len(handler) == 0 {
				fmt.Println("Please provide the full path to your function's handler.")
				return
			}
			if len(functionName) == 0 {
				fmt.Println("Please provide the deployed -name of your function.")
				return
			}
			builder.BuildImage(image, handler, functionName, language, nocache, squash)
		}
		break
	case "delete":
		if len(services.Functions) > 0 {
			if len(services.Provider.Network) == 0 {
				services.Provider.Network = defaultNetwork
			}

			for k, function := range services.Functions {
				function.Name = k
				fmt.Printf("Deleting: %s.\n", function.Name)

				proxy.DeleteFunction(services.Provider.GatewayURL, function.Name)
			}
		} else {
			if len(functionName) == 0 {
				fmt.Println("Please provide a -name for your function as it will be deployed on FaaS")
				return
			}
			fmt.Printf("Deleting: %s.\n", functionName)
			proxy.DeleteFunction(gateway, functionName)
		}

		break
	case "deploy":
		if len(services.Functions) > 0 {
			if len(services.Provider.Network) == 0 {
				services.Provider.Network = defaultNetwork
			}

			for k, function := range services.Functions {
				function.Name = k
				// fmt.Println(k, function)
				fmt.Printf("Deploying: %s.\n", function.Name)

				proxy.DeployFunction(function.FProcess, services.Provider.GatewayURL, function.Name, function.Image, function.Language, replace, function.Environment, services.Provider.Network)
			}
		} else {
			if len(image) == 0 {
				fmt.Println("Please provide an image name to be deployed.")
				return
			}
			if len(functionName) == 0 {
				fmt.Println("Please provide a -name for your function as it will be deployed on FaaS")
				return
			}

			proxy.DeployFunction(fprocess, gateway, functionName, image, language, replace, map[string]string{}, defaultNetwork)
		}
		break
	case "push":
		if len(services.Functions) > 0 {
			for k, function := range services.Functions {
				function.Name = k
				fmt.Printf("Pushing: %s to remote repository.\n", function.Name)
				pushImage(function.Image)
			}
		} else {
			fmt.Println("The '-action push' flag only works with a YAML file.")
			return
		}
	default:
		fmt.Println("-action must be 'build', 'deploy', 'push' or 'delete'.")
		break
	}
}

func pushImage(image string) {
	builder.ExecCommand("./", []string{"docker", "push", image})
}

func pullTemplates() error {
	var err error
	exists, err := os.Stat("./template")
	if err != nil || exists == nil {
		log.Println("No templates found in current directory.")

		err = fetchTemplates()
		if err != nil {
			log.Println("Unable to download templates from Github.")
			return err
		}
	}
	return err
}
