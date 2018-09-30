// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"time"

	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         Downloads,
		Desc:         "Checks Downloads integration is working",
		Attr:         []string{"informational"},
		SoftwareDeps: []string{"android", "chrome_login"},
		Data:         []string{"capybara.jpg"},
		Timeout:      4 * time.Minute,
	})
}

func Downloads(ctx context.Context, s *testing.State) {
	const (
		filename    = "capybara.jpg"
		crosPath    = "/home/chronos/user/Downloads/" + filename
		androidPath = "/storage/emulated/0/Download/" + filename
	)

	cr, err := chrome.New(ctx, chrome.ARCEnabled())
	if err != nil {
		s.Fatal("Failed to connect to Chrome: ", err)
	}
	defer cr.Close(ctx)

	a, err := arc.New(ctx, s.OutDir())
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
	actual, err := a.ReadFile(ctx, androidPath)
	if err != nil {
		s.Error("CrOS -> Android failed: ", err)
	} else if !bytes.Equal(actual, expected) {
		s.Error("CrOS -> Android failed: content mismatch")
	}
	if err = os.Remove(crosPath); err != nil {
		s.Fatal("Failed to remove a file: ", err)
	}

	// Android -> CrOS
	if err = a.WriteFile(ctx, androidPath, expected); err != nil {
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
