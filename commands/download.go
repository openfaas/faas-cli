package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:     `download`,
	Short:   "Download binaries",
	Long:    "Download binaries",
	Example: "faas-cli download",
	Run:     runFunc,
}

func runFunc(cmd *cobra.Command, args []string) {
	cmd.Help()
	if len(args) == 0 {
		fmt.Printf("You can download: %s\n", strings.TrimRight(strings.Join(getDownloads(), ", "), ", "))
	}
}

func getDownloads() []string {
	return []string{"kubeseal"}
}

func downloadBinary(url string) error {
	var netHTTPClient = &http.Client{
		Timeout: 5 * time.Minute,
	}

	fmt.Println("Downloading kubeseal")
	resp, err := netHTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("Could not download kubeseal: \n%v", err)
	}
	defer resp.Body.Close()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Could not get current user: \n%v", err)
	}
	kubesealDownloadDir := path.Join(homeDir, ".openfaas", "bin")
	err = os.MkdirAll(kubesealDownloadDir, 0700)
	if err != nil {
		return fmt.Errorf("Could not create dir for download: %v", err)
	}

	kubesealPath := fmt.Sprintf("%s/%s", kubesealDownloadDir, "kubeseal")
	out, err := os.Create(kubesealPath)
	if err != nil {
		return fmt.Errorf("Could not create file: \n%v", err)
	}
	err = os.Chmod(kubesealPath, 0700)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Could not save the downloaded binary: %v", err)
	}
	return nil
}
