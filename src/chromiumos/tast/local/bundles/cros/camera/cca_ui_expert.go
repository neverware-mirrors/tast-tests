// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package camera

import (
	"context"
	"time"

	"chromiumos/tast/local/bundles/cros/camera/cca"
	"chromiumos/tast/local/bundles/cros/camera/testutil"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         CCAUIExpert,
		Desc:         "Opens CCA and verifies the expert options",
		Contacts:     []string{"inker@chromium.org", "shik@chromium.org", "chromeos-camera-eng@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome", "arc_camera3", caps.BuiltinOrVividCamera},
		Data:         []string{"cca_ui.js"},
		Params: []testing.Param{{
			Pre: testutil.ChromeWithPlatformApp(),
			Val: testutil.PlatformApp,
		}, {
			Name: "swa",
			Pre:  testutil.ChromeWithSWA(),
			Val:  testutil.SWA,
		}},
	})
}

func CCAUIExpert(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*chrome.Chrome)
	useSWA := s.Param().(testutil.CCAAppType) == testutil.SWA
	tb, err := testutil.NewTestBridge(ctx, cr, useSWA)
	if err != nil {
		s.Fatal("Failed to construct test bridge: ", err)
	}
	defer tb.TearDown(ctx)

	if err := cca.ClearSavedDirs(ctx, cr); err != nil {
		s.Fatal("Failed to clear saved directory: ", err)
	}

	app, err := cca.New(ctx, cr, []string{s.DataPath("cca_ui.js")}, s.OutDir(), tb, useSWA)
	if err != nil {
		s.Fatal("Failed to open CCA: ", err)
	}
	defer func(ctx context.Context) {
		if err := app.Close(ctx); err != nil {
			s.Error("Failed to close app: ", err)
		}
	}(ctx)

	for i, action := range []struct {
		Name    string
		Func    func(context.Context, *cca.App) error
		Enabled bool
	}{
		// Expert mode is not reset after each test for persistency
		{"toggleExpertMode", toggleExpertMode, false},
		{"toggleExpertModeOptions", toggleExpertModeOptions, true},
		{"switchSquareMode", switchSquareMode, true},
		{"toggleExpertMode", toggleExpertMode, false},
		{"toggleExpertMode", toggleExpertMode, true},
		{"toggleExpertModeOptions", toggleExpertModeOptions, false},
	} {
		if err := action.Func(ctx, app); err != nil {
			s.Fatalf("Failed to perform action %v of test %v: %v", action.Name, i, err)
		}
		if err := verifyExpertMode(ctx, app, action.Enabled); err != nil {
			s.Errorf("Failed in test %v %v(): %v", i, action.Name, err)
		}
	}
}

func verifyExpertMode(ctx context.Context, app *cca.App, enabled bool) error {
	if err := app.CheckMetadataVisibility(ctx, enabled); err != nil {
		return err
	}
	if _, err := app.TakeSinglePhoto(ctx, cca.TimerOff); err != nil {
		return err
	}
	return nil
}

func toggleExpertMode(ctx context.Context, app *cca.App) error {
	_, err := app.ToggleExpertMode(ctx)
	// TODO(crbug.com/1039991): There are asynchronous mojo IPC calls happens
	// after toggling, and we don't have a way to poll it properly without
	// significantly refactor the logic.
	testing.Sleep(ctx, time.Second)
	return err
}

func toggleExpertModeOptions(ctx context.Context, app *cca.App) error {
	if _, err := app.ToggleShowMetadata(ctx); err != nil {
		return err
	}
	if _, err := app.ToggleSaveMetadata(ctx); err != nil {
		return err
	}
	return nil
}

func switchSquareMode(ctx context.Context, app *cca.App) error {
	return app.SwitchMode(ctx, cca.Square)
}
