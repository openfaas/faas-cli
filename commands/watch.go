package commands

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/openfaas/faas-cli/logger"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

// watchLoop will watch for changes to function handler files and the stack.yml
// then call onChange when a change is detected
func watchLoop(cmd *cobra.Command, args []string, onChange func(cmd *cobra.Command, args []string) error) error {
	mainCtx := cmd.Context()

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

	logger.Debugf("[Watch] added: %s\n", yamlPath)
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

	bounce := debounce.New(1500 * time.Millisecond)

	onChangeCtx, onChangeCancel := context.WithCancel(mainCtx)
	defer onChangeCancel()

	// the WaitGroup is used to enable the watch+debounce to easily wait for
	// each onChange invocation to complete or fully cancel before starting
	// the next one. Without this, because the `cmd` is a shared pointer instead
	// of a value, when we changed the onChangeCtx, it would propogate to the
	// currently cancelling onChange invocation. If this handler contains many
	// steps, it would be possible for it continue with the new context.
	// This was seen in the local-run, the build would cancel but not return,
	// so it would try to run the just aborted build and produce errors.
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// An initial build is usually done on first load with
		// live reloaders
		cmd.SetContext(onChangeCtx)
		if err := onChange(cmd, args); err != nil {
			fmt.Println("Error on initial run: ", err)
		}
	}()

	log.Printf("[Watch] Started")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher's Events channel is closed")
			}

			logger.Debugf("[Watch] event: %s on: %s", strings.ToLower(event.Op.String()), event.Name)

			info, trigger := shouldTrigger(event)
			if !trigger {
				continue
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

			// now check if the file is ignored and should not trigger the onChange
			if matcher.Match(strings.Split(event.Name, "/"), info.IsDir()) {
				continue
			}

			if target == "" {
				fmt.Printf("[Watch] Rebuilding %d functions reason: %s to %s\n", len(fnNames), strings.ToLower(event.Op.String()), event.Name)
			} else {
				fmt.Printf("[Watch] Reloading %s reason: %s %s\n", target, strings.ToLower(event.Op.String()), event.Name)
			}

			bounce(func() {
				log.Printf("[Watch] Cancelling")
				onChangeCancel()
				wg.Wait()

				log.Printf("[Watch] Cancelled")
				onChangeCtx, onChangeCancel = context.WithCancel(mainCtx)
				cmd.SetContext(onChangeCtx)

				// Assign --filter to "" for all functions if we can't determine the
				// changed function to direct the calls to build/push/deploy
				filter = target

				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := onChange(cmd, args); err != nil {
						fmt.Println("Error on change: ", err)
					}
				}()
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher's Errors channel is closed")
			}
			return err

		case <-mainCtx.Done():
			watcher.Close()
			return nil
		}
	}
}

// shouldTrigger returns true if the event should trigger a rebuild. This currently
// includes create, write, remove, and rename events.
func shouldTrigger(event fsnotify.Event) (fs.FileInfo, bool) {
	// skip temp and swap files
	if strings.HasSuffix(event.Name, ".swp") || strings.HasSuffix(event.Name, "~") || strings.HasSuffix(event.Name, ".swx") {
		return nil, false
	}

	// only trigger for content changes, this skips chmod, chown, etc.
	if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		info, err := os.Stat(event.Name)
		if err != nil {
			return nil, false
		}

		return info, true
	}

	return nil, false
}

func addPath(watcher *fsnotify.Watcher, rootPath string) error {
	return filepath.WalkDir(rootPath, func(subPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := watcher.Add(subPath); err != nil {
				return fmt.Errorf("unable to watch %s: %s", subPath, err)
			}

			logger.Debugf("[Watch] added: %s\n", subPath)
		}

		return nil
	})

}
