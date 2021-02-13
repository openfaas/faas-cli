// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
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
	tail            bool
	lines           int
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
	Use:   `logs <NAME> [--tls-no-verify] [--gateway] [--output=text/json]`,
	Short: "Fetch logs for a functions",
	Long:  "Fetch logs for a given function name in plain text or JSON format.",
	Example: `  faas-cli logs FN
  faas-cli logs FN --output=json
  faas-cli logs FN --lines=5
  faas-cli logs FN --tail=false --since=10m
  faas-cli logs FN --tail=false --since=2010-01-01T00:00:00Z
`,
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
	cmd.Flags().StringVarP(&functionNamespace, "namespace", "n", "", "Namespace of the function")

	cmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	cmd.Flags().DurationVar(&logFlagValues.since, "since", 0*time.Second, "return logs newer than a relative duration like 5s")
	cmd.Flags().Var(&logFlagValues.sinceTime, "since-time", "include logs since the given timestamp (RFC3339)")
	cmd.Flags().IntVar(&logFlagValues.lines, "lines", -1, "number of recent log lines file to display. Defaults to -1, unlimited if <=0")
	cmd.Flags().BoolVarP(&logFlagValues.tail, "tail", "t", true, "tail logs and continue printing new logs until the end of the request, up to 30s")
	cmd.Flags().StringVarP(&logFlagValues.token, "token", "k", "", "Pass a JWT token to use instead of basic auth")

	logFlagValues.timeFormat = flags.TimeFormat(time.RFC3339)
	cmd.Flags().VarP(&logFlagValues.logFormat, "output", "o", "output logs as (plain|keyvalue|json), JSON includes all available keys")
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
	cliAuth, err := proxy.NewCLIAuth(logFlagValues.token, gatewayAddress)
	if err != nil {
		return err
	}
	transport := getLogStreamingTransport(tlsInsecure)
	cliClient, err := proxy.NewClient(cliAuth, gatewayAddress, transport, nil)
	if err != nil {
		return err
	}

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

	ns, err := cmd.Flags().GetString("namespace")
	if err != nil {
		log.Printf("error getting namespace flag %s\n", err.Error())
	}

	return logs.Request{
		Name:      args[0],
		Namespace: ns,
		Tail:      logFlagValues.lines,
		Since:     sinceValue(logFlagValues.sinceTime.AsTime(), logFlagValues.since),
		Follow:    logFlagValues.tail,
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
