// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vm

import (
	"context"
	"path/filepath"
	"time"

	"chromiumos/tast/local/bundles/cros/vm/subtest"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/colorcmp"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         CrostiniStartEverything,
		Desc:         "Tests Termina VM startup, container startup and other Crostini functionality",
		Attr:         []string{"informational"},
		Data:         []string{"cros-tast-tests-deb.deb"},
		Timeout:      10 * time.Minute,
		SoftwareDeps: []string{"chrome_login", "vm_host"},
	})
}

func CrostiniStartEverything(ctx context.Context, s *testing.State) {
	cr, err := chrome.New(ctx)
	if err != nil {
		s.Fatal("Failed to connect to Chrome: ", err)
	}
	defer cr.Close(ctx)

	s.Log("Enabling Crostini preference setting")
	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create test API connection: ", err)
	}
	if err = vm.EnableCrostini(ctx, tconn); err != nil {
		s.Fatal("Failed to enable Crostini preference setting: ", err)
	}

	s.Log("Setting up component ", vm.StagingComponent)
	err = vm.SetUpComponent(ctx, vm.StagingComponent)
	if err != nil {
		s.Fatal("Failed to set up component: ", err)
	}

	s.Log("Creating default container")
	cont, err := vm.CreateDefaultContainer(ctx, cr.User(), vm.StagingImageServer)
	if err != nil {
		s.Fatal("Failed to set up default container: ", err)
	}
	defer func() {
		if err := cont.DumpLog(ctx, s.OutDir()); err != nil {
			s.Error("Failure dumping container log: ", err)
		}
	}()

	s.Log("Verifying pwd command works")
	cmd := cont.Command(ctx, "pwd")
	if err = cmd.Run(); err != nil {
		cmd.DumpLog(ctx)
		s.Fatal("Failed to run pwd: ", err)
	}

	// The VM and container have started up so we can now execute all of the other
	// Crostini tests. We need to be careful about this because we are going to be
	// testing multiple things in one test. This should be done so that no tests
	// have any known dependency on prior tests. If we hit a conflict at some
	// point then we will need to add functionality to save the VM/container image
	// at this point and then stop the VM/container and restore that image so we
	// can have a clean VM/container to start from again. Failures should not be
	// fatal so that all tests can get executed.
	subtest.Webserver(ctx, s, cr, cont)
	subtest.LaunchTerminal(ctx, s, cr, cont)
	subtest.LaunchBrowser(ctx, s, cr, cont)
	subtest.VerifyAppFromTerminal(ctx, s, cr, cont, "x11", "/opt/google/cros-containers/bin/x11_demo",
		colorcmp.RGB(0x99, 0xee, 0x44))
	subtest.VerifyAppFromTerminal(ctx, s, cr, cont, "wayland", "/opt/google/cros-containers/bin/wayland_demo",
		colorcmp.RGB(0x33, 0x88, 0xdd))

	// Copy a test Debian package file to the container which will be used by
	// subsequent tests.
	const debianFilename = "cros-tast-tests-deb.deb"
	containerDebPath := filepath.Join("/home/testuser", debianFilename)
	err = cont.PushFile(ctx, s.DataPath(debianFilename), containerDebPath)
	if err != nil {
		s.Fatal("Failed copying test Debian package to container:", err)
	}

	subtest.LinuxPackageInfo(ctx, s, cont, containerDebPath)
	err = subtest.InstallPackage(ctx, cont, containerDebPath)
	if err != nil {
		s.Error("Failure in performing Linux package install", err)
	} else {
		// The application IDs below are generated by the code here:
		// https://cs.chromium.org/chromium/src/chrome/browser/chromeos/crostini/crostini_registry_service.cc?g=0&l=75
		// It's a modified SHA256 hash output of a concatentation of a constant,
		// the VM name, the container name and the identifier for the .desktop file
		// the app is associated with.
		const x11DemoName = "x11_demo"
		const x11DemoID = "glkpdbkfmomgogbfppaajjcgbcgaicmi"
		subtest.VerifyLauncherApp(ctx, s, cr, tconn, cont.VM.Concierge.GetOwnerID(),
			x11DemoName, x11DemoID, colorcmp.RGB(0x99, 0xee, 0x44))
		subtest.VerifyLauncherApp(ctx, s, cr, tconn, cont.VM.Concierge.GetOwnerID(),
			"wayland_demo", "nodabfiipdopnjihbfpiengllkohmfkl", colorcmp.RGB(0x33, 0x88, 0xdd))

		subtest.UninstallApplication(ctx, s, cont, cont.VM.Concierge.GetOwnerID(),
			x11DemoName, x11DemoID)
	}

	subtest.UninstallInvalidApplication(ctx, s, cont)
}
