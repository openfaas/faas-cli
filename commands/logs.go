// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
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
	instance        string
	since           time.Duration
	sinceTime       flags.TimestampFlag
	follow          bool
	tail            int
	token           string
	logFormat       flags.LogFormat
	includeName     bool
	includeInstance bool
	timeFormat      flags.TimeFormat
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
	Args:    cobra.MaximumNArgs(1),
	RunE:    runLogs,
	PreRunE: noopPreRunCmd,
}

func noopPreRunCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("function name is required")
	}
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

	logFlagValues.timeFormat = flags.TimeFormat(time.RFC3339)
	cmd.Flags().Var(&logFlagValues.logFormat, "format", "output format.  Note that JSON format will always include all log message keys (plain|key-value|json)")
	cmd.Flags().Var(&logFlagValues.timeFormat, "time-format", "string format for the timestamp, any value go time format string is allowed, empty will not print the timestamp")
	cmd.Flags().BoolVar(&logFlagValues.includeName, "name", false, "print the function name")
	cmd.Flags().BoolVar(&logFlagValues.includeInstance, "instance", false, "print the function instance name/id")
}

func runLogs(cmd *cobra.Command, args []string) error {

	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	if msg := checkTLSInsecure(gatewayAddress, tlsInsecure); len(msg) > 0 {
		fmt.Println(msg)
	}

	logRequest := logRequestFromFlags(cmd, args)
	cliAuth := NewCLIAuth(logFlagValues.token, gatewayAddress)
	transport := getLogStreamingTransport(tlsInsecure)
	cliClient := proxy.NewClient(cliAuth, gatewayAddress, transport, nil)
	logEvents, err := cliClient.GetLogs(context.Background(), logRequest)
	if err != nil {
		return err
	}

	formatter := GetLogFormatter(string(logFlagValues.logFormat))
	for logMsg := range logEvents {
		fmt.Fprintln(os.Stdout, formatter(logMsg, logFlagValues.timeFormat.String(), logFlagValues.includeName, logFlagValues.includeInstance))
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

func getLogStreamingTransport(tlsInsecure bool) http.RoundTripper {
	if tlsInsecure {
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: tlsInsecure}

		return tr
	}
	return nil
}
