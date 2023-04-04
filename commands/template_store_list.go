// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"os"
	"sort"
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
	recommended      bool
	official         bool
)

func init() {
	templateStoreListCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Shows additional language and platform")
	templateStoreListCmd.PersistentFlags().StringVarP(&templateStoreURL, "url", "u", DefaultTemplatesStore, "Use as alternative store for templates")
	templateStoreListCmd.Flags().StringVarP(&inputPlatform, "platform", "p", mainPlatform, "Shows the platform if the output is verbose")
	templateStoreListCmd.Flags().BoolVarP(&recommended, "recommended", "r", false, "Shows only recommended templates")
	templateStoreListCmd.Flags().BoolVarP(&official, "official", "o", false, "Shows only official templates")

	templateStoreCmd.AddCommand(templateStoreListCmd)
}

// templateStoreListCmd lists templates from default store or custom store if set
var templateStoreListCmd = &cobra.Command{
	Use:     `list`,
	Short:   `List templates from OpenFaaS organizations`,
	Aliases: []string{"ls"},
	Long: `List templates from a template store manifest file, by default the 
official list maintained by the OpenFaaS community is used. You can override this.`,
	Example: `  faas-cli template store list
  # List only recommended templates
  faas-cli template store list --recommended

  # List only official templates
  faas-cli template store list --official

  # Override the store via a flag
  faas-cli template store ls \
  --url=https://raw.githubusercontent.com/openfaas/store/master/templates.json

  # Specify an alternative store via environment variable
  export OPENFAAS_TEMPLATE_STORE_URL=https://example.com/templates.json

  # See additional language and platform
  faas-cli template store ls --verbose=true

  # Filter by platform for arm64 only
  faas-cli template store list --platform arm64 
`,
	RunE: runTemplateStoreList,
}

func runTemplateStoreList(cmd *cobra.Command, args []string) error {
	envTemplateRepoStore := os.Getenv(templateStoreURLEnvironment)
	storeURL := getTemplateStoreURL(templateStoreURL, envTemplateRepoStore, DefaultTemplatesStore)

	templatesInfo, err := getTemplateInfo(storeURL)
	if err != nil {
		return fmt.Errorf("error while getting templates info: %s", err)
	}
	list := []TemplateInfo{}

	if recommended {
		for i := 0; i < len(templatesInfo); i++ {
			if templatesInfo[i].Recommended {
				list = append(list, templatesInfo[i])
			}
		}
	} else if official {
		for i := 0; i < len(templatesInfo); i++ {
			if templatesInfo[i].Official == "true" {
				list = append(list, templatesInfo[i])
			}
		}
	} else {
		list = templatesInfo
	}

	formattedOutput := formatTemplatesOutput(list, verbose, inputPlatform)

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
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while requesting template list: %s", err.Error())
	}

	if res.Body == nil {
		return nil, fmt.Errorf("error empty response body from: %s", templateStoreURL)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code wanted: %d got: %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %s", err.Error())
	}

	templatesInfo := []TemplateInfo{}
	if err := json.Unmarshal(body, &templatesInfo); err != nil {
		return nil, fmt.Errorf("can't unmarshal text: %s", err.Error())
	}

	sortTemplates(templatesInfo)

	return templatesInfo, nil
}

func sortTemplates(templatesInfo []TemplateInfo) {
	sort.Slice(templatesInfo, func(i, j int) bool {
		if templatesInfo[i].Recommended == templatesInfo[j].Recommended {
			if templatesInfo[i].Official == templatesInfo[j].Official {
				return strings.ToLower(templatesInfo[i].TemplateName) < strings.ToLower(templatesInfo[j].TemplateName)
			} else {
				return templatesInfo[i].Official < templatesInfo[j].Official
			}
		} else if templatesInfo[i].Recommended {
			return true
		} else {
			return false
		}
	})
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

	fmt.Fprintf(lineWriter, "NAME\tRECOMMENDED\tDESCRIPTION\tSOURCE\n")
	for _, template := range templates {

		recommended := "[ ]"
		if template.Recommended {
			recommended = "[x]"
		}

		fmt.Fprintf(lineWriter, "%s\t%s\t%s\t%s\n",
			template.TemplateName,
			recommended,
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
	Recommended  bool   `json:"recommended"`
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
