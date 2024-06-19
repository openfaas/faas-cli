package execute

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type ExecTask struct {
	// Command is the command to execute. This can be the path to an executable
	// or the executable with arguments. The arguments are detected by looking for
	// a space.
	//
	// Examples:
	//  - Just a binary executable: `/bin/ls`
	//  - Binary executable with arguments: `/bin/ls -la /`
	Command string
	// Args are the arguments to pass to the command. These are ignored if the
	// Command contains arguments.
	Args []string
	// Shell run the command in a bash shell.
	// Note that the system must have `/bin/bash` installed.
	Shell bool
	// Env is a list of environment variables to add to the current environment,
	// these are used to override any existing environment variables.
	Env []string
	// Cwd is the working directory for the command
	Cwd string

	// Stdin connect a reader to stdin for the command
	// being executed.
	Stdin io.Reader

	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs.
	StreamStdio bool

	// PrintCommand prints the command before executing
	PrintCommand bool
}

type ExecResult struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Cancelled bool
}

func (et ExecTask) Execute(ctx context.Context) (ExecResult, error) {
	argsSt := ""
	if len(et.Args) > 0 {
		argsSt = strings.Join(et.Args, " ")
	}

	if et.PrintCommand {
		fmt.Println("exec: ", et.Command, argsSt)
	}

	// don't try to run if the context is already cancelled
	if ctx.Err() != nil {
		return ExecResult{
			// the exec package returns -1 for cancelled commands
			ExitCode:  -1,
			Cancelled: ctx.Err() == context.Canceled,
		}, ctx.Err()
	}

	var command string
	var commandArgs []string
	if et.Shell {
		command = "/bin/bash"
		if len(et.Args) == 0 {
			// use Split and Join to remove any extra whitespace?
			startArgs := strings.Split(et.Command, " ")
			script := strings.Join(startArgs, " ")
			commandArgs = append([]string{"-c"}, script)

		} else {
			script := strings.Join(et.Args, " ")
			commandArgs = append([]string{"-c"}, fmt.Sprintf("%s %s", et.Command, script))
		}
	} else {
		if strings.Contains(et.Command, " ") {
			parts := strings.Split(et.Command, " ")
			command = parts[0]
			commandArgs = parts[1:]
		} else {
			command = et.Command
			commandArgs = et.Args
		}
	}

	cmd := exec.CommandContext(ctx, command, commandArgs...)
	cmd.Dir = et.Cwd

	if len(et.Env) > 0 {
		overrides := map[string]bool{}
		for _, env := range et.Env {
			key := strings.Split(env, "=")[0]
			overrides[key] = true
			cmd.Env = append(cmd.Env, env)
		}

		for _, env := range os.Environ() {
			key := strings.Split(env, "=")[0]

			if _, ok := overrides[key]; !ok {
				cmd.Env = append(cmd.Env, env)
			}
		}
	}
	if et.Stdin != nil {
		cmd.Stdin = et.Stdin
	}

	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}

	var stdoutWriters io.Writer
	var stderrWriters io.Writer

	if et.StreamStdio {
		stdoutWriters = io.MultiWriter(os.Stdout, &stdoutBuff)
		stderrWriters = io.MultiWriter(os.Stderr, &stderrBuff)
	} else {
		stdoutWriters = &stdoutBuff
		stderrWriters = &stderrBuff
	}

	cmd.Stdout = stdoutWriters
	cmd.Stderr = stderrWriters

	startErr := cmd.Start()
	if startErr != nil {
		return ExecResult{}, startErr
	}

	exitCode := 0
	execErr := cmd.Wait()
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return ExecResult{
		Stdout:    stdoutBuff.String(),
		Stderr:    stderrBuff.String(),
		ExitCode:  exitCode,
		Cancelled: ctx.Err() == context.Canceled,
	}, ctx.Err()
}
