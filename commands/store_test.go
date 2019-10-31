package commands

import (
	"encoding/json"
	"testing"

	storeV2 "github.com/openfaas/faas-cli/schema/store/v2"
)

func matchFilteredOutout(t *testing.T, expectedOuputFunctions, filteredFunctions []storeV2.StoreFunction, platform string) {
	if len(expectedOuputFunctions) != len(filteredFunctions) {
		t.Errorf("Length did not match, expected: %v, got: %v", len(expectedOuputFunctions), len(filteredFunctions))
	}
	var isFound bool

	for _, expectedFunction := range expectedOuputFunctions {
		isFound = false

		for _, filteredFunction := range filteredFunctions {
			_, hasPlatform := filteredFunction.Images[platform]
			if expectedFunction.Name == filteredFunction.Name && hasPlatform {
				isFound = true
			}
		}

		if !isFound {
			t.Errorf("Function %s not present in the output", expectedFunction.Name)
		}
	}
}

func getInputStoreFunctions(t *testing.T) []storeV2.StoreFunction {
	inputJSONBytes := []byte(`[{
        "title": "NodeInfo",
        "name": "nodeinfo",
        "description": "Get info about the machine that you're deployed on. Tells CPU count, hostname, OS, and Uptime",
        "images": {
            "arm64": "functions/nodeinfo:arm64",
            "armhf": "functions/nodeinfo:latest-armhf",
            "x86_64": "functions/nodeinfo:latest"
        },
        "fprocess": "node main.js",
        "network": "func_functions",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/NodeInfo"
    },
    {
        "title": "sha512sum",
        "name": "sha512sum",
        "description": "Generate shasums in 512 format",
        "images": {
            "arm64": "functions/alpine:latest-arm64"
        },
        "fprocess": "sha512sum",
        "network": "func_functions",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/AlpineFunction"
    },
    {
        "title": "Figlet",
        "name": "figlet",
        "description": "Generate ASCII logos with the figlet CLI",
        "images": {
			"armhf": "functions/figlet:latest-armhf",
            "x86_64": "functions/figlet:0.9.6"
        },
        "fprocess": "figlet",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/figlet"
	}]`)

	var inputFunctions []storeV2.StoreFunction
	err := json.Unmarshal(inputJSONBytes, &inputFunctions)
	if err != nil {
		t.Errorf(err.Error())
	}

	return inputFunctions
}

func Test_filterStoreList_x86_64(t *testing.T) {
	outputJSONBytes := []byte(`[{
        "title": "NodeInfo",
        "name": "nodeinfo",
        "description": "Get info about the machine that you're deployed on. Tells CPU count, hostname, OS, and Uptime",
        "images": {
            "arm64": "functions/nodeinfo:arm64",
            "armhf": "functions/nodeinfo:latest-armhf",
            "x86_64": "functions/nodeinfo:latest"
        },
        "fprocess": "node main.js",
        "network": "func_functions",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/NodeInfo"
    }, {
		"title": "Figlet",
        "name": "figlet",
        "description": "Generate ASCII logos with the figlet CLI",
        "images": {
            "armhf": "functions/figlet:latest-armhf",
            "x86_64": "functions/figlet:0.9.6"
        },
        "fprocess": "figlet",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/figlet"
	}]`)

	inputFunctions := getInputStoreFunctions(t)
	var expectedOuputFunctions []storeV2.StoreFunction
	err := json.Unmarshal(outputJSONBytes, &expectedOuputFunctions)
	if err != nil {
		t.Errorf(err.Error())
	}

	filteredFunctions := filterStoreList(inputFunctions, "x86_64")
	matchFilteredOutout(t, expectedOuputFunctions, filteredFunctions, "x86_64")

}

