// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"time"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/terminalapp"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     AppAndroidStudio,
		Desc:     "Test android studio in Terminal window",
		Contacts: []string{"jinrongwu@google.com", "cros-containers-dev@google.com"},
		Attr:     []string{"group:mainline", "informational"},
		Vars:     []string{"keepState"},
		Params: []testing.Param{{
			Name:              "download_buster",
			Pre:               crostini.StartedByDownloadBusterLargeContainer(),
			ExtraHardwareDeps: crostini.CrostiniAppTest,
			Timeout:           15 * time.Minute,
		}},
		SoftwareDeps: []string{"chrome", "vm_host", "amd64"},
	})
}
func AppAndroidStudio(ctx context.Context, s *testing.State) {
	tconn := s.PreValue().(crostini.PreData).TestAPIConn
	cr := s.PreValue().(crostini.PreData).Chrome
	keyboard := s.PreValue().(crostini.PreData).Keyboard
	cont := s.PreValue().(crostini.PreData).Container

	// Use a shortened context for test operations to reserve time for cleanup.
	cleanupCtx := ctx
	ctx, cancel := ctxutil.Shorten(ctx, 90*time.Second)
	defer cancel()
	defer crostini.RunCrostiniPostTest(cleanupCtx, s.PreValue().(crostini.PreData))

	// Open Terminal app.
	terminalApp, err := terminalapp.Launch(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to open Terminal app: ", err)
	}
	defer func() {
		// Restart crostini in the end to close all Android Studio related windows.
		// This could be skipped once UI test is implemented against Crostini apps.
		if err := terminalApp.RestartCrostini(cleanupCtx, keyboard, cont, cr.User()); err != nil {
			s.Fatal("Failed to restart crostini: ", err)
		}
	}()

	// Open android studio.
	if err := terminalApp.RunCommand(ctx, keyboard, "/android-studio/bin/studio.sh &"); err != nil {
		s.Fatal("Failed to start android studio in Terminal: ", err)
	}

	// Find window.
	param := ui.FindParams{
		Name: "Import Android Studio Settings From...",
		Role: ui.RoleTypeWindow,
	}
	if _, err := ui.FindWithTimeout(ctx, tconn, param, 30*time.Second); err != nil {
		s.Fatal("Failed to find android studio window: ", err)
	}

	//TODO(jinrongwu): UI test on android studio code.
}