package builder

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// BuildImage construct Docker image from function parameters
func BuildImage(image string, handler string, functionName string, language string, nocache bool, squash bool) {

	switch language {
	case "node", "python", "ruby", "csharp":
		tempPath := createBuildTemplate(functionName, handler, language)

		fmt.Printf("Building: %s with %s template. Please wait..\n", image, language)

		flagStr := buildFlagString(nocache, squash, os.Getenv("http_proxy"), os.Getenv("https_proxy"))

		builder := strings.Split(fmt.Sprintf("docker build %s-t %s .", flagStr, image), " ")
		fmt.Println(strings.Join(builder, " "))
		ExecCommand(tempPath, builder)
		fmt.Printf("Image: %s built.\n", image)

		break
	case "Dockerfile", "dockerfile":
		tempPath := handler
		if _, err := os.Stat(handler); err != nil {
			fmt.Printf("Unable to build %s, %s is an invalid path\n", image, handler)
			fmt.Printf("Image: %s not built.\n", image)

			break
		}
		fmt.Printf("Building: %s with Dockerfile. Please wait..\n", image)

		flagStr := buildFlagString(nocache, squash, os.Getenv("http_proxy"), os.Getenv("https_proxy"))

		builder := strings.Split(fmt.Sprintf("docker build %s-t %s .", flagStr, image), " ")
		fmt.Println(strings.Join(builder, " "))
		ExecCommand(tempPath, builder)
		fmt.Printf("Image: %s built.\n", image)

	default:
		log.Fatalf("Language template: %s not supported. Build a custom Dockerfile instead.", language)
	}
}

// createBuildTemplate creates temporary build folder to perform a Docker build with Node template
func createBuildTemplate(functionName string, handler string, language string) string {
	tempPath := fmt.Sprintf("./build/%s/", functionName)
	fmt.Printf("Clearing temporary build folder: %s\n", tempPath)

	clearErr := os.RemoveAll(tempPath)
	if clearErr != nil {
		fmt.Printf("Error clearing temporary build folder %s\n", tempPath)
	}

	fmt.Printf("Preparing %s %s\n", handler+"/", tempPath+"function")

	functionPath := tempPath + "/function"
	mkdirErr := os.MkdirAll(functionPath, 0700)
	if mkdirErr != nil {
		fmt.Printf("Error creating path %s - %s.\n", functionPath, mkdirErr.Error())
	}

	// Drop in directory tree from template
	copyFiles("./template/"+language, tempPath, true)

	// Overlay in user-function
	copyFiles(handler, tempPath+"function/", false)

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
