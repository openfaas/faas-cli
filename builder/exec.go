// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"fmt"
	"os"
	"os/exec"
)

// ExecCommand run a system command
func ExecCommand(tempPath string, builder []string) error {
	targetCmd := exec.Command(builder[0], builder[1:]...)
	targetCmd.Dir = tempPath
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	targetCmd.Start()
	err := targetCmd.Wait()
	if err != nil {
		errString := fmt.Sprintf("%s", builder)
		return fmt.Errorf(errString)
	}
	return nil
}
