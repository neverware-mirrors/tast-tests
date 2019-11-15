// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ui

import (
	"context"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ash"
	"chromiumos/tast/local/chrome/display"
	"chromiumos/tast/local/chrome/metrics"
	"chromiumos/tast/local/input"
	"chromiumos/tast/local/media/cpu"
	"chromiumos/tast/local/perf"
	"chromiumos/tast/local/ui"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         OverviewScrollPerf,
		Desc:         "Measures the presentation time of scrolling the overview grid in tablet mode",
		Contacts:     []string{"sammiequon@chromium.org", "chromeos-wmp@google.com"},
		Attr:         []string{"group:crosbolt", "crosbolt_perbuild"},
		SoftwareDeps: []string{"chrome", "tablet_mode"},
		Pre:          chrome.LoggedIn(),
	})
}

func OverviewScrollPerf(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*chrome.Chrome)

	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to connect to test API: ", err)
	}
	defer tconn.Close()

	rotation, err := display.GetPanelRotation(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to obtain the panel rotation: ", err)
	}

	tabletModeEnabled, err := ash.TabletModeEnabled(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to get tablet mode: ", err)
	}
	// Be nice and restore tablet mode to its original state on exit.
	defer ash.SetTabletModeEnabled(ctx, tconn, tabletModeEnabled)

	// Overview scrolling is only available in tablet mode.
	if err = ash.SetTabletModeEnabled(ctx, tconn, true); err != nil {
		s.Fatal("Failed to enable tablet mode: ", err)
	}

	// Prepare the touch screen as this test requires touch scroll events.
	tsew, err := input.Touchscreen(ctx)
	if err != nil {
		s.Fatal("Failed to create touch screen event writer: ", err)
	}
	if err = tsew.SetRotation(-rotation); err != nil {
		s.Fatal("Failed to set rotation: ", err)
	}

	stw, err := tsew.NewSingleTouchWriter()
	if err != nil {
		s.Fatal("Failed to create single touch writer: ", err)
	}
	defer stw.Close()

	// Use a total of 16 windows for this test, so that scrolling can happen.
	const numWindows = 16
	conns, err := ash.CreateWindows(ctx, cr, ui.PerftestURL, numWindows)
	if err != nil {
		s.Fatal("Failed to open browser windows: ", err)
	}
	defer conns.Close()

	if err := cpu.WaitUntilIdle(ctx); err != nil {
		s.Fatal("Failed waiting for CPU to become idle: ", err)
	}

	if err = ash.SetOverviewModeAndWait(ctx, tconn, true); err != nil {
		s.Fatal("It does not appear to be in the overview mode: ", err)
	}

	histograms, err := metrics.Run(ctx, cr, func() error {
		// Scroll from the top right of the screen to the top left.
		if err := stw.Swipe(ctx, tsew.Width()-10, 10, 10, 10, 500*time.Millisecond); err != nil {
			return errors.Wrap(err, "failed to execute a swipe gesture")
		}

		if err := stw.End(); err != nil {
			return errors.Wrap(err, "failed to finish the swipe gesture")
		}

		return nil
	}, "Ash.Overview.Scroll.PresentationTime.TabletMode")
	if err != nil {
		s.Fatal("Failed to scroll or get histogram: ", err)
	}

	if err = ash.SetOverviewModeAndWait(ctx, tconn, false); err != nil {
		s.Fatal("It does not appear to be in the overview mode: ", err)
	}

	pv := perf.NewValues()
	pv.Set(perf.Metric{
		Name:      histograms[0].Name,
		Unit:      "ms",
		Direction: perf.SmallerIsBetter,
	}, histograms[0].Mean())

	if err := pv.Save(s.OutDir()); err != nil {
		s.Error("Failed saving perf data: ", err)
	}
}
