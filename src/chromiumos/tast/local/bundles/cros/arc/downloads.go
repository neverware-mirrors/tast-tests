// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/faillog"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         Downloads,
		Desc:         "Checks Downloads integration is working",
		Attr:         []string{"informational"},
		SoftwareDeps: []string{"android", "chrome_login"},
		Data:         []string{"capybara.jpg"},
		Timeout:      3 * time.Minute,
	})
}

func Downloads(s *testing.State) {
	const (
		filename    = "capybara.jpg"
		crosPath    = "/home/chronos/user/Downloads/" + filename
		androidPath = "/storage/emulated/0/Download/" + filename
	)

	defer faillog.SaveIfError(s)

	cr, err := chrome.New(s.Context())
	if err != nil {
		s.Fatal("Failed to connect to Chrome: ", err)
	}
	defer cr.Close(s.Context())

	a, err := arc.New(s.Context(), cr, s.OutDir())
	if err != nil {
		s.Fatal("Failed to start ARC: ", err)
	}
	defer a.Close()

	expected, err := ioutil.ReadFile(s.DataPath(filename))
	if err != nil {
		s.Fatal("Could not read the test file: ", err)
	}

	// CrOS -> Android
	if err = ioutil.WriteFile(crosPath, expected, 0666); err != nil {
		s.Fatalf("Could not write to %s: %v", crosPath, err)
	}
	actual, err := a.ReadFile(s.Context(), androidPath)
	if err != nil {
		s.Error("CrOS -> Android failed: ", err)
	} else if !bytes.Equal(actual, expected) {
		s.Error("CrOS -> Android failed: content mismatch")
	}
	if err = os.Remove(crosPath); err != nil {
		s.Fatal("Failed to remove a file: ", err)
	}

	// Android -> CrOS
	if err = a.WriteFile(s.Context(), androidPath, expected); err != nil {
		s.Fatalf("Could not write to %s: %v", androidPath, err)
	}
	actual, err = ioutil.ReadFile(crosPath)
	if err != nil {
		s.Error("Android -> CrOS failed: ", err)
	} else if !bytes.Equal(actual, expected) {
		s.Error("Android -> CrOS failed: content mismatch")
	}
	if err = os.Remove(crosPath); err != nil {
		s.Fatal("Failed to remove a file: ", err)
	}
}
