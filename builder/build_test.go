package builder

import (
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/stack"
)

func Test_buildFlagSlice(t *testing.T) {

	var buildFlagOpts = []struct {
		title         string
		nocache       bool
		squash        bool
		httpProxy     string
		httpsProxy    string
		buildArgMap   map[string]string
		buildPackages []string
		expectedSlice []string
	}{
		{
			title:         "no cache only",
			nocache:       true,
			squash:        false,
			httpProxy:     "",
			httpsProxy:    "",
			buildArgMap:   make(map[string]string),
			buildPackages: []string{},
			expectedSlice: []string{"--no-cache"},
		},
		{
			title:         "no cache & squash only",
			nocache:       true,
			squash:        true,
			httpProxy:     "",
			httpsProxy:    "",
			buildArgMap:   make(map[string]string),
			buildPackages: []string{},
			expectedSlice: []string{"--no-cache", "--squash"},
		},
		{
			title:         "no cache & squash & http proxy only",
			nocache:       true,
			squash:        true,
			httpProxy:     "192.168.0.1",
			httpsProxy:    "",
			buildArgMap:   make(map[string]string),
			buildPackages: []string{},
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "http_proxy=192.168.0.1"},
		},
		{
			title:         "no cache & squash & https-proxy only",
			nocache:       true,
			squash:        true,
			httpProxy:     "",
			httpsProxy:    "127.0.0.1",
			buildArgMap:   make(map[string]string),
			buildPackages: []string{},
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "https_proxy=127.0.0.1"},
		},
		{
			title:         "no cache & squash & http-proxy & https-proxy only",
			nocache:       true,
			squash:        true,
			httpProxy:     "192.168.0.1",
			httpsProxy:    "127.0.0.1",
			buildArgMap:   make(map[string]string),
			buildPackages: []string{},
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "http_proxy=192.168.0.1", "--build-arg", "https_proxy=127.0.0.1"},
		},
		{
			title:         "http-proxy & https-proxy only",
			nocache:       false,
			squash:        false,
			httpProxy:     "192.168.0.1",
			httpsProxy:    "127.0.0.1",
			buildArgMap:   make(map[string]string),
			buildPackages: []string{},
			expectedSlice: []string{"--build-arg", "http_proxy=192.168.0.1", "--build-arg", "https_proxy=127.0.0.1"},
		},
		{
			title:      "build arg map no spaces",
			nocache:    false,
			squash:     false,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppet": "ernie",
			},
			buildPackages: []string{},
			expectedSlice: []string{"--build-arg", "muppet=ernie"},
		},
		{
			title:      "build arg map with spaces",
			nocache:    false,
			squash:     false,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppets": "burt and ernie",
			},
			buildPackages: []string{},
			expectedSlice: []string{"--build-arg", "muppets=burt and ernie"},
		},
		{
			title:      "multiple build arg map with spaces",
			nocache:    false,
			squash:     false,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppets":    "burt and ernie",
				"playschool": "Jemima",
			},
			buildPackages: []string{},
			expectedSlice: []string{"--build-arg", "muppets=burt and ernie", "--build-arg", "playschool=Jemima"},
		},
		{
			title:      "no-cache and squash with multiple build arg map with spaces",
			nocache:    true,
			squash:     true,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppets":    "burt and ernie",
				"playschool": "Jemima",
			},
			buildPackages: []string{},
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "muppets=burt and ernie", "--build-arg", "playschool=Jemima"},
		},
	}

	for _, test := range buildFlagOpts {

		t.Run(test.title, func(t *testing.T) {

			flagSlice := buildFlagSlice(test.nocache, test.squash, test.httpProxy, test.httpsProxy, test.buildArgMap, test.buildPackages)

			if len(flagSlice) != len(test.expectedSlice) {
				t.Errorf("Slices differ in size - wanted: %d, found %d", len(test.expectedSlice), len(flagSlice))
			}

			//map iteration is non-deterministic so dont compare values if the buildArgMap is > 1
			if len(test.buildArgMap) <= 1 {
				for i, val := range flagSlice {

					if val != test.expectedSlice[i] {
						t.Errorf("Slices differ in values - wanted: %s, found %s", test.expectedSlice[i], val)
					}

				}
			}

		})
	}

}

