// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ExecCommand run a system command
func ExecCommand(tempPath string, builder []string) {
	var out bytes.Buffer
	targetCmd := exec.Command(builder[0], builder[1:]...)
	targetCmd.Dir = tempPath
	targetCmd.Stdout = &out
	targetCmd.Stderr = &out
	targetCmd.Run()

	logDir := ""
	if err := os.MkdirAll("logs/", 0755); err == nil {
		logDir = "logs/"
	}

	f, err := os.OpenFile(fmt.Sprintf("%s%s.log", logDir, strings.Replace(builder[3], "/", "-", 1)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("%s\n%s", strings.Join(builder, " "), out.String())
	log.SetOutput(os.Stdout)
}
