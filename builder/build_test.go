package builder

import (
	"testing"
)

func Test_buildFlagSlice(t *testing.T) {

	var buildFlagOpts = []struct {
		title         string
		nocache       bool
		squash        bool
		httpProxy     string
		httpsProxy    string
		buildArgMap   map[string]string
		expectedSlice []string
	}{
		{
			title:         "no cache only",
			nocache:       true,
			squash:        false,
			httpProxy:     "",
			httpsProxy:    "",
			buildArgMap:   make(map[string]string),
			expectedSlice: []string{"--no-cache"},
		},
		{
			title:         "no cache & squash only",
			nocache:       true,
			squash:        true,
			httpProxy:     "",
			httpsProxy:    "",
			buildArgMap:   make(map[string]string),
			expectedSlice: []string{"--no-cache", "--squash"},
		},
		{
			title:         "no cache & squash & http proxy only",
			nocache:       true,
			squash:        true,
			httpProxy:     "192.168.0.1",
			httpsProxy:    "",
			buildArgMap:   make(map[string]string),
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "http_proxy=192.168.0.1"},
		},
		{
			title:         "no cache & squash & https-proxy only",
			nocache:       true,
			squash:        true,
			httpProxy:     "",
			httpsProxy:    "127.0.0.1",
			buildArgMap:   make(map[string]string),
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "https_proxy=127.0.0.1"},
		},
		{
			title:         "no cache & squash & http-proxy & https-proxy only",
			nocache:       true,
			squash:        true,
			httpProxy:     "192.168.0.1",
			httpsProxy:    "127.0.0.1",
			buildArgMap:   make(map[string]string),
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "http_proxy=192.168.0.1", "--build-arg", "https_proxy=127.0.0.1"},
		},
		{
			title:         "http-proxy & https-proxy only",
			nocache:       false,
			squash:        false,
			httpProxy:     "192.168.0.1",
			httpsProxy:    "127.0.0.1",
			buildArgMap:   make(map[string]string),
			expectedSlice: []string{"--build-arg", "http_proxy=192.168.0.1", "--build-arg", "https_proxy=127.0.0.1"},
		},
		{
			title:      "build arg map no spaces",
			nocache:    false,
			squash:     false,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppet": "ernie",
			},
			expectedSlice: []string{"--build-arg", "muppet=ernie"},
		},
		{
			title:      "build arg map with spaces",
			nocache:    false,
			squash:     false,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppets": "burt and ernie",
			},
			expectedSlice: []string{"--build-arg", "muppets=burt and ernie"},
		},
		{
			title:      "multiple build arg map with spaces",
			nocache:    false,
			squash:     false,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppets":    "burt and ernie",
				"playschool": "Jemima",
			},
			expectedSlice: []string{"--build-arg", "muppets=burt and ernie", "--build-arg", "playschool=Jemima"},
		},
		{
			title:      "no-cache and squash with multiple build arg map with spaces",
			nocache:    true,
			squash:     true,
			httpProxy:  "",
			httpsProxy: "",
			buildArgMap: map[string]string{
				"muppets":    "burt and ernie",
				"playschool": "Jemima",
			},
			expectedSlice: []string{"--no-cache", "--squash", "--build-arg", "muppets=burt and ernie", "--build-arg", "playschool=Jemima"},
		},
	}

	for _, test := range buildFlagOpts {

		t.Run(test.title, func(t *testing.T) {

			flagSlice := buildFlagSlice(test.nocache, test.squash, test.httpProxy, test.httpsProxy, test.buildArgMap)

			if len(flagSlice) != len(test.expectedSlice) {
				t.Errorf("Slices differ in size - wanted: %d, found %d", len(test.expectedSlice), len(flagSlice))
			}

			//map iteration is non-deterministic so dont compare values if the buildArgMap is > 1
			if len(test.buildArgMap) <= 1 {
				for i, val := range flagSlice {

					if val != test.expectedSlice[i] {
						t.Errorf("Slices differ in values - wanted: %s, found %s", test.expectedSlice[i], val)
					}

				}
			}

		})
	}

}
