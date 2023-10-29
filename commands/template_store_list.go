// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"context"
	"fmt"

	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

const (
	// DefaultTemplatesStore is the URL where the official store can be found
	DefaultTemplatesStore = "https://raw.githubusercontent.com/openfaas/store/master/templates.json"
	mainPlatform          = "x86_64"
	templateStoreDoc      = `Alternative path to the template store metadata. It may be an http(s) URL or a local path to a JSON file.`
)

var (
	templateStoreURL string
	inputPlatform    string
	recommended      bool
	official         bool
)

func init() {
	templateStoreListCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Shows additional language and platform")
	templateStoreListCmd.PersistentFlags().StringVarP(&templateStoreURL, "url", "u", DefaultTemplatesStore, templateStoreDoc)
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
	templatesInfo := []TemplateInfo{}
	err := proxy.ReadJSON(context.TODO(), repository, &templatesInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot read templates info from: %s", repository)
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

	fmt.Fprintf(lineWriter, "NAME\tRECOMMENDED\tSOURCE\tDESCRIPTION\tLANGUAGE\tPLATFORM\n")
	for _, template := range templates {
		recommended := "[ ]"
		if template.Recommended {
			recommended = "[x]"
		}

		fmt.Fprintf(lineWriter, "%s\t%s\t%s\t%s\t%s\t%s\n",
			template.TemplateName,
			recommended,
			template.Source,
			template.Description,
			template.Language,
			template.Platform)
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
