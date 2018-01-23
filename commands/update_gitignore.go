package commands

import (
	"io/ioutil"
	"os"
	"strings"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func updateContent(content string) (updatedContent string) {
	// append files to ignore to file content if it is not already ignored

	filesToIgnore := []string{"template", "build"}

	lines := strings.Split(content, "\n")

	for _, file := range filesToIgnore {
		if !contains(lines, file) {
			lines = append(lines, file)
		}
	}

	updatedContent = strings.Join(lines, "\n")
	updatedContent = strings.Trim(updatedContent, "\n")
	return updatedContent
}

func updateGitignore() (err error) {
	// update .gitignore file if it already present othewise creates it

	f, err := os.OpenFile(".gitignore", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	content, err := ioutil.ReadFile(".gitignore")
	if err != nil {
		return err
	}

	stringContent := string(content[:])
	writeContent := updateContent(stringContent)

	_, err = f.WriteString(writeContent + "\n")
	if err != nil {
		return err
	}

	return nil
}
