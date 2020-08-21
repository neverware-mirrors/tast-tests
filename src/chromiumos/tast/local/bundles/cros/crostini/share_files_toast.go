// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"time"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/bundles/cros/crostini/listset"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ui/filesapp"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/settings"
	"chromiumos/tast/local/crostini/ui/sharedfolders"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         ShareFilesToast,
		Desc:         "Test sharing My files with Crostini and clicking Manage on toast nofication",
		Contacts:     []string{"jinrongwu@google.com", "cros-containers-dev@google.com"},
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
func ShareFilesToast(ctx context.Context, s *testing.State) {
	tconn := s.PreValue().(crostini.PreData).TestAPIConn
	cont := s.PreValue().(crostini.PreData).Container
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	// Use a shortened context for test operations to reserve time for cleanup.
	cleanupCtx := ctx
	ctx, cancel := ctxutil.Shorten(ctx, 30*time.Second)
	defer cancel()
	defer crostini.RunCrostiniPostTest(cleanupCtx, s.PreValue().(crostini.PreData))

	// Open Files app.
	filesApp, err := filesapp.Launch(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to open Files app: ", err)
	}
	defer filesApp.Close(cleanupCtx)

	sharedFolders := sharedfolders.NewSharedFolders()
	// Clean up shared folders in the end.
	defer func() {
		if err := sharedFolders.UnshareAll(cleanupCtx, tconn, cont); err != nil {
			s.Error("Failed to unshare all folders: ", err)
		}
	}()

	if err := shareMyFilesOKManage(ctx, sharedFolders, filesApp, tconn); err != nil {
		s.Fatal("Failed to share My files: ", err)
	}

	if err := checkShareResults(ctx, tconn, cont); err != nil {
		s.Fatal("Failed to check result after sharing My files: ", err)
	}

	// Unshare My files. This is part of the test, different from clean up in line 72.
	if err := unshareMyFiles(ctx, tconn, cont, sharedFolders); err != nil {
		s.Fatal("Failed to unshare My files: ", err)
	}
}

func shareMyFilesOKManage(ctx context.Context, sharedFolders *sharedfolders.SharedFolders, filesApp *filesapp.FilesApp, tconn *chrome.TestConn) error {
	// Share My files, click OK on the confirm dialog.
	if err := sharedFolders.ShareMyFilesOK(ctx, tconn, filesApp); err != nil {
		return errors.Wrap(err, "failed to share My files")
	}
	toast, err := sharedfolders.FindToast(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to find toast after sharing My files")
	}
	defer toast.Release(ctx)

	// Click button Manage.
	if err := toast.ClickManage(ctx, tconn); err != nil {
		return errors.Wrap(err, "failed to click button Manage on toast notification")
	}

	return nil
}

func checkShareResults(ctx context.Context, tconn *chrome.TestConn, cont *vm.Container) error {
	// Check the shared folders on Settings.
	s, err := settings.FindSettingsPage(ctx, tconn, settings.PageNameMSF)
	if err != nil {
		return errors.Wrap(err, "failed to find Manage shared folders")
	}
	defer s.Close(ctx)

	sharedFoldersList, err := s.GetSharedFolders(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to find the shared folders list")
	}
	if err := listset.CheckListsMatch(sharedFoldersList, sharedfolders.MyFiles); err != nil {
		return errors.Wrap(err, "failed to verify shared folders list")
	}

	// Check the file list in the container.
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		list, err := cont.GetFileList(ctx, sharedfolders.MountPath)
		if err != nil {
			return err
		}
		if err := listset.CheckListsMatch(list, "fonts", sharedfolders.MountFolderMyFiles); err != nil {
			return err
		}
		return nil
	}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to verify file list in container")
	}

	return nil
}

func unshareMyFiles(ctx context.Context, tconn *chrome.TestConn, cont *vm.Container, sharedFolders *sharedfolders.SharedFolders) error {
	if err := sharedFolders.Unshare(ctx, tconn, sharedfolders.MyFiles); err != nil {
		return errors.Wrap(err, "failed to delete shared folder My files")
	}

	if err := sharedFolders.CheckNoSharedFolders(ctx, tconn, cont); err != nil {
		return errors.Wrap(err, "failed to verify shared folder list after unshare My files")
	}

	return nil
}
