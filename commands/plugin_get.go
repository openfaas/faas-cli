package commands

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/alexellis/arkade/pkg/archive"
	"github.com/alexellis/arkade/pkg/env"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/spf13/cobra"
)

var (
	pluginRegistry string
	clientOS       string
	clientArch     string
	tag            string
	pluginPath     string
)

func init() {
	pluginGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a plugin",
		Long: `Download and extract a plugin for faas-cli from a container
registry`,
		Example: `# Download a plugin by name:
faas-cli plugin get NAME

# Give a version
faas-cli plugin get NAME --version 0.0.1

# Give an explicit OS and architecture
faas-cli plugin get NAME --arch armhf --os linux

# Use a custom registry
faas-cli plugin get NAME --registry ghcr.io/openfaasltd`,
		RunE: runPluginGetCmd,
	}

	pluginGetCmd.Flags().StringVar(&pluginRegistry, "registry", "ghcr.io/openfaasltd", "The registry to pull the plugin from")
	pluginGetCmd.Flags().StringVar(&clientArch, "arch", "", "The architecture to pull the plugin for, give a value or leave blank for auto-detection")
	pluginGetCmd.Flags().StringVar(&clientOS, "os", "", "The OS to pull the plugin for, give a value or leave blank for auto-detection")
	pluginGetCmd.Flags().StringVar(&tag, "version", "latest", "Version or SHA for plugin")
	pluginGetCmd.Flags().StringVar(&pluginPath, "path", "$HOME/.openfaas/plugins", "The path for the plugin")

	pluginGetCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	pluginCmd.AddCommand(pluginGetCmd)
}

// preRunPublish validates args & flags
func runPluginGetCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the name of the plugin")
	}

	if len(tag) == 0 {
		return fmt.Errorf("please provide the version of the plugin or \"latest\"")
	}

	arch, operatingSystem := getClientArch()

	if len(clientArch) == 0 {
		clientArch = arch
	}

	if len(clientOS) == 0 {
		clientOS = operatingSystem
	}

	st := time.Now()
	pluginName := args[0]
	src := fmt.Sprintf("%s/%s:%s", pluginRegistry, pluginName, tag)

	if verbose {
		fmt.Printf("Fetching plugin: %s %s for: %s/%s\n", pluginName, src, clientOS, clientArch)
	} else {
		fmt.Printf("Fetching plugin: %s\n", pluginName)
	}

	var pluginDir string

	if cmd.Flags().Changed("path") {
		pluginPath = strings.ReplaceAll(pluginPath, "$HOME", os.Getenv("HOME"))

		pluginDir = pluginPath
	} else {
		if runtime.GOOS == "windows" {
			pluginDir = os.Expand("$HOMEPATH/.openfaas/plugins", os.Getenv)
		} else {
			pluginDir = os.ExpandEnv("$HOME/.openfaas/plugins")
		}
	}

	if _, err := os.Stat(pluginDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(pluginDir, 0755); err != nil && os.ErrExist != err {
			return fmt.Errorf("failed to create plugin directory %s: %w", pluginDir, err)
		}
	}

	tmpTar := path.Join(os.TempDir(), pluginName+".tar")

	f, err := os.Create(tmpTar)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tmpTar, err)
	}
	defer f.Close()

	var img v1.Image

	downloadArch, downloadOS := getDownloadArch(clientArch, clientOS)

	img, err = crane.Pull(src, crane.WithPlatform(&v1.Platform{Architecture: downloadArch, OS: downloadOS}))
	if err != nil {
		return fmt.Errorf("pulling %s: %w", src, err)
	}

	if err := crane.Export(img, f); err != nil {
		return fmt.Errorf("exporting %s: %w", src, err)
	}

	if verbose {
		fmt.Printf("Wrote OCI filesystem to: %s\n", tmpTar)
	}
	tarFile, err := os.Open(tmpTar)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tmpTar, err)
	}

	defer tarFile.Close()

	if verbose {
		fmt.Printf("Writing %q\n", path.Join(pluginDir, pluginName))
	}

	defer os.Remove(tmpTar)
	gzipped := false
	if err := archive.Untar(tarFile, pluginDir, gzipped, true); err != nil {
		return fmt.Errorf("failed to untar %s: %w", tmpTar, err)
	}

	// Add the .exe filename extension to the plugin executable on windows.
	// If the .exe extension is missing the plugin will not execute.
	if runtime.GOOS == "windows" {
		pluginPath := path.Join(pluginDir, pluginName)
		err := os.Rename(pluginPath, fmt.Sprintf("%s.exe", pluginPath))
		if err != nil {
			return fmt.Errorf("failed to move plugin %w", err)
		}
	}

	if cmd.Flags().Changed("path") {
		fmt.Printf("Wrote: %s (%s/%s) in (%s)\n", path.Join(pluginPath, pluginName), clientOS, clientArch, time.Since(st).Round(time.Millisecond))
	} else {
		fmt.Printf("Downloaded in (%s)\n\nUsage:\n  faas-cli %s\n", time.Since(st).Round(time.Millisecond), pluginName)
	}
	return nil
}

func getClientArch() (arch string, os string) {
	if runtime.GOOS == "windows" {
		return runtime.GOARCH, runtime.GOOS
	}

	return env.GetClientArch()
}

func getDownloadArch(clientArch, clientOS string) (arch string, os string) {
	downloadArch := strings.ToLower(clientArch)
	downloadOS := strings.ToLower(clientOS)

	if downloadArch == "x86_64" {
		downloadArch = "amd64"
	} else if downloadArch == "aarch64" {
		downloadArch = "arm64"
	}

	return downloadArch, downloadOS
}