func Test_filterStoreList_armhf(t *testing.T) {
	outputJSONBytes := []byte(`[{
        "title": "NodeInfo",
        "name": "nodeinfo",
        "description": "Get info about the machine that you're deployed on. Tells CPU count, hostname, OS, and Uptime",
        "images": {
            "arm64": "functions/nodeinfo:arm64",
            "armhf": "functions/nodeinfo:latest-armhf",
            "x86_64": "functions/nodeinfo:latest"
        },
        "fprocess": "node main.js",
        "network": "func_functions",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/NodeInfo"
    }, {
		"title": "Figlet",
        "name": "figlet",
        "description": "Generate ASCII logos with the figlet CLI",
        "images": {
            "armhf": "functions/figlet:latest-armhf",
            "x86_64": "functions/figlet:0.9.6"
        },
        "fprocess": "figlet",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/figlet"
	}]`)

	inputFunctions := getInputStoreFunctions(t)
	var expectedOuputFunctions []storeV2.StoreFunction
	err := json.Unmarshal(outputJSONBytes, &expectedOuputFunctions)
	if err != nil {
		t.Errorf(err.Error())
	}

	filteredFunctions := filterStoreList(inputFunctions, "armhf")
	matchFilteredOutout(t, expectedOuputFunctions, filteredFunctions, "armhf")

}

func Test_filterStoreList_arm64(t *testing.T) {
	outputJSONBytes := []byte(`[{
        "title": "NodeInfo",
        "name": "nodeinfo",
        "description": "Get info about the machine that you're deployed on. Tells CPU count, hostname, OS, and Uptime",
        "images": {
            "arm64": "functions/nodeinfo:arm64",
            "armhf": "functions/nodeinfo:latest-armhf",
            "x86_64": "functions/nodeinfo:latest"
        },
        "fprocess": "node main.js",
        "network": "func_functions",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/NodeInfo"
    },{
        "title": "sha512sum",
        "name": "sha512sum",
        "description": "Generate shasums in 512 format",
        "images": {
            "arm64": "functions/alpine:latest-arm64"
        },
        "fprocess": "sha512sum",
        "network": "func_functions",
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/AlpineFunction"
    }]`)

	inputFunctions := getInputStoreFunctions(t)
	var expectedOuputFunctions []storeV2.StoreFunction
	err := json.Unmarshal(outputJSONBytes, &expectedOuputFunctions)
	if err != nil {
		t.Errorf(err.Error())
	}

	filteredFunctions := filterStoreList(inputFunctions, "arm64")
	matchFilteredOutout(t, expectedOuputFunctions, filteredFunctions, "arm64")

}

func Test_filterStoreList_other(t *testing.T) {
	outputJSONBytes := []byte(`[]`)

	inputFunctions := getInputStoreFunctions(t)
	var expectedOuputFunctions []storeV2.StoreFunction
	err := json.Unmarshal(outputJSONBytes, &expectedOuputFunctions)
	if err != nil {
		t.Errorf(err.Error())
	}

	filteredFunctions := filterStoreList(inputFunctions, "other")
	matchFilteredOutout(t, expectedOuputFunctions, filteredFunctions, "other")

}

func Test_getStorePlatforms(t *testing.T) {
	var expectedPlatforms = []string{"arm64", "armhf", "x86_64"}
	inputFunctions := getInputStoreFunctions(t)
	actualPlatforms := getStorePlatforms(inputFunctions)

	if len(expectedPlatforms) != len(actualPlatforms) {
		t.Errorf("Length of platforms did not match, expected: %d, got: %d", len(expectedPlatforms), len(actualPlatforms))
	}

	for _, expectedPlatform := range expectedPlatforms {
		isFound := false
		for _, actualPlatform := range actualPlatforms {
			if expectedPlatform == actualPlatform {
				isFound = true
			}
		}

		if !isFound {
			t.Errorf("Expected value '%s' not found in actual platforms, expected: %v, actual: %v", expectedPlatform, expectedPlatforms, actualPlatforms)
		}
	}
}

func Test_storeFindFunction_Positive(t *testing.T) {
	inputFunctions := getInputStoreFunctions(t)
	expectedFunctionName := "nodeinfo"

	actualFunction := storeFindFunction(expectedFunctionName, inputFunctions)

	if actualFunction.Name != expectedFunctionName {
		t.Errorf("Function %s not found in store", expectedFunctionName)
	}
}

func Test_storeFindFunction_Negative(t *testing.T) {
	inputFunctions := getInputStoreFunctions(t)
	expectedFunctionName := "openfaas-ocr"

	actualFunction := storeFindFunction(expectedFunctionName, inputFunctions)

	if actualFunction != nil {
		t.Errorf("Function %s is found in store", expectedFunctionName)
	}
}
