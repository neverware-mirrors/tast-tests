// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"time"

	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         PackageInstallUninstall,
		Desc:         "Installs and then uninstalls a package that we have copied into the container",
		Contacts:     []string{"smbarber@chromium.org", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline"},
		Vars:         []string{"keepState"},
		Data:         []string{"package.deb"},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Params: []testing.Param{
			// Parameters generated by params_test.go. DO NOT EDIT.
			{
				Name:              "stretch_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "stretch_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "stretch_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "stretch_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
			},
		},
	})
}

func PackageInstallUninstall(ctx context.Context, s *testing.State) {
	cont := s.PreValue().(crostini.PreData).Container
	filePath := "/home/testuser/package.deb"
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	if err := crostini.TransferToContainer(ctx, cont, s.DataPath("package.deb"), filePath); err != nil {
		s.Fatal("Failed to transfer .deb to the container: ", err)
	}
	defer func() {
		if err := crostini.RemoveContainerFile(ctx, cont, filePath); err != nil {
			s.Fatal("Failed to remove .deb from the container: ", err)
		}
	}()

	var installedFiles = []string{
		"/usr/share/applications/x11_demo.desktop",
		"/usr/share/applications/x11_demo_fixed_size.desktop",
		"/usr/share/applications/wayland_demo.desktop",
		"/usr/share/applications/wayland_demo_fixed_size.desktop",
		"/usr/share/icons/hicolor/32x32/apps/x11_demo.png",
		"/usr/share/icons/hicolor/32x32/apps/wayland_demo.png",
	}
	const desktopFileID = "x11_demo"

	// Check the files are not present before we run.
	for _, testFile := range installedFiles {
		if err := crostini.VerifyFileNotInContainer(ctx, cont, testFile); err != nil {
			s.Errorf("Failed to check file absence of %q: %v", testFile, err)
		}
	}

	// Install the package.
	if err := cont.InstallPackage(ctx, filePath); err != nil {
		s.Fatal("Failed executing LinuxPackageInstall: ", err)
	}
	if err := cont.Command(ctx, "dpkg", "-s", "cros-tast-tests").Run(testexec.DumpLogOnError); err != nil {
		s.Error("Failed checking for cros-tast-tests in dpkg -s: ", err)
	}

	// Check the files are present once we install.
	for _, testFile := range installedFiles {
		if err := crostini.VerifyFileInContainer(ctx, cont, testFile); err != nil {
			s.Errorf("Failed to check file existence of %q: %v", testFile, err)
		}
	}

	// When uninstalling, we have to poll because we are racing against the
	// chrome process (which might not know we have installed the package
	// yet).
	if err := testing.Poll(ctx, func(context.Context) error {
		return cont.UninstallPackageOwningFile(ctx, desktopFileID)
	}, &testing.PollOptions{Timeout: 20 * time.Second}); err != nil {
		s.Fatal("Failed executing UninstallPackageOwningFile: ", err)
	}
	// Verify the package does not show up in the dpkg installed list.
	err := cont.Command(ctx, "dpkg", "-s", "cros-tast-tests").Run(testexec.DumpLogOnError)
	// A wait status of 1 indicates that the package could not be found. 0
	// indicates the package is still installed. Other wait statii indicate a dpkg
	// issue.
	if waitStatus, ok := testexec.GetWaitStatus(err); !ok {
		s.Fatal("Error running dpkg -s: ", err)
	} else if waitStatus.ExitStatus() == 0 {
		s.Fatal("The cros-tast-tests package is still installed")
	} else if waitStatus.ExitStatus() != 1 {
		s.Fatal("Internal dpkg error: ", err)
	}

	// Check the files are not present after we uninstall.
	for _, testFile := range installedFiles {
		if err := crostini.VerifyFileNotInContainer(ctx, cont, testFile); err != nil {
			s.Errorf("Failed to check file absence of %q: %v", testFile, err)
		}
	}
}
