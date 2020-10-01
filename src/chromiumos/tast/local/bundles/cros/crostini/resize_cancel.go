// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"time"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/chrome/uig"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/settings"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         ResizeCancel,
		Desc:         "Test cancelling resizing of Crostini from the Settings app",
		Contacts:     []string{"jinrongwu@google.com", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Vars:         []string{"keepState"},
		Params: []testing.Param{
			// Parameters generated by params_test.go. DO NOT EDIT.
			{
				Name:              "artifact",
				ExtraData:         []string{"crostini_guest_images.tar"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByArtifact(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "artifact_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_guest_images.tar"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByArtifact(),
				Timeout:           7 * time.Minute,
			}, {
				Name:      "download_stretch",
				ExtraAttr: []string{"informational"},
				Pre:       crostini.StartedByDownloadStretch(),
				Timeout:   10 * time.Minute,
			}, {
				Name:      "download_buster",
				ExtraAttr: []string{"informational"},
				Pre:       crostini.StartedByDownloadBuster(),
				Timeout:   10 * time.Minute,
			},
		},
	})
}

func ResizeCancel(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	tconn := pre.TestAPIConn

	// Use a shortened context for test operations to reserve time for cleanup.
	cleanupCtx := ctx
	ctx, cancel := ctxutil.Shorten(ctx, 30*time.Second)
	defer cancel()
	defer crostini.RunCrostiniPostTest(cleanupCtx, pre)

	// Open the Linux settings.
	st, err := settings.OpenLinuxSettings(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to open Linux Settings: ", err)
	}
	defer st.Close(cleanupCtx)

	disk, err := pre.Container.VM.Concierge.GetVMDiskInfo(ctx, vm.DefaultVMName)
	if err != nil {
		s.Fatal("Failed to get VM disk info: ", err)
	}
	originalContDS := disk.GetSize()

	originalDSOnSettings, err := st.GetDiskSize(ctx)
	if err != nil {
		s.Fatal("Failed to get the disk size from the Settings app: ", err)
	}

	// Click Resize on Linux settings page.
	resizeDlg, err := st.ClickChange(ctx)
	if err != nil {
		s.Fatal("Failed to click button Change on Linux settings page: ", err)
	}

	// Get the dialog node and params.
	dialog, err := uig.GetNode(ctx, tconn, resizeDlg.Self)
	if err != nil {
		s.Fatal("Failed to get the node of the Resize Linux disk dialog")
	}
	dialogParams := ui.FindParams{
		Role: dialog.Role,
		Name: dialog.Name,
	}

	// Click cancel on the resize dialog.
	if err := uig.Do(ctx, tconn, uig.WaitForLocationChangeCompleted(), resizeDlg.Cancel.LeftClick()); err != nil {
		s.Fatal("Failed to click button Cancel on Resize Linux disk dialog: ", err)
	}

	// Wait the resize dialog gone.
	if err := ui.WaitUntilGone(ctx, tconn, dialogParams, 15*time.Second); err != nil {
		s.Fatal("Failed to close the Resize Linux disk dialog: ", err)
	}

	newDSOnSettings, err := st.GetDiskSize(ctx)
	if err != nil {
		s.Fatal("Failed to get the disk size from the Settings app after cancelling resizing: ", err)
	}

	disk, err = pre.Container.VM.Concierge.GetVMDiskInfo(ctx, vm.DefaultVMName)
	if err != nil {
		s.Fatal("Failed to get VM disk info after cancelling resizing: ", err)
	}

	newContDS := disk.GetSize()
	if originalContDS != newContDS {
		s.Fatalf("Failed to verify disk size of Crostini after cancelling resizing, got %d, want %d", newContDS, originalContDS)
	}
	if originalDSOnSettings != newDSOnSettings {
		s.Fatalf("Failed to verify disk size on the Settings app after cancelling resizing, got %s, want %s", newDSOnSettings, originalDSOnSettings)
	}
}
