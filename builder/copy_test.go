package builder

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func Test_CopyFiles(t *testing.T) {
	fileModes := []int{0600, 0640, 0644, 0700, 0755}

	dir := os.TempDir()
	for _, mode := range fileModes {
		// set up a source folder with 2 file
		srcDir, srcDirErr := setupSourceFolder(2, mode)
		if srcDirErr != nil {
			log.Fatal("Error creating source folder")
		}
		defer os.RemoveAll(srcDir)

		// create a destination folder to copy the files to
		destDir, destDirErr := ioutil.TempDir(dir, "openfaas-test-destination-")
		if destDirErr != nil {
			t.Fatalf("Error creating destination folder\n%v", destDirErr)
		}
		defer os.RemoveAll(destDir)

		CopyFiles(srcDir, destDir+"/")
		err := checkDestinationFiles(destDir, 2, mode)
		if err != nil {
			t.Fatalf("Destination file mode differs from source file mode\n%v", err)
		}
	}
}

func setupSourceFolder(numberOfFiles, mode int) (string, error) {
	dir := os.TempDir()
	data := []byte("open faas")

	// create a folder for source files
	srcDir, dirError := ioutil.TempDir(dir, "openfaas-test-source-")
	if dirError != nil {
		return "", dirError
	}

	// create n files inside the created folder
	for i := 1; i <= numberOfFiles; i++ {
		srcFile := fmt.Sprintf("%s/test-file-%d", srcDir, i)
		fileErr := ioutil.WriteFile(srcFile, data, os.FileMode(mode))
		if fileErr != nil {
			return "", fileErr
		}
	}

	return srcDir, nil
}

func checkDestinationFiles(dir string, numberOfFiles, mode int) error {
	// Check each file inside the destination folder
	for i := 1; i <= numberOfFiles; i++ {
		fileStat, err := os.Stat(fmt.Sprintf("%s/test-file-%d", dir, i))
		if os.IsNotExist(err) {
			return err
		}
		if fileStat.Mode() != os.FileMode(mode) {
			return errors.New("expected mode did not match")
		}
	}

	return nil
}
