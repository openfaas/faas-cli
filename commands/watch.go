package commands

import (
	"context"
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
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

// watchLoop will watch for changes to function handler files and the stack.yml
// then call onChange when a change is detected
func watchLoop(cmd *cobra.Command, args []string, onChange func(cmd *cobra.Command, args []string, ctx context.Context) error) error {

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

	canceller := Cancel{}

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

	// map to determine which function belongs to changed files
	// when responding to events
	handlerMap := make(map[string]string)

	for serviceName, service := range services.Functions {
		handlerMap[serviceName] = path.Join(cwd, service.Handler)

		handlerFullPath := path.Join(cwd, service.Handler)

		if err := addPath(watcher, handlerFullPath); err != nil {
			return err
		}
	}

	signalChannel := make(chan os.Signal, 1)

	// Exit on Ctrl+C or kill
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	bounce := debounce.New(1500 * time.Millisecond)

	go func() {
		// An initial build is usually done on first load with
		// live reloaders
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		canceller.Set(ctx, cancel)

		if err := onChange(cmd, args, ctx); err != nil {
			fmt.Println("Error rebuilding: ", err)
			os.Exit(1)
		}
	}()

	log.Printf("[Watch] Started")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher's Events channel is closed")
			}

			if debug == "1" {
				log.Printf("[Watch] event: %s on: %s", strings.ToLower(event.Op.String()), event.Name)
			}
			if strings.HasSuffix(event.Name, ".swp") || strings.HasSuffix(event.Name, "~") || strings.HasSuffix(event.Name, ".swx") {
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

				// fuzzy match after, if none matched exactly
				if target == "" {
					for fnName, fnPath := range handlerMap {

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
						fmt.Printf("[Watch] Rebuilding %d functions reason: %s to %s\n", len(fnNames), strings.ToLower(event.Op.String()), event.Name)
					} else {
						fmt.Printf("[Watch] Reloading %s reason: %s %s\n", target, strings.ToLower(event.Op.String()), event.Name)
					}

					bounce(func() {
						log.Printf("[Watch] Cancelling")

						canceller.Cancel()

						log.Printf("[Watch] Cancelled")
						ctx, cancel := context.WithCancel(context.Background())
						canceller.Set(ctx, cancel)

						// Assign --filter to "" for all functions if we can't determine the
						// changed function to direct the calls to build/push/deploy
						filter = target

						go func() {
							if err := onChange(cmd, args, ctx); err != nil {
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

// Cancel is a struct to hold a reference to a context and
// cancellation function between closures
type Cancel struct {
	cancel context.CancelFunc
	ctx    context.Context
}

func (c *Cancel) Set(ctx context.Context, cancel context.CancelFunc) {
	c.cancel = cancel
	c.ctx = ctx
}

func (c *Cancel) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}

}
