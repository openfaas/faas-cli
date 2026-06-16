package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	types "github.com/openfaas/faas-provider/types"
	"github.com/openfaas/go-sdk/stack"
	"github.com/spf13/cobra"
)

type funcDiff struct {
	Image                  string
	FProcess               string
	Env                    map[string]string
	Secrets                []string
	Constraints            []string
	Labels                 map[string]string
	Annotations            map[string]string
	Limits                 *functionResourcesDiff
	Requests               *functionResourcesDiff
	ReadOnlyRootFilesystem bool
}

type functionResourcesDiff struct {
	Memory string
	CPU    string
}

const (
	diffEnvLocal  = "local"
	diffEnvRemote = "remote"
	diffEnvAll    = "all"
)

var diffEnvMode = diffEnvLocal

func init() {
	diffCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	diffCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	diffCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yaml file")
	diffCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
	diffCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", "", "Namespace override for the diff")
	diffCmd.Flags().Var(&tagFormat, "tag", "Override latest tag on function Docker image, accepts 'digest', 'sha', 'branch', or 'describe', or 'latest'")
	diffCmd.Flags().StringVar(&diffEnvMode, "env", diffEnvLocal, "Environment diff mode: local, remote, or all")

	faasCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   `diff [--gateway GATEWAY_URL]`,
	Short: "Diff stack.yaml against deployed functions",
	Long: `Compares the function definitions in stack.yaml with what is actually
deployed on the gateway and shows the differences.

This command is read-only - it only makes a GET request to list functions.`,
	Example: `  faas-cli diff
  faas-cli diff -f my-stack.yaml
  faas-cli diff --gateway https://my-gateway.example.com`,
	RunE: runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {
	if len(yamlFile) == 0 {
		return fmt.Errorf("no YAML file specified - use -f to provide a stack.yaml")
	}
	if !validDiffEnvMode(diffEnvMode) {
		return fmt.Errorf("unknown env diff mode: %s", diffEnvMode)
	}

	parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
	if err != nil {
		return err
	}

	if parsedServices == nil || len(parsedServices.Functions) == 0 {
		return fmt.Errorf("no functions found in %s", yamlFile)
	}

	gatewayAddress := getGatewayURL(gateway, defaultGateway, parsedServices.Provider.GatewayURL, os.Getenv(openFaaSURLEnvironment))
	cliAuth, err := proxy.NewCLIAuth(token, gatewayAddress)
	if err != nil {
		return err
	}
	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)
	proxyClient, err := proxy.NewClient(cliAuth, gatewayAddress, transport, &commandTimeout)
	if err != nil {
		return err
	}

	functions, err := proxyClient.ListFunctions(context.Background(), functionNamespace)
	if err != nil {
		return err
	}

	yamlFns := parsedServices.Functions

	yamlMap := make(map[string]funcDiff)
	for name, fn := range yamlFns {
		key := diffKey(name, namespaceForDiffKey(functionNamespace, fn.Namespace))

		imageName, err := buildDiffImageName(fn.Image, fn.Handler, tagFormat)
		if err != nil {
			return err
		}

		env := fn.Environment
		if env == nil {
			env = make(map[string]string)
		}
		yamlMap[key] = funcDiff{
			Image:                  imageName,
			FProcess:               fn.FProcess,
			Env:                    env,
			Secrets:                fn.Secrets,
			Constraints:            constraintsFromPtr(fn.Constraints),
			Labels:                 mapFromPtr(fn.Labels),
			Annotations:            mapFromPtr(fn.Annotations),
			Limits:                 resourcesFromStack(fn.Limits),
			Requests:               resourcesFromStack(fn.Requests),
			ReadOnlyRootFilesystem: fn.ReadOnlyRootFilesystem,
		}
	}

	deployedMap := make(map[string]funcDiff)
	for _, fn := range functions {
		envMap := fn.EnvVars
		if envMap == nil {
			envMap = make(map[string]string)
		}

		fnDiff := funcDiff{
			Image:                  fn.Image,
			FProcess:               fn.EnvProcess,
			Env:                    envMap,
			Secrets:                fn.Secrets,
			Constraints:            fn.Constraints,
			Labels:                 mapFromPtr(fn.Labels),
			Annotations:            mapFromPtr(fn.Annotations),
			Limits:                 resourcesFromStatus(fn.Limits),
			Requests:               resourcesFromStatus(fn.Requests),
			ReadOnlyRootFilesystem: fn.ReadOnlyRootFilesystem,
		}
		deployedMap[diffKey(fn.Name, "")] = fnDiff
		if fn.Namespace != "" {
			deployedMap[diffKey(fn.Name, fn.Namespace)] = fnDiff
		}
	}

	keys := funcDiffKeys(yamlMap)
	sort.Strings(keys)

	hasDiff := false
	for _, key := range keys {
		yamlF, yamlExists := yamlMap[key]
		deployedF, deployedExists := deployedMap[key]

		rows, changed := buildSideBySide(yamlF, yamlExists, deployedF, deployedExists, diffEnvMode)
		if !changed {
			continue
		}
		hasDiff = true
		printDifftool(key, rows)
	}

	if !hasDiff {
		fmt.Println("YAML matches deployment, no differences found.")
		return nil
	}

	return fmt.Errorf("differences found")
}

