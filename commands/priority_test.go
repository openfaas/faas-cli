package commands

import "testing"

func Test_getTemplateStoreURL(t *testing.T) {
	tests := []struct {
		title       string
		envURL      string
		defaultURL  string
		argURL      string
		expectedURL string
	}{
		{
			title:       "Environmental variable is set and argument equals defaultURL which should be priority",
			envURL:      "https://github.com/custom/url",
			defaultURL:  DefaultTemplatesStore,
			argURL:      DefaultTemplatesStore,
			expectedURL: "https://github.com/custom/url",
		},
		{
			title:       "Environmental variable is unset and argument is unset which falls back to default store",
			envURL:      "",
			defaultURL:  DefaultTemplatesStore,
			argURL:      DefaultTemplatesStore,
			expectedURL: DefaultTemplatesStore,
		},
		{
			title:       "Environmental variable is unset but argument is set which should set URL as argument",
			envURL:      "",
			defaultURL:  DefaultTemplatesStore,
			argURL:      "https://github.com/openfaas/store/official",
			expectedURL: "https://github.com/openfaas/store/official",
		},
		{
			title:       "Environmental variable is set and argument is set which should set URL as argument",
			envURL:      "https://github.com/custom/url",
			defaultURL:  DefaultTemplatesStore,
			argURL:      "https://github.com/openfaas/store/official",
			expectedURL: "https://github.com/openfaas/store/official",
		},
	}
	// defaultURL is always present that is why we don't test that case
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			storeURL := getTemplateStoreURL(test.argURL, test.envURL, test.defaultURL)
			if storeURL != test.expectedURL {
				t.Errorf("expected store URL: `%s` got: `%s`", test.expectedURL, storeURL)
			}
		})
	}
}

func Test_getOverrideNamespace(t *testing.T) {
	tests := []struct {
		stack    string
		flag     string
		want     string
		scenario string
	}{
		// Test cases
		{
			stack:    "",
			flag:     "",
			want:     "",
			scenario: "no namespace value set in flag and in namespace field of stack file",
		},

		{
			stack:    "openfaas-fn",
			flag:     "foo",
			want:     "foo",
			scenario: "both stack file and CLI flag provide namespace values",
		},

		{
			stack:    "bar",
			flag:     "",
			want:     "bar",
			scenario: "stack file provides namespace value whereas no namespace is provided by CLI",
		},

		{
			stack:    "",
			flag:     "foo",
			want:     "foo",
			scenario: "flag provides namespace value whereas no namespace is provided by stack file",
		},
	}

	// Run the test for each test case defined in "tests"
	for _, testCase := range tests {
		testCase := testCase
		functionNamespace := getNamespace(testCase.flag, testCase.stack)

		t.Run(testCase.scenario, func(t *testing.T) {
			if functionNamespace != testCase.want {
				t.Fatalf("Namespace incorrect want: %q but got: %q\n", testCase.want, functionNamespace)
			}
		})
	}
}
