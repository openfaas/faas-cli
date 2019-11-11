package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	kubesealLatestReleaseURL = "https://api.github.com/repos/bitnami-labs/sealed-secrets/releases/latest"
	msg                      = `# Add $HOME/.openfaas/bin/ to your $PATH variable

export PATH=$PATH:$HOME/.openfaas/bin`
)

func init() {
	downloadCmd.AddCommand(downloadKubeSeal)
}

var downloadKubeSeal = &cobra.Command{
	Use:     `kubeseal`,
	Short:   "Download kubeseal",
	Long:    "Download kubeseal",
	Example: "faas-cli download kubeseal",
	RunE:    runDownloadKubeSeal,
}

func runDownloadKubeSeal(cmd *cobra.Command, args []string) error {
	err := downloadKubesealBinary()
	return err
}

func downloadKubesealBinary() error {
	// Get the latest kubeseal release tag
	fmt.Println("Getting the latest release info")
	resp, err := http.Get(kubesealLatestReleaseURL)
	if err != nil {
		return fmt.Errorf("Error getting the latest kubeseal release info: \n%v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

	}
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return fmt.Errorf("Error unmarshalling response: \n%v", err)
	}

	latestRelease := dat["tag_name"]
	clientOS := runtime.GOOS
	clientArch := runtime.GOARCH
	kubesealURL := fmt.Sprintf("https://github.com/bitnami/sealed-secrets/releases/download/%s/kubeseal-%s-%s", latestRelease, clientOS, clientArch)

	// download kubeseal binary
	err = downloadBinary(kubesealURL)
	if err != nil {
		return err
	}
	fmt.Println("kubeseal downloaded")

	fmt.Println(msg)
	return nil
}
