// Copyright (c) Alex Ellis, Eric Stoekl 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var SmallestZipFile = []byte{80, 75, 05, 06, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00}

func Test_PullTemplates(t *testing.T) {
	// Remove existing master.zip file if it exists
	if _, err := os.Stat("master.zip"); err == nil {
		t.Log("Found a master.zip file, removing it...")

		err := os.Remove("master.zip")
		if err != nil {
			t.Fatal(err)
		}
	}

	// Remove existing templates folder, if it exist
	if _, err := os.Stat("template/"); err == nil {
		t.Log("Found a template/ directory, removing it...")

		err := os.RemoveAll("template/")
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create fake server for testing.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Write out the minimum number of bytes to make the response a valid .zip file
		w.Write(SmallestZipFile)

	}))
	defer ts.Close()

	err := PullTemplates(ts.URL)
	if err != nil {
		t.Error(err)
	}
}
