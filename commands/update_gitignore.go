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

func updateContent(content string) (updated_content string) {
	// append files to ignore to file content if it is not already ignored

	filesToIgnore := []string{"template", "build"}

	lines := strings.Split(content, "\n")

	for _, file := range filesToIgnore {
		if !contains(lines, file) {
			lines = append(lines, file)
		}
	}

	updated_content = strings.Join(lines, "\n")
	updated_content = strings.Trim(updated_content, "\n")
	return updated_content
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

	string_content := string(content[:])
	write_content := updateContent(string_content)

	_, err = f.WriteString(write_content + "\n")
	if err != nil {
		return err
	}

	return nil
}