func buildDiffImageName(image string, handler string, tagMode schema.BuildFormat) (string, error) {
	branch, version, err := builder.GetImageTagValues(tagMode, handler)
	if err != nil {
		return "", err
	}

	return schema.BuildImageName(tagMode, image, version, branch), nil
}

type diffRow struct {
	left  diffCell
	right diffCell
}

type diffCell struct {
	marker string
	field  string
	value  string
}

func buildSideBySide(yamlF funcDiff, yamlExists bool, deployedF funcDiff, deployedExists bool, envMode string) ([]diffRow, bool) {
	var rows []diffRow

	if yamlExists && !deployedExists {
		rows = append(rows, diffRow{
			left:  diffCell{marker: "-", field: "function", value: "defined in stack.yaml"},
			right: diffCell{marker: "+", field: "function", value: "not deployed"},
		})
		return rows, true
	}

	yamlAttrs := sortedAttrMap(yamlF)
	depAttrs := sortedAttrMap(deployedF)
	fields := allAttrKeys(yamlAttrs, depAttrs)

	for _, field := range fields {
		leftValue, leftHas := yamlAttrs[field]
		rightValue, rightHas := depAttrs[field]

		if !leftHas && !rightHas || (leftHas && rightHas && leftValue == rightValue) {
			continue
		}
		if ignoreAttr(field, leftHas, rightHas, envMode) {
			continue
		}

		row := diffRow{}
		if leftHas {
			row.left = diffCell{marker: "-", field: field, value: leftValue}
		}
		if rightHas {
			row.right = diffCell{marker: "+", field: field, value: rightValue}
		}

		rows = append(rows, row)
	}

	return rows, len(rows) > 0
}

func ignoreAttr(field string, leftHas bool, rightHas bool, envMode string) bool {
	if field == "fprocess" {
		return !leftHas && rightHas
	}

	if !strings.HasPrefix(field, "env.") {
		return false
	}

	switch envMode {
	case diffEnvLocal:
		return !leftHas && rightHas
	case diffEnvRemote:
		return leftHas && !rightHas
	default:
		return false
	}
}

func validDiffEnvMode(envMode string) bool {
	return envMode == diffEnvLocal || envMode == diffEnvRemote || envMode == diffEnvAll
}

func sortedAttrMap(f funcDiff) map[string]string {
	m := make(map[string]string)
	if f.Image != "" {
		m["image"] = f.Image
	}
	if f.FProcess != "" {
		m["fprocess"] = f.FProcess
	}
	m["readonly_root_filesystem"] = fmt.Sprintf("%t", f.ReadOnlyRootFilesystem)
	for k, v := range f.Env {
		m["env."+k] = v
	}
	for _, s := range f.Secrets {
		m["secret."+s] = "present"
	}
	for _, constraint := range f.Constraints {
		m["constraint."+constraint] = "present"
	}
	for k, v := range f.Labels {
		m["label."+k] = v
	}
	for k, v := range f.Annotations {
		m["annotation."+k] = v
	}
	addResources(m, "limits", f.Limits)
	addResources(m, "requests", f.Requests)
	return m
}

