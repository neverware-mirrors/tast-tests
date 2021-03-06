// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shelf

import (
	"context"

	"chromiumos/tast/local/apps"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ash"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/chrome/ui/faillog"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: LaunchedApps,
		Desc: "Checks that launched apps appear in the shelf",
		Contacts: []string{
			"chromeos-sw-engprod@google.com",
			"tbarzic@chromium.org",
		},
		Attr:         []string{"group:mainline"},
		SoftwareDeps: []string{"chrome"},
		Pre:          chrome.LoggedIn(),
	})
}

// LaunchedApps tests that apps launched appear in the ChromeOS shelf.
func LaunchedApps(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*chrome.Chrome)
	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create Test API connection: ", err)
	}
	defer faillog.DumpUITreeOnError(ctx, s.OutDir(), s.HasError, tconn)

	// At login, we should have just Chrome and maybe Files in the Shelf.
	shelfItems, err := ash.ShelfItems(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to get shelf items: ", err)
	}
	if len(shelfItems) != 1 {
		if len(shelfItems) != 2 || shelfItems[1].AppID != apps.Files.ID {
			s.Fatal("Unexpected apps in the shelf. Expected only Chrome and Files: ", shelfItems)
		}
	}
	// Get the expected browser.
	chromeApp, err := apps.ChromeOrChromium(ctx, tconn)
	if err != nil {
		s.Fatal("Could not find the Chrome app: ", err)
	}

	// Chrome app name doesn't exactly match the chrome shelf name so modify it here for simpler code later.
	if chromeApp.Name == apps.Chrome.Name {
		chromeApp.Name = "Google Chrome"
	}
	// Chrome must be first because it is automatically opened upon login.
	defaultApps := []apps.App{chromeApp, apps.Files, apps.WallpaperPicker}

	for _, app := range defaultApps {
		s.Logf("Launching %s", app.Name)
		if err := apps.Launch(ctx, tconn, app.ID); err != nil {
			s.Fatalf("Failed to launch %s: %s", app.Name, err)
		}
		if err := ash.WaitForApp(ctx, tconn, app.ID); err != nil {
			s.Fatalf("%s did not appear in shelf after launch: %s", app.Name, err)
		}
	}

	// Get the list of apps in the shelf via API.
	shelfItems, err = ash.ShelfItems(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to get shelf items: ", err)
	}

	s.Log("Checking that all expected apps are in the shelf")
	if len(shelfItems) != len(defaultApps) {
		s.Fatalf("Shelf items count does not match expected apps. Got: %v; Want: %v", len(shelfItems), len(defaultApps))
	}
	for i, shelfItem := range shelfItems {
		expectedApp := defaultApps[i]
		if shelfItem.AppID != expectedApp.ID {
			s.Errorf("App IDs did not match. Got: %v; Want: %v", shelfItem.AppID, expectedApp.ID)
		}
		if shelfItem.Title != expectedApp.Name {
			s.Errorf("App names did not match. Got: %v; Want: %v", shelfItem.Title, expectedApp.Name)
		}
	}

	icons, err := ui.FindAll(ctx, tconn, ui.FindParams{Role: ui.RoleTypeButton})
	if err != nil {
		s.Fatal("Failed to get all buttons: ", err)
	}
	defer icons.Release(ctx)

	// Check that the icons are also present in the UI
	for _, app := range defaultApps {
		var found = false
		for _, icon := range icons {
			if icon.Name == app.Name {
				s.Logf("Found icon for %s", app.Name)
				found = true
				break
			}
		}
		if !found {
			s.Errorf("There was no icon for %s in the shelf", app.Name)
		}
	}
}
