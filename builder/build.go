package builder

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// BuildImage construct Docker image from function parameters
func BuildImage(image string, handler string, functionName string, languageTemplate string, nocache bool, squash bool) {
	var tempPath string
	var builder []string
	flagStr := buildFlagString(nocache, squash, os.Getenv("http_proxy"), os.Getenv("https_proxy"))

	switch languageTemplate {
	case "node", "python", "ruby", "csharp":
		tempPath = createBuildTemplate(functionName, handler, languageTemplate)
		builder = strings.Split(fmt.Sprintf("docker build %s-t %s .", flagStr, image), " ")
	case "custom":
		tempPath = createBuildCustom(functionName, handler, handler)
		builder = strings.Split(fmt.Sprintf("docker build %s-t %s .", flagStr, image), " ")
	default:
		log.Fatalf("Language template: %s not supported. Build a custom Dockerfile instead and set lang to 'custom'.", languageTemplate)
	}

	fmt.Printf("Building: %s with Docker. Please wait..\n", image)
	fmt.Println(strings.Join(builder, " "))
	ExecCommand(tempPath, builder)
	fmt.Printf("Image: %s built.\n", image)
}

// createBuildTemplate creates temporary build folder to perform a Docker build with Node template
func createBuildTemplate(functionName, handler, languageTemplate string) string {
	tempPath := initBuild(functionName)
	fmt.Printf("Prepared %s %s\n", handler+"/", tempPath+"function")

	// Drop in directory tree from languageTemplate
	copyFiles("./template/"+languageTemplate, tempPath, true)

	// Overlay in user-function
	copyFiles(handler, tempPath+"function/", false)

	return tempPath
}

// createBuildCustom creates temporary build folder to perform a Docker build with Node template
func createBuildCustom(functionName, handler, customTemplatePath string) string {
	tempPath := initBuild(functionName)
	fmt.Printf("Prepared %s %s\n", handler+"/", tempPath+"function")

	// Drop in directory tree from customTemplatePath
	copyFiles(customTemplatePath, tempPath, true)

	return tempPath
}

func initBuild(functionName string) string {
	tempPath := fmt.Sprintf("./build/%s/", functionName)
	fmt.Printf("Clearing temporary build folder: %s\n", tempPath)

	clearErr := os.RemoveAll(tempPath)
	if clearErr != nil {
		fmt.Printf("Error clearing temporary build folder %s\n", tempPath)
	}

	functionPath := tempPath + "/function"
	mkdirErr := os.MkdirAll(functionPath, 0700)
	if mkdirErr != nil {
		fmt.Printf("Error creating path %s - %s.\n", functionPath, mkdirErr.Error())
	}

	return tempPath
}

func copyFiles(src string, destination string, recursive bool) {

	files, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {

		if file.IsDir() == false {

			cp(src+"/"+file.Name(), destination+file.Name())

		} else {

			//make new destination dir
			newDir := destination + file.Name() + "/"
			if !pathExists(newDir) {

				newDirErr := os.Mkdir(newDir, 0700)

				if err != nil {
					fmt.Printf("Error creating path %s - %s.\n", newDir, newDirErr.Error())
				}
			}

			//did the call ask to recurse into sub directories?
			if recursive == true {
				//call copyTree to copy the contents
				copyFiles(src+"/"+file.Name(), newDir, true)
			}
		}
	}
}

func pathExists(path string) bool {
	exists := true

	if _, err := os.Stat(path); os.IsNotExist(err) {
		exists = false
	}

	return exists
}

func cp(src string, destination string) error {

	if val, exists := os.LookupEnv("debug"); exists && (val == "1" || val == "true") {
		fmt.Printf("cp - %s %s\n", src, destination)
	}

	memoryBuffer, readErr := ioutil.ReadFile(src)
	if readErr != nil {
		return fmt.Errorf("Error reading source file: %s\n" + readErr.Error())
	}
	writeErr := ioutil.WriteFile(destination, memoryBuffer, 0660)
	if writeErr != nil {
		return fmt.Errorf("Error writing file: %s\n" + writeErr.Error())
	}

	return nil
}

func buildFlagString(nocache bool, squash bool, httpProxy string, httpsProxy string) string {

	buildFlags := ""

	if nocache {
		buildFlags += "--no-cache "
	}
	if squash {
		buildFlags += "--squash "
	}

	if len(httpProxy) > 0 {
		buildFlags += fmt.Sprintf("--build-arg http_proxy=%s ", httpProxy)
	}

	if len(httpsProxy) > 0 {
		buildFlags += fmt.Sprintf("--build-arg https_proxy=%s ", httpsProxy)
	}

	return buildFlags
}
