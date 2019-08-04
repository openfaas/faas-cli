package commands

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	templateURL    string
	customRepoName string
)

func init() {
	templatePullStackCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing templates?")
	templatePullStackCmd.Flags().BoolVar(&pullDebug, "debug", false, "Enable debug output")
	templatePullStackCmd.PersistentFlags().StringVarP(&customRepoName, "repo", "r", "", "The custom name of the template repo")

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
	if len(customRepoName) > 0 {
		return pullSpecificTemplate(templatesConfig, customRepoName, cmd)
	}
	return pullAllTemplates(templatesConfig, cmd)
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

	configFieldBytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return configField, fmt.Errorf("Error while reading files %s", err.Error())
	}
	unmarshallErr := yaml.Unmarshal(configFieldBytes, &configField)
	if unmarshallErr != nil {
		return configField, fmt.Errorf("Error while reading configuration: %s", err.Error())
	}
	if len(configField.StackConfig.TemplateConfigs) == 0 {
		return configField, fmt.Errorf("Error while reading configuration: no template repos currently configured")
	}
	return configField, nil
}

func pullAllTemplates(templateInfo []stack.TemplateSource, cmd *cobra.Command) error {
	for _, val := range templateInfo {
		fmt.Printf("Pulling template: %s from configuration file: %s\n", val.Name, yamlFile)
		if len(val.Source) == 0 {
			pullErr := runTemplateStorePull(cmd, []string{val.Name})
			if pullErr != nil {
				return pullErr
			}
		} else {
			pullErr := pullTemplate(val.Source)
			if pullErr != nil {
				return pullErr
			}
		}
	}
	return nil
}

func findTemplate(templateInfo []stack.TemplateSource, customName string) (specificTemplate *stack.TemplateSource) {
	for _, val := range templateInfo {
		if val.Name == customName {
			return &val
		}
	}
	return nil
}

func pullSpecificTemplate(templateInfo []stack.TemplateSource, customName string, cmd *cobra.Command) error {
	desiredTemplate := findTemplate(templateInfo, customName)
	if desiredTemplate == nil {
		return fmt.Errorf("Unable to find template repo with name: `%s`", customName)
	}
	if len(desiredTemplate.Source) == 0 {
		return runTemplateStorePull(cmd, []string{desiredTemplate.Name})
	}
	return pullTemplate(desiredTemplate.Source)
}
