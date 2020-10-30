// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package arcappcompat will have tast tests for android apps on Chromebooks.
package arcappcompat

import (
	"context"
	"time"

	"chromiumos/tast/local/android/ui"
	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/bundles/cros/arcappcompat/pre"
	"chromiumos/tast/local/bundles/cros/arcappcompat/testutil"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

// ClamshellTests are placed here.
var clamshellTestsForDiscord = []testutil.TestCase{
	{Name: "Launch app in Clamshell", Fn: launchAppForDiscord},
	{Name: "Clamshell: Fullscreen app", Fn: testutil.ClamshellFullscreenApp},
	{Name: "Clamshell: Minimise and Restore", Fn: testutil.MinimizeRestoreApp},
	{Name: "Clamshell: Reopen app", Fn: testutil.ReOpenWindow},
	{Name: "Clamshell: Resize window", Fn: testutil.ClamshellResizeWindow},
}

// TouchviewTests are placed here.
var touchviewTestsForDiscord = []testutil.TestCase{
	{Name: "Launch app in Touchview", Fn: launchAppForDiscord},
	{Name: "Touchview: Minimise and Restore", Fn: testutil.MinimizeRestoreApp},
	{Name: "Touchview: Reopen app", Fn: testutil.ReOpenWindow},
}

func init() {
	testing.AddTest(&testing.Test{
		Func:         Discord,
		Desc:         "Functional test for Discord that installs the app also verifies it is logged in and that the main page is open, checks Discord correctly changes the window state in both clamshell and touchview mode",
		Contacts:     []string{"mthiyagarajan@chromium.org", "cros-appcompat-test-team@google.com"},
		Attr:         []string{"group:appcompat"},
		SoftwareDeps: []string{"chrome"},
		Params: []testing.Param{{
			Val:               clamshellTestsForDiscord,
			ExtraSoftwareDeps: []string{"android_p"},
			Pre:               pre.AppCompatBooted,
		}, {
			Name:              "tablet_mode",
			Val:               touchviewTestsForDiscord,
			ExtraSoftwareDeps: []string{"android_p", "tablet_mode"},
			Pre:               pre.AppCompatBootedInTabletMode,
		}, {
			Name:              "vm",
			Val:               clamshellTestsForDiscord,
			ExtraSoftwareDeps: []string{"android_vm"},
			Pre:               pre.AppCompatBooted,
		}, {
			Name:              "vm_tablet_mode",
			Val:               touchviewTestsForDiscord,
			ExtraSoftwareDeps: []string{"android_vm", "tablet_mode"},
			Pre:               pre.AppCompatBootedInTabletMode,
		}},
		Timeout: 10 * time.Minute,
		Vars: []string{"arcappcompat.username", "arcappcompat.password",
			"arcappcompat.Discord.emailid", "arcappcompat.Discord.password"},
	})
}

// Discord test uses library for opting into the playstore and installing app.
// Checks Discord correctly changes the window states in both clamshell and touchview mode.
func Discord(ctx context.Context, s *testing.State) {
	const (
		appPkgName  = "com.discord"
		appActivity = ".app.AppActivity$Main"
	)
	testCases := s.Param().([]testutil.TestCase)
	testutil.RunTestCases(ctx, s, appPkgName, appActivity, testCases)
}

// launchAppForDiscord verifies Discord is logged in and
// verify Discord reached main activity page of the app.
func launchAppForDiscord(ctx context.Context, s *testing.State, tconn *chrome.TestConn, a *arc.ARC, d *ui.Device, appPkgName, appActivity string) {
	const (
		signInText        = "Login"
		textEditClassName = "android.widget.EditText"
		enterEmailText    = "Email"
		enterPasswordText = "Password"
		homeIconID        = "com.discord:id/tabs_host_bottom_nav_friends_item"
		notNowID          = "android:id/autofill_save_no"
		verifyCaptchaID   = "com.discord:id/auth_captcha_verify"
	)

	// Click on sign in button.
	signInButton := d.Object(ui.ClassName(testutil.AndroidButtonClassName), ui.Text(signInText))
	if err := signInButton.WaitForExists(ctx, testutil.DefaultUITimeout); err != nil {
		s.Error("SignIn Button doesn't exist: ", err)
	} else if err := signInButton.Click(ctx); err != nil {
		s.Fatal("Failed to click on signInButton: ", err)
	}

	// Enter email address.
	DiscordEmailID := s.RequiredVar("arcappcompat.Discord.emailid")
	enterEmailAddress := d.Object(ui.ClassName(textEditClassName), ui.Text(enterEmailText))
	if err := enterEmailAddress.WaitForExists(ctx, testutil.DefaultUITimeout); err != nil {
		s.Error("EnterEmailAddress doesn't exist: ", err)
	} else if err := enterEmailAddress.Click(ctx); err != nil {
		s.Fatal("Failed to click on enterEmailAddress: ", err)
	} else if err := enterEmailAddress.SetText(ctx, DiscordEmailID); err != nil {
		s.Fatal("Failed to enterEmailAddress: ", err)
	}

	// Enter Password.
	DiscordPassword := s.RequiredVar("arcappcompat.Discord.password")
	enterPassword := d.Object(ui.ClassName(textEditClassName), ui.Text(enterPasswordText))
	if err := enterPassword.WaitForExists(ctx, testutil.DefaultUITimeout); err != nil {
		s.Error("EnterPassword doesn't exist: ", err)
	} else if err := enterPassword.Click(ctx); err != nil {
		s.Fatal("Failed to click on enterPassword: ", err)
	} else if err := enterPassword.SetText(ctx, DiscordPassword); err != nil {
		s.Fatal("Failed to enterPassword: ", err)
	}

	// Click on sign in button.
	if err := signInButton.WaitForExists(ctx, testutil.DefaultUITimeout); err != nil {
		s.Error("SignIn Button doesn't exist: ", err)
	} else if err := signInButton.Click(ctx); err != nil {
		s.Fatal("Failed to click on signInButton: ", err)
	}

	// Click on not now button.
	notNowButton := d.Object(ui.ID(notNowID))
	if err := notNowButton.WaitForExists(ctx, testutil.DefaultUITimeout); err != nil {
		s.Log("notNowButton doesn't exists: ", err)
	} else if err := notNowButton.Click(ctx); err != nil {
		s.Fatal("Failed to click on notNowButton: ", err)
	}

	// Check for captcha.
	verifyCaptcha := d.Object(ui.ID(verifyCaptchaID))
	if err := verifyCaptcha.WaitForExists(ctx, testutil.DefaultUITimeout); err != nil {
		s.Log("verifyCaptcha doesn't exist: ", err)
		// Check for home icon.
		homeIcon := d.Object(ui.ID(homeIconID))
		if err := homeIcon.WaitForExists(ctx, testutil.LongUITimeout); err != nil {
			s.Error("HomeIcon doesn't exist: ", err)
		}
	} else {
		s.Log("Verify by reCaptcha exists")
	}

}