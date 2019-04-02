// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const (
	// DefaultTemplatesStore is the URL where the official store can be found
	DefaultTemplatesStore = "https://raw.githubusercontent.com/openfaas/store/master/templates.json"
	mainPlatform          = "x86_64"
)

var (
	templateStoreURL string
	inputPlatform    string
)

func init() {
	templateStoreListCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Shows additional language and platform")
	templateStoreListCmd.PersistentFlags().StringVarP(&templateStoreURL, "url", "u", DefaultTemplatesStore, "Use as alternative store for templates")
	templateStoreListCmd.Flags().StringVarP(&inputPlatform, "platform", "p", mainPlatform, "Shows the platform if the output is verbose")

	templateStoreCmd.AddCommand(templateStoreListCmd)
}

// templateStoreListCmd lists templates from default store or custom store if set
var templateStoreListCmd = &cobra.Command{
	Use:     `list`,
	Short:   `List templates from OpenFaaS organizations`,
	Aliases: []string{"ls"},
	Long:    `List templates from official store or from custom URL or set the environmental variable OPENFAAS_TEMPLATE_STORE_URL to be the default store location`,
	Example: `  faas-cli template store list
  faas-cli template store ls
  faas-cli template store ls --url=https://raw.githubusercontent.com/openfaas/store/master/templates.json
  faas-cli template store ls --verbose=true
  faas-cli template store list --platform arm64`,
	RunE: runTemplateStoreList,
}

func runTemplateStoreList(cmd *cobra.Command, args []string) error {
	envTemplateRepoStore := os.Getenv(templateStoreURLEnvironment)
	storeURL := getTemplateStoreURL(templateStoreURL, envTemplateRepoStore, DefaultTemplatesStore)

	templatesInfo, templatesErr := getTemplateInfo(storeURL)
	if templatesErr != nil {
		return fmt.Errorf("error while getting templates info: %s", templatesErr)
	}

	formattedOutput := formatTemplatesOutput(templatesInfo, verbose, inputPlatform)

	fmt.Fprintf(cmd.OutOrStdout(), "%s", formattedOutput)

	return nil
}

func getTemplateInfo(repository string) ([]TemplateInfo, error) {
	req, reqErr := http.NewRequest(http.MethodGet, repository, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("error while trying to create request to take template info: %s", reqErr.Error())
	}

	reqContext, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	req = req.WithContext(reqContext)

	client := http.DefaultClient
	res, clientErr := client.Do(req)
	if clientErr != nil {
		return nil, fmt.Errorf("error while requesting template list: %s", clientErr.Error())
	}

	if res.Body == nil {
		return nil, fmt.Errorf("error empty response body from: %s", templateStoreURL)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code wanted: %d got: %d", http.StatusOK, res.StatusCode)
	}

	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf("error while reading data from templates body: %s", bodyErr.Error())
	}

	templatesInfo := []TemplateInfo{}
	unmarshallErr := json.Unmarshal(body, &templatesInfo)
	if unmarshallErr != nil {
		return nil, fmt.Errorf("error while unmarshalling into templates struct: %s", unmarshallErr.Error())
	}
	return templatesInfo, nil
}

func formatTemplatesOutput(templates []TemplateInfo, verbose bool, platform string) string {

	if platform != mainPlatform {
		templates = filterTemplate(templates, platform)
	} else {
		templates = filterTemplate(templates, mainPlatform)
	}

	if len(templates) == 0 {
		return ""
	}

	var buff bytes.Buffer
	lineWriter := tabwriter.NewWriter(&buff, 0, 0, 1, ' ', 0)

	fmt.Fprintln(lineWriter)
	if verbose {
		formatVerboseOutput(lineWriter, templates)
	} else {
		formatBasicOutput(lineWriter, templates)
	}
	fmt.Fprintln(lineWriter)

	lineWriter.Flush()

	return buff.String()
}

func formatBasicOutput(lineWriter *tabwriter.Writer, templates []TemplateInfo) {

	fmt.Fprintf(lineWriter, "NAME\tSOURCE\tDESCRIPTION\n")
	for _, template := range templates {
		fmt.Fprintf(lineWriter, "%s\t%s\t%s\n",
			template.TemplateName,
			template.Source,
			template.Description)
	}
}

func formatVerboseOutput(lineWriter *tabwriter.Writer, templates []TemplateInfo) {

	fmt.Fprintf(lineWriter, "NAME\tLANGUAGE\tPLATFORM\tSOURCE\tDESCRIPTION\n")
	for _, template := range templates {
		fmt.Fprintf(lineWriter, "%s\t%s\t%s\t%s\t%s\n",
			template.TemplateName,
			template.Language,
			template.Platform,
			template.Source,
			template.Description)
	}
}

// TemplateInfo is the definition of a template which is part of the store
type TemplateInfo struct {
	TemplateName string `json:"template"`
	Platform     string `json:"platform"`
	Language     string `json:"language"`
	Source       string `json:"source"`
	Description  string `json:"description"`
	Repository   string `json:"repo"`
	Official     string `json:"official"`
}

func filterTemplate(templates []TemplateInfo, platform string) []TemplateInfo {
	var filteredTemplates []TemplateInfo

	for _, template := range templates {
		if strings.EqualFold(template.Platform, platform) {
			filteredTemplates = append(filteredTemplates, template)
		}
	}
	return filteredTemplates
}