func addResources(m map[string]string, prefix string, resources *functionResourcesDiff) {
	if resources == nil {
		return
	}

	if resources.CPU != "" {
		m[prefix+".cpu"] = resources.CPU
	}
	if resources.Memory != "" {
		m[prefix+".memory"] = resources.Memory
	}
}

func mapFromPtr(m *map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return *m
}

func constraintsFromPtr(constraints *[]string) []string {
	if constraints == nil {
		return []string{}
	}
	return *constraints
}

func resourcesFromStack(resources *stack.FunctionResources) *functionResourcesDiff {
	if resources == nil {
		return nil
	}

	return &functionResourcesDiff{
		Memory: resources.Memory,
		CPU:    resources.CPU,
	}
}

func resourcesFromStatus(resources *types.FunctionResources) *functionResourcesDiff {
	if resources == nil {
		return nil
	}

	return &functionResourcesDiff{
		Memory: resources.Memory,
		CPU:    resources.CPU,
	}
}

func allAttrKeys(maps ...map[string]string) []string {
	set := make(map[string]bool)
	for _, m := range maps {
		for k := range m {
			set[k] = true
		}
	}
	keys := sortedKeys(set)
	return keys
}

func printDifftool(key string, rows []diffRow) {
	leftWidth, rightWidth, fieldWidth := diffColumnWidths(rows)

	fmt.Printf("%s\n", key)
	fmt.Fprintf(os.Stdout, "  %s | %s\n", pad("stack.yaml", leftWidth), pad("deployed", rightWidth))
	fmt.Fprintf(os.Stdout, "  %s-+-%s\n", strings.Repeat("-", leftWidth), strings.Repeat("-", rightWidth))

	for _, row := range rows {
		leftText := formatDiffCell(row.left, fieldWidth)
		rightText := formatDiffCell(row.right, fieldWidth)

		fmt.Fprintf(os.Stdout, "  %s | %s\n", pad(leftText, leftWidth), pad(rightText, rightWidth))
	}
	fmt.Println()
}

func diffColumnWidths(rows []diffRow) (int, int, int) {
	fieldWidth := len("image:")
	for _, row := range rows {
		if row.left.field != "" && len(row.left.field)+1 > fieldWidth {
			fieldWidth = len(row.left.field) + 1
		}
		if row.right.field != "" && len(row.right.field)+1 > fieldWidth {
			fieldWidth = len(row.right.field) + 1
		}
	}

	leftWidth := len("stack.yaml")
	rightWidth := len("deployed")
	for _, row := range rows {
		leftText := formatDiffCell(row.left, fieldWidth)
		rightText := formatDiffCell(row.right, fieldWidth)

		if len(leftText) > leftWidth {
			leftWidth = len(leftText)
		}
		if len(rightText) > rightWidth {
			rightWidth = len(rightText)
		}
	}

	if leftWidth < 40 {
		leftWidth = 40
	}
	if rightWidth < 40 {
		rightWidth = 40
	}

	return leftWidth, rightWidth, fieldWidth
}

func formatDiffCell(cell diffCell, fieldWidth int) string {
	if cell.field == "" {
		return ""
	}

	return fmt.Sprintf("%s %-*s %s", cell.marker, fieldWidth, cell.field+":", cell.value)
}

func pad(s string, w int) string {
	out := s
	for i := len(out); i < w; i++ {
		out += " "
	}
	return out
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedAttrKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func namespaceForDiffKey(flagNamespace string, stackNamespace string) string {
	if flagNamespace != "" {
		return flagNamespace
	}
	return stackNamespace
}

func diffKey(name string, namespace string) string {
	if namespace != "" && !strings.Contains(name, ".") {
		return name + "." + namespace
	}
	return name
}

func funcDiffKeys(functions map[string]funcDiff) []string {
	keys := make([]string, 0, len(functions))
	for k := range functions {
		keys = append(keys, k)
	}
	return keys
}
