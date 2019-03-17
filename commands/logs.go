// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

var (
	nowFunc = time.Now
)

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
	cmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	cmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	cmd.Flags().String("instance", "", "filter to a specific function instance")
	cmd.Flags().String("since", "", "include logs since the given timestamp (RFC3339)")
	cmd.Flags().String("pattern", "", "filter logs that matching the given pattern")
	cmd.Flags().Bool("invert", false, "invert the pattern match")
	cmd.Flags().Bool("follow", true, "tail logs")
	cmd.Flags().Int("limit", 0, "maximum number of log entries to return, unlimited if <=0 ")
}

func runLogs(cmd *cobra.Command, args []string) error {

	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	if msg := checkTLSInsecure(gatewayAddress, tlsInsecure); len(msg) > 0 {
		fmt.Println(msg)
	}

	logRequest := logRequestFromFlags(cmd, args)

	logEvents, err := proxy.GetLogs(gatewayAddress, tlsInsecure, logRequest)
	if err != nil {
		return err
	}

	for logMsg := range logEvents {
		fmt.Fprintln(os.Stdout, logMsg.String())
	}

	return nil
}

func logRequestFromFlags(cmd *cobra.Command, args []string) schema.LogRequest {
	flags := cmd.Flags()
	return schema.LogRequest{
		Name:     args[0],
		Instance: mustString(flags.GetString("instance")),
		Limit:    mustInt(flags.GetInt("limit")),
		Since:    mustTimestampP(flags.GetString("since")),
		Follow:   mustBool(flags.GetBool("follow")),
		Pattern:  mustStringP(flags.GetString("pattern")),
		Invert:   mustBool(flags.GetBool("invert")),
	}
}

func mustString(v string, e error) string {
	if e != nil {
		panic(e)
	}
	return v
}

func mustStringP(v string, e error) *string {
	if e != nil {
		panic(e)
	}
	if v == "" {
		return nil
	}

	return &v
}

func mustBool(v bool, e error) bool {
	if e != nil {
		panic(e)
	}
	return v
}

func mustInt(v int, e error) int {
	if e != nil {
		panic(e)
	}
	return v
}

// return timestamp from a string flag, if it is not a valid duration, then we
// attempt to parse the string as RFC3339
func mustTimestampP(v string, e error) *time.Time {
	if e != nil {
		panic(e)
	}

	if v == "" {
		return nil
	}

	d, err := time.ParseDuration(v)
	if err == nil {
		ts := nowFunc().Add(-1 * d)
		return &ts
	}

	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		panic(e)
	}

	return &ts
}

// func mustTimestamp(v string, e error)
