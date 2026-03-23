// Copyright (c) OpenFaaS Author(s) 2024. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

var (
	chartExportOutput    string
	chartExportValues    []string
	chartExportSet       []string
	chartExportCRDs      bool
	chartExportNamespace string
	chartExportRelease   string
)

var chartExportCmd = &cobra.Command{
	Use:   `export [CHART_NAME] [flags]`,
	Short: "Render a Helm chart and export each resource as a separate YAML file",
	Long: `Renders a Helm chart using "helm template" and splits the output into
individual YAML files, organised into folders by resource kind.

CustomResourceDefinitions are prefixed with 00_ so that they sort first
when applied with "kubectl apply -f".

CHART_NAME is optional and defaults to "openfaas". It maps to
"chart/<CHART_NAME>" relative to the current directory.`,
	Example: `  # Export the openfaas chart with default values
  faas-cli chart export

  # Export with pro values
  faas-cli chart export --values chart/openfaas/values-pro.yaml

  # Export kafka-connector chart to a custom directory
  faas-cli chart export kafka-connector -o ./rendered

  # Export without CRDs
  faas-cli chart export --crds=false

  # Export with value overrides
  faas-cli chart export --values chart/openfaas/values-pro.yaml --set openfaasPro=true`,
	RunE:    runChartExport,
	PreRunE: preRunChartExport,
}

func init() {
	chartExportCmd.Flags().StringVarP(&chartExportOutput, "output", "o", "./yaml", "Output directory for rendered YAML files")
	chartExportCmd.Flags().StringArrayVar(&chartExportValues, "values", nil, "Path to values file(s) to use during rendering")
	chartExportCmd.Flags().StringArrayVar(&chartExportSet, "set", nil, "Set individual values (key=value)")
	chartExportCmd.Flags().BoolVar(&chartExportCRDs, "crds", true, "Include CRDs in the output")
	chartExportCmd.Flags().StringVarP(&chartExportNamespace, "namespace", "n", "", "Kubernetes namespace for rendered manifests")
	chartExportCmd.Flags().StringVar(&chartExportRelease, "release", "openfaas", "Helm release name")

	chartCmd.AddCommand(chartExportCmd)
}

func preRunChartExport(cmd *cobra.Command, args []string) error {
	if _, err := exec.LookPath("helm"); err != nil {
		return fmt.Errorf("helm is required but was not found in PATH")
	}

	chartPath := resolveChartPath(args)
	info, err := os.Stat(chartPath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("chart directory not found: %s", chartPath)
	}

	return nil
}

func runChartExport(cmd *cobra.Command, args []string) error {
	chartPath := resolveChartPath(args)

	helmArgs := []string{"template", chartExportRelease, chartPath}

	if chartExportCRDs {
		helmArgs = append(helmArgs, "--include-crds")
	}

	for _, vf := range chartExportValues {
		helmArgs = append(helmArgs, "-f", vf)
	}

	for _, s := range chartExportSet {
		helmArgs = append(helmArgs, "--set", s)
	}

	if chartExportNamespace != "" {
		helmArgs = append(helmArgs, "--namespace", chartExportNamespace)
	}

	fmt.Printf("Running: helm %s\n", strings.Join(helmArgs, " "))

	helmCmd := exec.Command("helm", helmArgs...)
	var stdout, stderr bytes.Buffer
	helmCmd.Stdout = &stdout
	helmCmd.Stderr = &stderr

	if err := helmCmd.Run(); err != nil {
		return fmt.Errorf("helm template failed: %s\n%s", err, stderr.String())
	}

	resources, err := splitYAMLStream(&stdout)
	if err != nil {
		return fmt.Errorf("failed to parse YAML output: %s", err)
	}

	if len(resources) == 0 {
		return fmt.Errorf("no resources found in helm output")
	}

	outputDir, err := filepath.Abs(chartExportOutput)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("failed to clean output directory: %s", err)
	}

	// Detect duplicate kind+name pairs so we can disambiguate with namespace
	type kindName struct{ kind, name string }
	seen := make(map[kindName]int)
	for _, res := range resources {
		seen[kindName{res.Kind, res.Name}]++
	}

	written := 0
	for _, res := range resources {
		dir := strings.ToLower(res.Kind)
		if res.Kind == "CustomResourceDefinition" {
			dir = "00_" + dir
		}

		destDir := filepath.Join(outputDir, dir)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %s", destDir, err)
		}

		filename := res.Name
		if seen[kindName{res.Kind, res.Name}] > 1 && res.Namespace != "" {
			filename = res.Name + "." + res.Namespace
		}

		destFile := filepath.Join(destDir, filename+".yaml")
		if err := os.WriteFile(destFile, res.Raw, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %s", destFile, err)
		}

		rel, _ := filepath.Rel(outputDir, destFile)
		fmt.Printf("  wrote %s\n", rel)
		written++
	}

	fmt.Printf("\nExported %d resources to %s\n", written, outputDir)
	return nil
}

type chartResource struct {
	Kind      string
	Name      string
	Namespace string
	Raw       []byte
}

func splitYAMLStream(r io.Reader) ([]chartResource, error) {
	decoder := yaml.NewDecoder(r)
	var resources []chartResource

	for {
		var doc map[string]interface{}
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if doc == nil {
			continue
		}

		kind, _ := doc["kind"].(string)
		if kind == "" {
			continue
		}

		meta, _ := doc["metadata"].(map[string]interface{})
		if meta == nil {
			continue
		}
		name, _ := meta["name"].(string)
		if name == "" {
			continue
		}
		namespace, _ := meta["namespace"].(string)

		raw, err := marshalYAML(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s/%s: %s", kind, name, err)
		}

		resources = append(resources, chartResource{
			Kind:      kind,
			Name:      name,
			Namespace: namespace,
			Raw:       raw,
		})
	}

	return resources, nil
}

func marshalYAML(doc map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	enc.Close()
	return buf.Bytes(), nil
}

func resolveChartPath(args []string) string {
	chartName := "openfaas"
	if len(args) > 0 && args[0] != "" {
		chartName = args[0]
	}
	return filepath.Join("chart", chartName)
}
