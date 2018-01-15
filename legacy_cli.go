// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func translateLegacyOpts(args []string) ([]string, error) {

	legacyOptMapping := map[string]string{
		"-handler":  "--handler",
		"-image":    "--image",
		"-name":     "--name",
		"-gateway":  "--gateway",
		"-fprocess": "--fprocess",
		"-lang":     "--lang",
		"-replace":  "--replace",
		"-no-cache": "--no-cache",
		"-yaml":     "--yaml",
		"-squash":   "--squash",
	}

	validActions := map[string]string{
		"build":  "build",
		"delete": "remove",
		"deploy": "deploy",
		"push":   "push",
	}

	action := ""
	translatedArgs := []string{args[0]}
	optsCache := args[1:]

	// Replace action
	for idx, opt := range optsCache {
		if opt == "-version" {
			translatedArgs = append(translatedArgs, "version")
			optsCache = append(optsCache[:idx], optsCache[idx+1:]...)
			action = "version"
		}
		if opt == "-action" {
			if len(optsCache) == idx+1 {
				return []string{""}, fmt.Errorf("no action supplied after deprecated -action flag")
			}
			if translated, ok := validActions[optsCache[idx+1]]; ok {
				translatedArgs = append(translatedArgs, translated)
				optsCache = append(optsCache[:idx], optsCache[idx+2:]...)
				action = translated
			} else {
				return []string{""}, fmt.Errorf("unknown action supplied to deprecated -action flag: %s", optsCache[idx+1])
			}
		}
		if strings.HasPrefix(opt, "-action"+"=") {
			s := strings.SplitN(opt, "=", 2)
			if len(s[1]) == 0 {
				return []string{""}, fmt.Errorf("no action supplied after deprecated -action= flag")
			}
			if translated, ok := validActions[s[1]]; ok {
				translatedArgs = append(translatedArgs, translated)
				optsCache = append(optsCache[:idx], optsCache[idx+1:]...)
				action = translated
			} else {
				return []string{""}, fmt.Errorf("unknown action supplied to deprecated -action= flag: %s", s[1])
			}
		}
	}
	for idx, arg := range optsCache {
		if action == "remove" {
			if arg == "-name" {
				optsCache = append(optsCache[:idx], optsCache[idx+1:]...)
				continue
			}
		}
		if translated, ok := legacyOptMapping[arg]; ok {
			optsCache[idx] = translated
		}
		for legacyOpt, translated := range legacyOptMapping {
			if strings.HasPrefix(arg, legacyOpt) {
				optsCache[idx] = strings.Replace(arg, legacyOpt, translated, 1)
			}
		}
	}

	translatedArgs = append(translatedArgs, optsCache...)

	if !reflect.DeepEqual(args, translatedArgs) {
		fmt.Fprintln(os.Stderr, "Found deprecated go-style flags in command, translating to new format:")
		fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(translatedArgs, " "))
	}

	return translatedArgs, nil
}