func Test_getPackages(t *testing.T) {
	var buildOpts = []struct {
		title                 string
		availableBuildOptions []stack.BuildOption
		requestedBuildOptions []string
		expectedPackages      []string
	}{
		{
			title: "Single Option",
			availableBuildOptions: []stack.BuildOption{
				stack.BuildOption{Name: "dev",
					Packages: []string{"jq", "hw", "ke"}},
			},
			requestedBuildOptions: []string{"dev"},
			expectedPackages:      []string{"jq", "hw", "ke"},
		},
		{
			title: "Two Options one chosen",
			availableBuildOptions: []stack.BuildOption{
				stack.BuildOption{Name: "dev",
					Packages: []string{"jq", "hw", "ke"}},
				stack.BuildOption{Name: "debug",
					Packages: []string{"lr", "kt", "jy"}},
			},
			requestedBuildOptions: []string{"dev"},
			expectedPackages:      []string{"jq", "hw", "ke"},
		},
		{
			title: "Two Options two chosen",
			availableBuildOptions: []stack.BuildOption{
				stack.BuildOption{Name: "dev",
					Packages: []string{"jq", "hw", "ke"}},
				stack.BuildOption{Name: "debug",
					Packages: []string{"lr", "kt", "jy"}},
			},
			requestedBuildOptions: []string{"dev", "debug"},
			expectedPackages:      []string{"jq", "hw", "ke", "lr", "kt", "jy"},
		},
		{
			title: "Two Options two chosen with overlaps",
			availableBuildOptions: []stack.BuildOption{
				stack.BuildOption{Name: "dev",
					Packages: []string{"jq", "hw", "ke"}},
				stack.BuildOption{Name: "debug",
					Packages: []string{"lr", "jq", "hw"}},
			},
			requestedBuildOptions: []string{"dev", "debug"},
			expectedPackages:      []string{"jq", "hw", "ke", "lr"},
		},
	}
	for _, test := range buildOpts {

		t.Run(test.title, func(t *testing.T) {

			buildOptPackages, _ := getPackages(test.availableBuildOptions, test.requestedBuildOptions)

			if len(buildOptPackages) != len(test.expectedPackages) {
				t.Errorf("Slices differ in size - wanted: %d, found %d", len(test.expectedPackages), len(buildOptPackages))
			}
			for _, expectedOptPackage := range test.expectedPackages {
				found := false
				for _, buildOptPackage := range buildOptPackages {
					if buildOptPackage == expectedOptPackage {
						found = true
						break
					}
				}
				if found == false {

					t.Errorf("Slices differ in values  - wanted: %s, found %s", strings.Join(test.expectedPackages, " "), strings.Join(buildOptPackages, " "))
				}
			}
		})
	}
}

func Test_deDuplicate(t *testing.T) {
	var stringOpts = []struct {
		title           string
		inputStrings    []string
		expectedStrings []string
	}{
		{
			title:           "No Duplicates",
			inputStrings:    []string{"jq", "hw", "ke"},
			expectedStrings: []string{"jq", "hw", "ke"},
		},
		{
			title:           "Duplicates",
			inputStrings:    []string{"jq", "hw", "ke", "jq", "hw", "ke"},
			expectedStrings: []string{"jq", "hw", "ke"},
		},
	}
	for _, test := range stringOpts {

		t.Run(test.title, func(t *testing.T) {

			uniqueStrings := deDuplicate(test.inputStrings)

			if len(uniqueStrings) != len(test.expectedStrings) {
				t.Errorf("Slices differ in size - wanted: %d, found %d", len(test.expectedStrings), len(uniqueStrings))
			}

			for _, expectedString := range test.expectedStrings {

				found := false

				for _, uniqueString := range uniqueStrings {

					if expectedString == uniqueString {
						found = true
						break
					}
				}
				if found == false {

					t.Errorf("Slices differ in values  - wanted: %s, found %s", strings.Join(test.expectedStrings, " "), strings.Join(uniqueStrings, " "))
				}
			}
		})
	}
}

/*func Test_validateBuildOption(t *testing.T) {

	buildOptions := []struct {
		buildOption           string
		expectedBuildArgValue string
	}{
		{
			buildOption:           "dev",
			expectedBuildArgValue: "ADDITIONAL_PACKAGE=make automake gcc g++ subversion python3-dev musl-dev libffi-dev",
		},

		{
			buildOption:           "undefined",
			expectedBuildArgValue: "",
		},
	}

	os.MkdirAll("template/python3", os.ModePerm)
	python3_template_yml, err := os.Create("template/python3/template.yml")
	if err != nil {
		t.Errorf("Error creating template/python3/template.yml file")
	}

	_, err = python3_template_yml.WriteString("language: python3\n" +
		"fprocess: python3 index.py\n" +
		"build_options: \n" +
		"  - name: dev\n" +
		"    packages: \n" +
		"      - make\n" +
		"      - automake\n" +
		"      - gcc\n" +
		"      - g++\n" +
		"      - subversion\n" +
		"      - python3-dev\n" +
		"      - musl-dev\n" +
		"      - libffi-dev\n")

	if err != nil {
		t.Errorf("Error writing to template/python3/template.yml file")
	}

	os.MkdirAll("template/unsupported", os.ModePerm)
	unsupported_template_yml, err := os.Create("template/unsupported/template.yml")
	if err != nil {
		t.Errorf("Error creating template/unsupported/template.yml file")
	}

	_, err = unsupported_template_yml.WriteString("language: python3\n" +
		"fprocess: python3 index.py\n")

	if err != nil {
		t.Errorf("Error writing to template/pythunsupportedon3/template.yml file")
	}

	for _, test := range buildOptions {
		t.Run(test.buildOption, func(t *testing.T) {
			res, _, _ := validateBuildOption(test.buildOption, "python3")
			_, isValid, _ := validateBuildOption(test.buildOption, "unsupported")

			if res != test.expectedBuildArgValue {
				t.Errorf("validateBuildOption failed for build-option %s. Expected to return %s, but returned %s",
					test.buildOption, test.expectedBuildArgValue, res)
			}

			if isValid && test.buildOption == "dev" {
				t.Errorf("validateBuildOption failed for build-option %s and unsupported language. Expected validation to fail, but it was successful",
					test.buildOption)
			}
		})
	}

	os.RemoveAll("template")
}
*/
