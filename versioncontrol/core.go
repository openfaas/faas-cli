// Package versioncontrol is a simplified/stripped down version of go/internal/get/vcs that
// is aimed at the simplier temporary git clone needed for OpenFaaS template fetch.
package versioncontrol

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type vcsCmd struct {
	name string

	// name of binary to invoke command
	cmd string

	// commands to execute with the binary
	cmds []string

	// uri schemes the command can apply to
	scheme []string
}

// Invoke executes the vcsCmd replacing varibables in the cmds with the keyval
// variables passed.
func (v *vcsCmd) Invoke(dir string, args map[string]string) error {
	for _, cmd := range v.cmds {
		if _, err := v.run(dir, cmd, args, true); err != nil {
			return err
		}
	}
	return nil
}

// run is the generalized implementation of executing our commands.
func (v *vcsCmd) run(dir string, cmdline string, keyval map[string]string, verbose bool) ([]byte, error) {
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = replaceVars(keyval, arg)
	}

	// run external command
	_, err := exec.LookPath(v.cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "missing %s command", v.name)
		return nil, err
	}

	cmd := exec.Command(v.cmd, args...)
	cmd.Dir = dir
	cmd.Env = envWithPWD(cmd.Dir)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		if verbose {
			os.Stderr.Write(out)
		}
		return out, err
	}
	return out, nil
}

// replaceVars rewrites a string to replace variables written as {k}
// with the value vars[k] for each key k in vars.
func replaceVars(vars map[string]string, s string) string {
	for key, value := range vars {
		s = strings.Replace(s, "{"+key+"}", value, -1)
	}
	return s
}

// envWithPWD creates a new ENV slice from the existing ENV, updating or adding
// the PWD flag to the specified dir. Our commands are usually given abs paths,
// but just this is set just incase the command is sensitive to the value.
func envWithPWD(dir string) []string {
	env := os.Environ()
	updated := false
	for i, envVar := range env {
		if strings.HasPrefix(envVar, "PWD") {
			env[i] = "PWD=" + dir
			updated = true
		}
	}

	if !updated {
		env = append(env, "PWD="+dir)
	}

	return env
}
