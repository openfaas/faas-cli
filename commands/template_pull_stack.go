package commands

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/openfaas/go-sdk/stack"
	"github.com/spf13/cobra"
)

var (
	templateURL    string
	customRepoName string
)

func init() {
	templatePullStackCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing templates?")
	templatePullStackCmd.Flags().BoolVar(&pullDebug, "debug", false, "Enable debug output")

	templatePullCmd.AddCommand(templatePullStackCmd)
}

var templatePullStackCmd = &cobra.Command{
	Use:   `stack`,
	Short: `Downloads templates specified in the function definition yaml file`,
	Long: `Downloads templates specified in the function yaml file, in the current directory
	`,
	Example: `
  faas-cli template pull stack
  faas-cli template pull stack -f myfunction.yml
  faas-cli template pull stack -r custom_repo_name
`,
	RunE: runTemplatePullStack,
}

func runTemplatePullStack(cmd *cobra.Command, args []string) error {
	templatesConfig, err := loadTemplateConfig()
	if err != nil {
		return err
	}
	return pullStackTemplates(templatesConfig, cmd)
}

func loadTemplateConfig() ([]stack.TemplateSource, error) {
	stackConfig, err := readStackConfig()
	if err != nil {
		return nil, err
	}
	return stackConfig.StackConfig.TemplateConfigs, nil
}

func readStackConfig() (stack.Configuration, error) {
	configField := stack.Configuration{}

	configFieldBytes, err := os.ReadFile(yamlFile)
	if err != nil {
		return configField, fmt.Errorf("can't read file %s, error: %s", yamlFile, err.Error())
	}
	if err := yaml.Unmarshal(configFieldBytes, &configField); err != nil {
		return configField, fmt.Errorf("can't read: %s", err.Error())
	}

	if len(configField.StackConfig.TemplateConfigs) == 0 {
		return configField, fmt.Errorf("can't read configuration: no template repos currently configured")
	}
	return configField, nil
}

func pullStackTemplates(templateInfo []stack.TemplateSource, cmd *cobra.Command) error {
	for _, val := range templateInfo {
		fmt.Printf("Pulling template: %s from configuration file: %s\n", val.Name, yamlFile)
		if len(val.Source) == 0 {
			pullErr := runTemplateStorePull(cmd, []string{val.Name})
			if pullErr != nil {
				return pullErr
			}
		} else {

			templateName := val.Name
			pullErr := pullTemplate(val.Source, templateName)
			if pullErr != nil {
				return pullErr
			}
		}
	}
	return nil
}

// filter templates which are already available on filesystem
func filterExistingTemplates(templateInfo []stack.TemplateSource, templatesDir string) ([]stack.TemplateSource, error) {
	var newTemplates []stack.TemplateSource
	for _, info := range templateInfo {
		templatePath := fmt.Sprintf("%s/%s", templatesDir, info.Name)
		if _, err := os.Stat(templatePath); err != nil && os.IsNotExist(err) {
			newTemplates = append(newTemplates, info)
		}
	}

	return newTemplates, nil
}
