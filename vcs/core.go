// Package vcs is a simplified/stripped down version of go/internal/get/vcs that
// is aimed at the simplier temporary git clone needed for OpenFaaS template fetch.
package vcs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// following utils were copied and modified from
// https://github.com/golang/go/blob/master/src/cmd/go/internal/get/vcs.go

type vcsCmd struct {
	name string
	cmd  string // name of binary to invoke command

	createCmd []string // commands to download a fresh copy of a repository

	scheme  []string
	pingCmd string
}

// ping pings to determine scheme to use.
func (v *vcsCmd) Ping(scheme, repo string) error {
	return v.runVerboseOnly(".", v.pingCmd, "scheme", scheme, "repo", repo)
}

// create creates a new copy of repo in dir.
// The parent of dir must exist; dir must not.
func (v *vcsCmd) Create(dir, repo string) error {
	for _, cmd := range v.createCmd {
		if err := v.run(".", cmd, "dir", dir, "repo", repo); err != nil {
			return err
		}
	}
	return nil
}

func (v *vcsCmd) run(dir string, cmd string, keyval ...string) error {
	_, err := v.run1(dir, cmd, keyval, true)
	return err
}

// runVerboseOnly is like run but only generates error output to standard error in verbose mode.
func (v *vcsCmd) runVerboseOnly(dir string, cmd string, keyval ...string) error {
	_, err := v.run1(dir, cmd, keyval, false)
	return err
}

// run1 is the generalized implementation of run and runOutput.
func (v *vcsCmd) run1(dir string, cmdline string, keyval []string, verbose bool) ([]byte, error) {
	m := make(map[string]string)
	for i := 0; i < len(keyval); i += 2 {
		m[keyval[i]] = keyval[i+1]
	}
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = expand(m, arg)
	}

	// add `-go-internal-mkdir {dir}` to use golang to create the dir
	if len(args) >= 2 && args[0] == "-go-internal-mkdir" {
		var err error
		if filepath.IsAbs(args[1]) {
			err = os.Mkdir(args[1], os.ModePerm)
		} else {
			err = os.Mkdir(filepath.Join(dir, args[1]), os.ModePerm)
		}
		if err != nil {
			return nil, err
		}
		args = args[2:]
	}

	// add `-go-internal-cd {dir}` to use golang to change directories
	if len(args) >= 2 && args[0] == "-go-internal-cd" {
		if filepath.IsAbs(args[1]) {
			dir = args[1]
		} else {
			dir = filepath.Join(dir, args[1])
		}
		args = args[2:]
	}

	// run external command
	_, err := exec.LookPath(v.cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"go: missing %s command. See https://golang.org/s/gogetcmd\n",
			v.name)
		return nil, err
	}

	cmd := exec.Command(v.cmd, args...)
	cmd.Dir = dir
	cmd.Env = envForDir(cmd.Dir, os.Environ())

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "# cd %s; %s %s\n", dir, v.cmd, strings.Join(args, " "))
			os.Stderr.Write(out)
		}
		return out, err
	}
	return out, nil
}

// expand rewrites s to replace {k} with match[k] for each key k in match.
func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

// following utils were copied from
// https://github.com/golang/go/blob/master/src/cmd/go/internal/base/env.go

// EnvForDir returns a copy of the environment
// suitable for running in the given directory.
// The environment is the current process's environment
// but with an updated $PWD, so that an os.Getwd in the
// child will be faster.
func envForDir(dir string, base []string) []string {
	// Internally we only use rooted paths, so dir is rooted.
	// Even if dir is not rooted, no harm done.
	return mergeEnvLists([]string{"PWD=" + dir}, base)
}

// MergeEnvLists merges the two environment lists such that
// variables with the same name in "in" replace those in "out".
// This always returns a newly allocated slice.
func mergeEnvLists(in, out []string) []string {
	out = append([]string(nil), out...)

	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue
			}
		}
		out = append(out, inkv)
	}
	return out
}
