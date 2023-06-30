// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bep/debounce"
	"github.com/openfaas/faas-cli/stack"

	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	skipPush   bool
	skipDeploy bool
	watch      bool
)

func init() {

	upFlagset := pflag.NewFlagSet("up", pflag.ExitOnError)
	upFlagset.BoolVar(&skipPush, "skip-push", false, "Skip pushing function to remote registry")
	upFlagset.BoolVar(&skipDeploy, "skip-deploy", false, "Skip function deployment")
	upFlagset.StringVar(&remoteBuilder, "remote-builder", "", "URL to the builder")
	upFlagset.StringVar(&payloadSecretPath, "payload-secret", "", "Path to payload secret file")

	upFlagset.BoolVar(&watch, "watch", false, "Watch for changes in files and re-deploy")
	upCmd.Flags().AddFlagSet(upFlagset)

	build, _, _ := faasCmd.Find([]string{"build"})
	upCmd.Flags().AddFlagSet(build.Flags())

	push, _, _ := faasCmd.Find([]string{"push"})
	upCmd.Flags().AddFlagSet(push.Flags())

	deploy, _, _ := faasCmd.Find([]string{"deploy"})
	upCmd.Flags().AddFlagSet(deploy.Flags())

	faasCmd.AddCommand(upCmd)
}

// upCmd is a wrapper to the build, push and deploy commands
var upCmd = &cobra.Command{
	Use:   `up -f [YAML_FILE] [--skip-push] [--skip-deploy] [flags from build, push, deploy]`,
	Short: "Builds, pushes and deploys OpenFaaS function containers",
	Long: `Build, Push, and Deploy OpenFaaS function containers either via the
supplied YAML config using the "--yaml" flag (which may contain multiple function
definitions), or directly via flags.

The push step may be skipped by setting the --skip-push flag
and the deploy step with --skip-deploy.

Note: All flags from the build, push and deploy flags are valid and can be combined,
see the --help text for those commands for details.`,
	Example: `  faas-cli up -f myfn.yaml
faas-cli up --filter "*gif*" --secret dockerhuborg`,
	PreRunE: preRunUp,
	RunE:    upHandler,
}

func preRunUp(cmd *cobra.Command, args []string) error {
	if err := preRunBuild(cmd, args); err != nil {
		return err
	}
	if err := preRunDeploy(cmd, args); err != nil {
		return err
	}
	return nil
}

func upHandler(cmd *cobra.Command, args []string) error {
	handler := upRunner(cmd, args)

	// Always run an initial build to freshen up
	if err := handler(); err != nil {
		return err
	}

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	fnNames := []string{}
	for name, _ := range services.Functions {
		fnNames = append(fnNames, name)
	}

	fmt.Printf("[Watch] monitoring %d functions: %s\n", len(fnNames), strings.Join(fnNames, ", "))

	if watch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()

		patterns, err := ignorePatterns()
		if err != nil {
			return err
		}

		matcher := gitignore.NewMatcher(patterns)

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		yamlPath := path.Join(cwd, yamlFile)

		debug := os.Getenv("FAAS_DEBUG")
		if debug == "1" {
			fmt.Printf("[Watch] added: %s\n", yamlPath)
		}
		watcher.Add(yamlPath)

		handlerMap := make(map[string]string)

		for serviceName, service := range services.Functions {
			handlerMap[serviceName] = path.Join(cwd, service.Handler)
		}

		for _, service := range services.Functions {
			handlerPath := path.Join(cwd, service.Handler)

			if err := addPath(watcher, handlerPath); err != nil {
				return err
			}
		}

		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		d := debounce.New(2 * time.Second)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return fmt.Errorf("watcher's Events channel is closed")
				}
				log.Printf("%s %s", event.Op, event.Name)

				if strings.HasSuffix(event.Name, ".swp") || strings.HasSuffix(event.Name, "~") {
					continue
				}

				if event.Op == fsnotify.Write || event.Op == fsnotify.Create || event.Op == fsnotify.Remove || event.Op == fsnotify.Rename {

					info, err := os.Stat(event.Name)
					if err != nil {
						continue
					}
					ignore := false
					if matcher.Match(strings.Split(event.Name, "/"), info.IsDir()) {
						ignore = true
					}

					// exact match first
					target := ""
					for fnName, fnPath := range handlerMap {
						if event.Name == fnPath {
							target = fnName
						}
					}

					// fuzzy match after
					if target == "" {
						for fnName, fnPath := range handlerMap {
							log.Printf("Checking %s against %s", event.Name, fnPath)

							if strings.HasPrefix(event.Name, fnPath) {
								target = fnName
							}
						}
					}

					// New sub-directory added for a function, start tracking it
					if event.Op == fsnotify.Create && info.IsDir() && target != "" {
						if err := addPath(watcher, event.Name); err != nil {
							return err
						}
					}

					if !ignore {
						if target == "" {
							fmt.Printf("Rebuilding %d functions reason: %s to %s\n", len(fnNames), event.Op, event.Name)
						} else {
							fmt.Printf("Reloading %s reason: %s to %s\n", target, event.Op, event.Name)
						}
					}

					if !ignore {
						d(func() {
							filter = target

							go func() {
								// Assign --filter to the function name or "" for all functions
								// in the stack.yml file

								if err := handler(); err != nil {
									fmt.Println("Error rebuilding: ", err)
									os.Exit(1)
								}
							}()
						})
					}

				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return fmt.Errorf("watcher's Errors channel is closed")
				}
				return err

			case <-signalChannel:
				watcher.Close()
				return nil
			}
		}
	}
	return nil
}

func addPath(watcher *fsnotify.Watcher, rootPath string) error {
	debug := os.Getenv("FAAS_DEBUG")

	return filepath.WalkDir(rootPath, func(subPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := watcher.Add(subPath); err != nil {
				return fmt.Errorf("unable to watch %s: %s", subPath, err)
			}

			if debug == "1" {
				fmt.Printf("[Watch] added: %s\n", subPath)
			}
		}

		return nil
	})

}

func upRunner(cmd *cobra.Command, args []string) func() error {
	return func() error {
		if err := runBuild(cmd, args); err != nil {
			return err
		}

		if !skipPush && remoteBuilder == "" {
			if err := runPush(cmd, args); err != nil {
				return err
			}
		}

		if !skipDeploy {
			if err := runDeploy(cmd, args); err != nil {
				return err
			}

			if watch {
				fmt.Println("[Watch] Change a file to trigger a rebuild...")
			}

		}

		return nil
	}
}

func ignorePatterns() ([]gitignore.Pattern, error) {
	gitignorePath := ".gitignore"

	file, err := os.Open(gitignorePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	patterns := []gitignore.Pattern{gitignore.ParsePattern(".git", nil)}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}
