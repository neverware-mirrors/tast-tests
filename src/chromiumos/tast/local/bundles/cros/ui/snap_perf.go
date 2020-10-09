// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ui

import (
	"context"

	"chromiumos/tast/local/bundles/cros/ui/perfutil"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ash"
	"chromiumos/tast/local/chrome/cdputil"
	"chromiumos/tast/local/power"
	"chromiumos/tast/local/ui"
	"chromiumos/tast/testing"
	"chromiumos/tast/testing/hwdep"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         SnapPerf,
		Desc:         "Measures the animation smoothess of snapping windows in clamshell mode",
		Contacts:     []string{"sammiequon@chromium.org", "chromeos-wmp@google.com"},
		Attr:         []string{"group:crosbolt", "crosbolt_perbuild"},
		SoftwareDeps: []string{"chrome"},
		HardwareDeps: hwdep.D(hwdep.InternalDisplay()),
		Pre:          chrome.LoggedIn(),
	})
}

func SnapPerf(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*chrome.Chrome)

	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to connect to test API: ", err)
	}

	conn, err := cr.NewConn(ctx, ui.PerftestURL, cdputil.WithNewWindow())
	if err != nil {
		s.Fatal("Failed to open a new connection: ", err)
	}
	defer conn.Close()

	cleanup, err := ash.EnsureTabletModeEnabled(ctx, tconn, false)
	if err != nil {
		s.Fatal("Failed to ensure in clamshell mode: ", err)
	}
	defer cleanup(ctx)

	if _, err := power.WaitUntilCPUCoolDown(ctx, power.CoolDownPreserveUI); err != nil {
		s.Fatal("Failed waiting for CPU to become idle: ", err)
	}

	window, err := ash.FindWindow(ctx, tconn, func(w *ash.Window) bool { return true })
	if err != nil {
		s.Fatal("Failed to obtain the window list: ", err)
	}
	pv := perfutil.RunMultiple(ctx, s, cr, perfutil.RunAndWaitAll(tconn, func(ctx context.Context) error {
		// Snap the window to the left.
		if err := ash.SetWindowStateAndWait(ctx, tconn, window.ID, ash.WindowStateLeftSnapped); err != nil {
			s.Fatalf("Failed to set the window (%d): %v", window.ID, err)
		}

		// Restore the normal state bounds, as no animation stats will be logged if the window size does not change.
		if err := ash.SetWindowStateAndWait(ctx, tconn, window.ID, ash.WindowStateNormal); err != nil {
			s.Fatalf("Failed to set the window (%d): %v", window.ID, err)
		}

		// Snap the window to the right.
		if err := ash.SetWindowStateAndWait(ctx, tconn, window.ID, ash.WindowStateRightSnapped); err != nil {
			s.Fatalf("Failed to set the window (%d): %v", window.ID, err)
		}

		// Restore the normal state bounds, as no animation stats will be logged if the window size does not change.
		if err := ash.SetWindowStateAndWait(ctx, tconn, window.ID, ash.WindowStateNormal); err != nil {
			s.Fatalf("Failed to set the window (%d): %v", window.ID, err)
		}

		return nil
	},
		"Ash.Window.AnimationSmoothness.Snap"), perfutil.StoreSmoothness)

	if err := pv.Save(ctx, s.OutDir()); err != nil {
		s.Error("Failed saving perf data: ", err)
	}
}
