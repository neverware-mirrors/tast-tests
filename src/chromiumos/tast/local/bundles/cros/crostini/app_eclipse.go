// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/terminalapp"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     AppEclipse,
		Desc:     "Test Eclipse in Terminal window",
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
func AppEclipse(ctx context.Context, s *testing.State) {
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
		// Restart Crostini in the end because it is not possible to control the Crostini app.
		// TODO(jinrongwu): modify this once it is possible to control Eclipse.
		if err := terminalApp.RestartCrostini(cleanupCtx, keyboard, cont, cr.User()); err != nil {
			s.Fatal("Failed to restart Crostini: ", err)
		}
	}()

	// Create a workspace and a test file.
	const (
		workspace = "ws"
		testFile  = "test.java"
	)
	if err := cont.Command(ctx, "mkdir", workspace).Run(testexec.DumpLogOnError); err != nil {
		s.Fatal("Failed to create workspace directory in the Container")
	}
	if err := cont.Command(ctx, "touch", fmt.Sprintf("%s/%s", workspace, testFile)).Run(testexec.DumpLogOnError); err != nil {
		s.Fatal("Failed to create test file in the Container: ", err)
	}

	// Open eclipse.
	if err := terminalApp.RunCommand(ctx, keyboard, fmt.Sprintf("eclipse -data %s --launcher.openFile %s/%s --noSplash", workspace, workspace, testFile)); err != nil {
		s.Fatal("Failed to start eclipse in Terminal: ", err)
	}

	// Find eclipse window.
	param := ui.FindParams{
		Name: fmt.Sprintf("%s - /home/%s/%s/%s - Eclipse IDE ", workspace, strings.Split(cr.User(), "@")[0], workspace, testFile),
		Role: ui.RoleTypeWindow,
	}
	if _, err := ui.FindWithTimeout(ctx, tconn, param, 90*time.Second); err != nil {
		s.Fatal("Failed to find eclipse window: ", err)
	}

	//TODO(jinrongwu): UI test on eclipse.
}
