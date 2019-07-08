// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/flags"
	"github.com/openfaas/faas-provider/logs"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

var (
	logFlagValues logFlags
	nowFunc       = time.Now
)

type logFlags struct {
	instance  string
	since     time.Duration
	sinceTime flags.TimestampFlag
	follow    bool
	tail      int
	token     string
}

func init() {
	initLogCmdFlags(functionLogsCmd)

	faasCmd.AddCommand(functionLogsCmd)
}

var functionLogsCmd = &cobra.Command{
	Use:     `logs <NAME> [--tls-no-verify] [--gateway]`,
	Aliases: []string{"ls"},
	Short:   "Tail logs from your functions",
	Long:    "Tail logs from your functions",
	Example: `faas-cli logs echo
faas-cli logs echo --tail=5
faas-cli logs echo --follow=false
faas-cli logs echo --follow=false --since=10m
faas-cli logs echo --follow=false --since=2010-01-01T00:00:00Z`,
	Args:    cobra.ExactArgs(1),
	RunE:    runLogs,
	PreRunE: noopPreRunCmd,
}

func noopPreRunCmd(cmd *cobra.Command, args []string) error {
	return nil
}

// initLogCmdFlags configures the logs command flags, this allows the developer to
// reset and reinitialize the flags on the command in unit tests
func initLogCmdFlags(cmd *cobra.Command) {
	logFlagValues = logFlags{}

	cmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	cmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	cmd.Flags().DurationVar(&logFlagValues.since, "since", 0*time.Second, "return logs newer than a relative duration like 5s")
	cmd.Flags().Var(&logFlagValues.sinceTime, "since-time", "include logs since the given timestamp (RFC3339)")
	cmd.Flags().IntVar(&logFlagValues.tail, "tail", -1, "number of recent log lines file to display. Defaults to -1, unlimited if <=0")
	cmd.Flags().BoolVar(&logFlagValues.follow, "follow", true, "continue printing new logs until the end of the request, up to 30s")
	cmd.Flags().StringVarP(&logFlagValues.token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
}

func runLogs(cmd *cobra.Command, args []string) error {

	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	if msg := checkTLSInsecure(gatewayAddress, tlsInsecure); len(msg) > 0 {
		fmt.Println(msg)
	}

	logRequest := logRequestFromFlags(cmd, args)
	logEvents, err := proxy.GetLogs(gatewayAddress, tlsInsecure, logRequest, logFlagValues.token)
	if err != nil {
		return err
	}

	for logMsg := range logEvents {
		fmt.Fprintln(os.Stdout, strings.TrimRight(logMsg.String(), "\n"))
	}

	return nil
}

func logRequestFromFlags(cmd *cobra.Command, args []string) logs.Request {
	return logs.Request{
		Name:   args[0],
		Tail:   logFlagValues.tail,
		Since:  sinceValue(logFlagValues.sinceTime.AsTime(), logFlagValues.since),
		Follow: logFlagValues.follow,
	}
}

func sinceValue(t time.Time, d time.Duration) *time.Time {
	if !t.IsZero() {
		return &t
	}

	if d.String() != "0s" {
		ts := nowFunc().Add(-1 * d)
		return &ts
	}
	return nil
}
