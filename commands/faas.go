// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/docker/docker/pkg/term"
	"github.com/openfaas/faas-cli/version"
	"github.com/spf13/cobra"
)

const (
	defaultGateway       = "http://127.0.0.1:8080"
	defaultNetwork       = ""
	defaultYAML          = "stack.yml"
	defaultSchemaVersion = "1.0"
)

// Flags that are to be added to all commands.
var (
	yamlFile string
	regex    string
	filter   string
)

// Flags that are to be added to subset of commands.
var (
	fprocess     string
	functionName string
	handlerDir   string
	network      string
	gateway      string
	handler      string
	image        string
	imagePrefix  string
	language     string
	tlsInsecure  bool
)

var stat = func(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// TODO: remove this workaround once these vars are no longer global
func resetForTest() {
	yamlFile = ""
	regex = ""
	filter = ""
	version.Version = ""
	shortVersion = false
	appendFile = ""
}

func init() {
	// Setup terminal std
	term.StdStreams()

	faasCmd.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")
	faasCmd.PersistentFlags().StringVarP(&regex, "regex", "", "", "Regex to match with function names in YAML file")
	faasCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "Wildcard to match with function names in YAML file")

	// Set Bash completion options
	validYAMLFilenames := []string{"yaml", "yml"}
	_ = faasCmd.PersistentFlags().SetAnnotation("yaml", cobra.BashCompFilenameExt, validYAMLFilenames)
}

func Execute(customArgs []string) {
	checkAndSetDefaultYaml()

	faasCmd.SilenceUsage = true
	faasCmd.SilenceErrors = true
	faasCmd.SetArgs(customArgs[1:])

	args1 := os.Args[1:]
	cmd1, _, _ := faasCmd.Find(args1)

	plugins, err := getPlugins()
	if err != nil {
		log.Fatal(err)
	}

	if cmd1 != nil && len(args1) > 0 {

		found := ""
		for _, plugin := range plugins {
			if path.Base(plugin) == args1[0] {
				found = plugin
			}
		}
		if len(found) > 0 {

			// if we have found the plugin then sysexec it by replacing current process.
			if err := syscall.Exec(found, append([]string{found}, os.Args[2:]...), os.Environ()); err != nil {
				fmt.Fprintf(os.Stderr, "Error from plugin: %v", err)
				os.Exit(127)
			}
			return
		}
	}

	if err := faasCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}

func checkAndSetDefaultYaml() {
	// Check if there is a default yaml file and set it
	if _, err := stat(defaultYAML); err == nil {
		yamlFile = defaultYAML
	}
}

// faasCmd is the faas-cli root command and mimics the legacy client behaviour
// Every other command attached to FaasCmd is a child command to it.
var faasCmd = &cobra.Command{
	Use:   "faas-cli",
	Short: "Manage your OpenFaaS functions from the command line",
	Long: `
Manage your OpenFaaS functions from the command line`,
	Run: runFaas,
}

// runFaas TODO
func runFaas(cmd *cobra.Command, args []string) {
	printLogo()
	cmd.Help()
}

func getPlugins() ([]string, error) {
	plugins := []string{}
	pluginHome := os.ExpandEnv("$HOME/.openfaas/plugins")

	if _, err := os.Stat(pluginHome); err != nil && os.IsNotExist(err) {
		return plugins, nil
	}

	res, err := os.ReadDir(pluginHome)
	if err != nil {
		return nil, err
	}

	for _, file := range res {
		plugins = append(plugins, path.Join(pluginHome, file.Name()))
	}

	return plugins, nil
}
