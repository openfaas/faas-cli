package commands

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alexellis/arkade/pkg/env"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/spf13/cobra"
)

var pluginRegistry string
var clientOS string
var clientArch string
var tag string

func init() {
	pluginGetCmd := &cobra.Command{
		Use:     "get",
		Short:   "Get a plugin",
		Long:    `Get a plugin`,
		Example: `faas-cli plugin get NAME`,
		RunE:    runPluginGetCmd,
	}

	pluginGetCmd.Flags().StringVar(&pluginRegistry, "registry", "ghcr.io/openfaasltd", "The registry to pull the plugin from")
	pluginGetCmd.Flags().StringVar(&clientArch, "arch", "", "The architecture to pull the plugin for, give a value or leave blank for auto-detection")
	pluginGetCmd.Flags().StringVar(&clientOS, "os", "", "The OS to pull the plugin for, give a value or leave blank for auto-detection")
	pluginGetCmd.Flags().StringVar(&tag, "version", "latest", "Version or SHA for plugin")
	pluginGetCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")

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

	pluginDir := os.ExpandEnv("$HOME/.openfaas/plugins")

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		os.MkdirAll(pluginDir, 0755)
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
	if err := untar(tarFile, pluginDir, true); err != nil {
		return fmt.Errorf("failed to untar %s: %w", tmpTar, err)
	}
	fmt.Printf("OK.. took: (%ds)\n", int(time.Since(st).Seconds()))
	return nil
}

// TODO: switch to arkade package, but update it first so that it can take a gzip option true/false
func untar(r io.Reader, dir string, quiet bool) (err error) {
	t0 := time.Now()
	nFiles := 0
	madeDir := map[string]bool{}
	defer func() {
		td := time.Since(t0)

		if err == nil {
			if !quiet {
				log.Printf("extracted tarball into %s: %d files, %d dirs (%v)", dir, nFiles, len(madeDir), td)
			}
		} else {
			log.Printf("error extracting tarball into %s after %d files, %d dirs, %v: %v", dir, nFiles, len(madeDir), td, err)
		}

	}()

	tr := tar.NewReader(r)
	loggedChtimesError := false
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("tar reading error: %v", err)
			return fmt.Errorf("tar error: %v", err)
		}
		if !validRelPath(f.Name) {
			return fmt.Errorf("tar contained invalid name error %q", f.Name)
		}
		baseFile := filepath.Base(f.Name)
		abs := path.Join(dir, baseFile)
		if !quiet {
			fmt.Printf("Extracting: %s to\t%s\n", f.Name, abs)
		}

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsDir():

			break

		case mode.IsRegular():

			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			modTime := f.ModTime
			if modTime.After(t0) {
				// Clamp modtimes at system time. See
				// golang.org/issue/19062 when clock on
				// buildlet was behind the gitmirror server
				// doing the git-archive.
				modTime = t0
			}
			if !modTime.IsZero() {
				if err := os.Chtimes(abs, modTime, modTime); err != nil && !loggedChtimesError {
					// benign error. Gerrit doesn't even set the
					// modtime in these, and we don't end up relying
					// on it anywhere (the gomote push command relies
					// on digests only), so this is a little pointless
					// for now.
					log.Printf("error changing modtime: %v (further Chtimes errors suppressed)", err)
					loggedChtimesError = true // once is enough
				}
			}
			nFiles++
		default:
		}
	}
	return nil
}

func validRelativeDir(dir string) bool {
	if strings.Contains(dir, `\`) || path.IsAbs(dir) {
		return false
	}
	dir = path.Clean(dir)
	if strings.HasPrefix(dir, "../") || strings.HasSuffix(dir, "/..") || dir == ".." {
		return false
	}
	return true
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
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
