// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"time"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/chrome/ui/mouse"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/terminalapp"
	"chromiumos/tast/local/input"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     AppGedit,
		Desc:     "Test gedit in Terminal window",
		Contacts: []string{"jinrongwu@google.com", "cros-containers-dev@google.com"},
		Attr:     []string{"group:mainline", "informational"},
		Vars:     []string{"keepState"},
		Params: []testing.Param{{
			Name:              "download_buster",
			Pre:               crostini.StartedByDownloadBusterLargeContainer(),
			ExtraHardwareDeps: crostini.CrostiniAppTest,
			Timeout:           15 * time.Minute,
		}},
		SoftwareDeps: []string{"chrome", "vm_host"},
	})
}
func AppGedit(ctx context.Context, s *testing.State) {
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

	restartIfError := true

	defer func() {
		// Restart crostini in the end in case any error in the middle and gedit is not closed.
		// This also closes the Terminal window.
		if restartIfError {
			if err := terminalApp.RestartCrostini(cleanupCtx, keyboard, cont, cr.User()); err != nil {
				s.Fatal("Failed to restart crostini: ", err)
			}
		}
	}()

	// Create a file using gedit in Terminal.
	if err := testCreateFileWithGedit(ctx, terminalApp, keyboard, tconn, cont); err != nil {
		s.Fatal("Failed to create file with gedit in Terminal: ", err)
	}

	restartIfError = false

}

func testCreateFileWithGedit(ctx context.Context, terminalApp *terminalapp.TerminalApp, keyboard *input.KeyboardEventWriter, tconn *chrome.TestConn, cont *vm.Container) error {
	const (
		testFile   = "test.txt"
		testString = "This is a test string"
		uiString   = testFile + " (~/) - gedit"
	)

	// Launch Gedit.
	if err := terminalApp.RunCommand(ctx, keyboard, "gedit "+testFile); err != nil {
		return errors.Wrapf(err, "failed to run command %q in Terminal window", "gedit "+testFile)
	}
	// Find the app window.
	appWindow, err := ui.FindWithTimeout(ctx, tconn, ui.FindParams{Name: uiString, Role: ui.RoleTypeWindow}, 15*time.Second)
	if err != nil {
		return errors.Wrap(err, "failed to find the app window")
	}
	defer appWindow.Release(ctx)

	// Sometimes left click could not focus on the new window. Moving the mouse first to make sure the cursor goes to the app window.
	if err := mouse.Move(ctx, tconn, appWindow.Location.CenterPoint(), 5*time.Second); err != nil {
		return errors.Wrap(err, "failed to move to the center of the app window")
	}

	// Left click the app window.
	if err := appWindow.LeftClick(ctx); err != nil {
		return errors.Wrap(err, "failed left click on the app window")
	}

	// Type test string into the new file.
	if err := keyboard.Type(ctx, testString); err != nil {
		return errors.Wrapf(err, "failed to type %q into the app window", testString)
	}

	// Press ctrl+S to save the file.
	if err := keyboard.Accel(ctx, "ctrl+S"); err != nil {
		return errors.Wrap(err, "failed to press ctrl+S on the app window")
	}

	// Press ctrl+W twice to exit window.
	if err = keyboard.Accel(ctx, "ctrl+W"); err != nil {
		return errors.Wrap(err, "failed to press ctrl+W on the app window")
	}
	if err = keyboard.Accel(ctx, "ctrl+W"); err != nil {
		return errors.Wrap(err, "failed to press ctrl+W on the app window")
	}

	if err = ui.WaitUntilGone(ctx, tconn, ui.FindParams{Name: uiString, Role: ui.RoleTypeWindow}, 15*time.Second); err != nil {
		return errors.Wrap(err, "failed to close Gedit window")
	}

	// Check the content of the test file.
	if err := cont.CheckFileContent(ctx, testFile, testString+"\n"); err != nil {
		return errors.Wrap(err, "failed to verify the content of the test file")
	}

	return nil
}
