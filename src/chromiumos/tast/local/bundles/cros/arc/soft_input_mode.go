// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"context"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/android/ui"
	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/chrome/ash"
	"chromiumos/tast/local/chrome/display"
	"chromiumos/tast/local/chrome/vkb"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         SoftInputMode,
		Desc:         "Verifies that Ash split view works properly with softInputMode=adjustPan|adjustResize activity flags",
		Contacts:     []string{"tetsui@chromium.org", "arc-framework+tast@google.com"},
		Attr:         []string{"informational", "group:mainline"},
		SoftwareDeps: []string{"chrome"},
		Pre:          arc.BootedInTabletMode(),
		Params: []testing.Param{{
			ExtraSoftwareDeps: []string{"android_p"},
		}, {
			Name:              "vm",
			ExtraSoftwareDeps: []string{"android_vm"},
		}},
	})
}

func SoftInputMode(ctx context.Context, s *testing.State) {
	p := s.PreValue().(arc.PreData)
	cr := p.Chrome
	a := p.ARC

	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Creating test API connection failed: ", err)
	}

	d, err := a.NewUIDevice(ctx)
	if err != nil {
		s.Fatal("Failed initializing UI Automator: ", err)
	}
	defer d.Close()

	s.Log("Installing app")
	const apk = "ArcSoftInputModeTest.apk"
	if err := a.Install(ctx, arc.APKPath(apk)); err != nil {
		s.Fatal("Failed installing app: ", err)
	}

	infos, err := display.GetInfo(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to get display info: ", err)
	}
	if len(infos) == 0 {
		s.Fatal("No display found")
	}
	var info *display.Info
	for i := range infos {
		if infos[i].IsInternal {
			info = &infos[i]
		}
	}
	if info == nil {
		s.Log("No internal display found. Default to the first display")
		info = &infos[0]
	}

	waitForRotation := func(expectLandscape bool) error {
		return testing.Poll(ctx, func(ctx context.Context) error {
			disp, err := arc.NewDisplay(a, arc.DefaultDisplayID)
			if err != nil {
				return testing.PollBreak(err)
			}
			defer disp.Close()
			s, err := disp.Size(ctx)
			if err != nil {
				// It may return error while transition, keep retrying.
				return err
			}
			if s.Width > s.Height == expectLandscape {
				return nil
			}

			return errors.New("display not rotated in ARC")
		}, nil)
	}

	// TODO(tetsui): Use camera position for getting the default orientation.
	portraitByDefault := info.Bounds.Height > info.Bounds.Width

	runTest := func(ctx context.Context, s *testing.State, activityName string, rotation int) {
		// When the device is portrait by default, rotate for additional 90 degrees.
		actualRotation := rotation
		if portraitByDefault {
			actualRotation = (rotation + 90) % 360
		}
		if err := display.SetDisplayProperties(ctx, tconn, info.ID,
			display.DisplayProperties{Rotation: &actualRotation}); err != nil {
			s.Fatalf("Failed to rotate display to %d: %q", actualRotation, err)
		}

		if err := waitForRotation(rotation%180 == 0); err != nil {
			s.Fatal("Failed to wait for rotation: ", err)
		}

		firstAct, err := arc.NewActivity(a, "com.android.settings", ".Settings")
		if err != nil {
			s.Fatal("Failed to create a new activity: ", err)
		}
		defer firstAct.Close()

		if err := firstAct.Start(ctx, tconn); err != nil {
			s.Fatal("Failed to start the activity: ", err)
		}
		defer firstAct.Stop(ctx, tconn)

		const pkg = "org.chromium.arc.testapp.softinputmode"
		secondAct, err := arc.NewActivity(a, pkg, activityName)
		if err != nil {
			s.Fatal("Failed to create a new activity: ", err)
		}
		defer secondAct.Close()

		if err := secondAct.Start(ctx, tconn); err != nil {
			s.Fatal("Failed to start the activity: ", err)
		}
		defer secondAct.Stop(ctx, tconn)

		if _, err := ash.SetARCAppWindowState(ctx, tconn, secondAct.PackageName(), ash.WMEventSnapRight); err != nil {
			s.Fatal("Failed to snap app in split view: ", err)
		}

		if _, err := ash.SetARCAppWindowState(ctx, tconn, firstAct.PackageName(), ash.WMEventSnapLeft); err != nil {
			s.Fatal("Failed to snap app in split view: ", err)
		}

		const fieldID = "org.chromium.arc.testapp.softinputmode:id/text"
		field := d.Object(ui.ID(fieldID))
		if err := field.WaitForExists(ctx, 30*time.Second); err != nil {
			s.Fatal("Failed to find field: ", err)
		}
		if err := field.Click(ctx); err != nil {
			s.Fatal("Failed to click the field: ", err)
		}
		if err := vkb.WaitLocationStable(ctx, tconn); err != nil {
			s.Fatal("Failed to wait for the virtual keyboard to show: ", err)
		}

		if err := field.Exists(ctx); err != nil {
			s.Fatal("Could not find the field; probably hidden by the virtual keyboard?")
		}
	}

	// Restore the initial rotation.
	defer func() {
		if err := display.SetDisplayProperties(ctx, tconn, info.ID,
			display.DisplayProperties{Rotation: &info.Rotation}); err != nil {
			s.Fatal("Failed to restore the initial display rotation: ", err)
		}
	}()

	for _, data := range []struct {
		activityName string
		rotation     int
	}{
		{".AdjustPanActivity", 0},
		{".AdjustResizeActivity", 270},
	} {
		s.Run(ctx, data.activityName, func(ctx context.Context, s *testing.State) {
			runTest(ctx, s, data.activityName, data.rotation)
		})
	}
}
