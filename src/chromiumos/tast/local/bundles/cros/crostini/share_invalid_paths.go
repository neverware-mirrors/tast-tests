// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"chromiumos/tast/local/crostini"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         ShareInvalidPaths,
		Desc:         "Tests that seneschal rejects paths that contain symlinks or point to non-regular files/directories",
		Contacts:     []string{"chirantan@chromium.org", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		Vars:         []string{"keepState"},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Params: []testing.Param{
			// Parameters generated by params_test.go. DO NOT EDIT.
			{
				Name:              "stretch_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByArtifactStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "stretch_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByArtifactStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "stretch_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByArtifactStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "stretch_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByArtifactStretch(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByArtifactBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByArtifactBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByArtifactBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "buster_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByArtifactBuster(),
				Timeout:           7 * time.Minute,
			},
		},
	})
}

func ShareInvalidPaths(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	userData := filepath.Join("/home/user", pre.Container.VM.Concierge.GetOwnerID())
	downloads := filepath.Join(userData, "MyFiles/Downloads")
	sym := filepath.Join(downloads, "userdata")
	if err := os.Symlink(userData, sym); err != nil {
		s.Fatal("Failed to create symlink: ", err)
	}
	defer os.Remove(sym)

	if path, err := pre.Container.VM.ShareDownloadsPath(ctx, "userdata/Cookies", false); err == nil {
		s.Error("Unexpectedly shared path containing symlink")
		if err := pre.Container.VM.UnshareDownloadsPath(ctx, path); err != nil {
			s.Fatal("Failed to un-share path containing symlink: ", err)
		}
	} else if !strings.Contains(err.Error(), "symlink") {
		s.Error("Unexpected error when sharing a path containing a symlink: ", err)
	}

	devs := []int{syscall.S_IFBLK, syscall.S_IFIFO, syscall.S_IFCHR, syscall.S_IFSOCK}
	for _, dev := range devs {
		p := path.Join(downloads, fmt.Sprintf("dev_node%d", dev))
		if err := syscall.Mknod(p, 0o600, dev); err != nil {
			s.Fatal("Failed to create dev node: ", err)
		}
		defer os.Remove(p)

		if sharedPath, err := pre.Container.VM.ShareDownloadsPath(ctx, path.Base(p), false); err == nil {
			s.Error("Unexpectedly shared path to non-regular file")
			if err := pre.Container.VM.UnshareDownloadsPath(ctx, sharedPath); err != nil {
				s.Fatal("Failed to un-share device node: ", err)
			}
		} else if !strings.Contains(err.Error(), "non-regular") {
			s.Error("Unexpected error when sharing a path to a non-regular file: ", err)
		}
	}
}
