package commands

import (
	"testing"
)

func Test_FilterTemplate(t *testing.T) {
	// Template objects containing
	// architectures as the template
	// store would provide
	templates := []TemplateInfo{
		{
			Platform: "arm64",
		},
		{
			Platform: "armhf",
		},
		{
			Platform: "x86_64",
		},
	}
	// Valid values a user can give to the --platform flag
	validPlatforms := []string{"ARMHF", "armhf", "ARM64", "arm64", "X86_64", "x86_64"}
	for _, validPlatform := range validPlatforms {
		filteredTemplate := filterTemplate(templates, validPlatform)
		if len(filteredTemplate) != 1 {
			t.Errorf("Expected one object to be filtered got: %d", len(filteredTemplate))
		}
	}
}
