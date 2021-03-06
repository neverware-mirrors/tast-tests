// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"fmt"
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
		Func:         AppVscode,
		Desc:         "Test Visual Studio Code in Terminal window",
		Contacts:     []string{"jinrongwu@google.com", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		Vars:         []string{"keepState"},
		SoftwareDeps: []string{"chrome", "vm_host", "amd64"},
		Params: []testing.Param{
			// Parameters generated by params_test.go. DO NOT EDIT.
			{
				Name:              "amd64",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_app_test_container_metadata_buster_amd64.tar.xz", "crostini_app_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniAppTest,
				Pre:               crostini.StartedByComponentBusterLargeContainer(),
				Timeout:           15 * time.Minute,
			}, {
				Name:              "arm",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_app_test_container_metadata_buster_arm.tar.xz", "crostini_app_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniAppTest,
				Pre:               crostini.StartedByComponentBusterLargeContainer(),
				Timeout:           15 * time.Minute,
			},
		},
	})
}
func AppVscode(ctx context.Context, s *testing.State) {
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
		// Restart Crostini in the end in case any error in the middle and Visual Studio Code is not closed.
		// This also closes the Terminal window.
		if restartIfError {
			if err := terminalApp.RestartCrostini(cleanupCtx, keyboard, cont, cr.User()); err != nil {
				s.Log("Failed to restart Crostini: ", err)
			}
		} else {
			terminalApp.Exit(cleanupCtx, keyboard)
		}
	}()

	if err := testCreateFileWithVSCode(ctx, terminalApp, keyboard, tconn, cont); err != nil {
		s.Fatal("Failed to create file with Visual Studio Code in Terminal: ", err)
	}

	restartIfError = false
}

func testCreateFileWithVSCode(ctx context.Context, terminalApp *terminalapp.TerminalApp, keyboard *input.KeyboardEventWriter, tconn *chrome.TestConn, cont *vm.Container) error {
	const (
		testFile   = "test.go"
		testString = "//This is a test string."
	)

	// Launch Visual Studio Code.
	cmd := fmt.Sprintf("code %s", testFile)
	if err := terminalApp.RunCommand(ctx, keyboard, cmd); err != nil {
		return errors.Wrapf(err, "failed to run command '%q' in Terminal window", cmd)
	}

	param := ui.FindParams{
		Name: fmt.Sprintf("● %s - Visual Studio Code", testFile),
		Role: ui.RoleTypeWindow,
	}

	// Find the app window.
	appWindow, err := ui.FindWithTimeout(ctx, tconn, param, 15*time.Second)
	if err != nil {
		return errors.Wrap(err, "failed to find the Visual Studio Code window")
	}

	// Sometimes left click could not focus on the new window. Moving the mouse first to make sure the cursor goes to the new window.
	if err = mouse.Move(ctx, tconn, appWindow.Location.CenterPoint(), 5*time.Second); err != nil {
		return errors.Wrap(err, "failed to move to the center of the Visual Studio Code window")
	}

	// Left click the app window.
	if err = appWindow.LeftClick(ctx); err != nil {
		return errors.Wrap(err, "failed left click on Visual Studio Code window")
	}

	// Type test string into the new file.
	if err = keyboard.Type(ctx, testString); err != nil {
		return errors.Wrapf(err, "failed to type %q in Visual Studio Code window", testString)
	}

	// Press ctrl+S to save the file.
	if err = keyboard.Accel(ctx, "ctrl+S"); err != nil {
		return errors.Wrap(err, "failed to press ctrl+S in Visual Studio Code window")
	}

	// Press ctrl+W twice to exit window.
	if err = keyboard.Accel(ctx, "ctrl+W"); err != nil {
		return errors.Wrap(err, "failed to press ctrl+W in Visual Studio Code window")
	}
	if err = keyboard.Accel(ctx, "ctrl+W"); err != nil {
		return errors.Wrap(err, "failed to press ctrl+W in Visual Studio Code window")
	}

	if err = ui.WaitUntilGone(ctx, tconn, param, 15*time.Second); err != nil {
		return errors.Wrap(err, "failed to close Visual Studio Code window")
	}

	// Check the content of the test file.
	if err := cont.CheckFileContent(ctx, testFile, testString); err != nil {
		return errors.Wrap(err, "failed to verify the content of the file")
	}

	return nil
}
