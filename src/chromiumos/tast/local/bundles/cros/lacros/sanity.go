// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package lacros tests linux-chrome running on ChromeOS.
package lacros

import (
	"context"
	"time"

	"chromiumos/tast/local/bundles/cros/lacros/launcher"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         Sanity,
		Desc:         "Tests basic lacros startup",
		Contacts:     []string{"erikchen@chromium.org", "hidehiko@chromium.org", "edcourtney@chromium.org", "lacros-team@google.com"},
		SoftwareDeps: []string{"chrome"},
		Attr:         []string{"disabled"},
		Pre:          launcher.StartedByData(),
		Timeout:      7 * time.Minute,
		Data:         []string{launcher.DataArtifact},
	})
}

func Sanity(ctx context.Context, s *testing.State) {
	l, err := launcher.LaunchLinuxChrome(ctx, s.PreValue().(launcher.PreData))
	if err != nil {
		s.Fatal("Failed to launch linux-chrome")
	}
	defer l.Close(ctx)

	if _, err = l.Devsess.CreateTarget(ctx, "about:blank"); err != nil {
		s.Fatal("Failed to open new tab: ", err)
	}
}
