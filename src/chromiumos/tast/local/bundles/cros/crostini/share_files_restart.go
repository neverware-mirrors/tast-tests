// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ui/filesapp"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/settings"
	"chromiumos/tast/local/crostini/ui/sharedfolders"
	"chromiumos/tast/local/crostini/ui/terminalapp"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     ShareFilesRestart,
		Desc:     "Test shared folders are persistent after restarting Crostini",
		Contacts: []string{"jinrongwu@google.com", "cros-containers-dev@google.com"},
		Attr:     []string{"group:mainline", "informational"},
		Vars:     []string{"keepState"},
		Params: []testing.Param{{
			Name:              "artifact",
			Pre:               crostini.StartedByArtifact(),
			ExtraData:         []string{crostini.ImageArtifact},
			Timeout:           7 * time.Minute,
			ExtraHardwareDeps: crostini.CrostiniStable,
		}, {
			Name:              "artifact_unstable",
			Pre:               crostini.StartedByArtifact(),
			ExtraData:         []string{crostini.ImageArtifact},
			Timeout:           7 * time.Minute,
			ExtraHardwareDeps: crostini.CrostiniUnstable,
		}, {
			Name:    "download_stretch",
			Pre:     crostini.StartedByDownloadStretch(),
			Timeout: 10 * time.Minute,
		}, {
			Name:    "download_buster",
			Pre:     crostini.StartedByDownloadBuster(),
			Timeout: 10 * time.Minute,
		}},
		SoftwareDeps: []string{"chrome", "vm_host"},
	})
}
func ShareFilesRestart(ctx context.Context, s *testing.State) {
	tconn := s.PreValue().(crostini.PreData).TestAPIConn
	cont := s.PreValue().(crostini.PreData).Container
	cr := s.PreValue().(crostini.PreData).Chrome
	keyboard := s.PreValue().(crostini.PreData).Keyboard

	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	// Open Files app.
	filesApp, err := filesapp.Launch(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to open Files app: ", err)
	}
	defer filesApp.Close(ctx)

	sharedFolders := sharedfolders.NewSharedFolders()
	// Clean up shared folders in the end.
	defer func() {
		if err := sharedFolders.UnshareAll(ctx, tconn, cont); err != nil {
			s.Error("Failed to unshare all folders: ", err)
		}
	}()

	if err := sharedFolders.ShareMyFilesOK(ctx, tconn, filesApp); err != nil {
		s.Fatal("Failed to share My files: ", err)
	}

	if err := checkResults(ctx, tconn, cont); err != nil {
		s.Fatal("Faied to verify results after sharing My files: ", err)
	}

	// Restart Crostini.
	terminalApp, err := terminalapp.Launch(ctx, tconn, strings.Split(cr.User(), "@")[0])
	if err != nil {
		s.Fatal("Failed to lauch terminal: ", err)
	}
	if err := terminalApp.RestartCrostini(ctx, keyboard, cont, cr.User()); err != nil {
		s.Fatal("Failed to restart crostini: ", err)
	}

	// Check the shared folders again after restart Crostini.
	if err := checkResults(ctx, tconn, cont); err != nil {
		s.Fatal("Faied to verify results after restarting Crostini: ", err)
	}
}

func checkResults(ctx context.Context, tconn *chrome.TestConn, cont *vm.Container) error {
	// Check shared folders on the Settings app.
	st, err := settings.OpenLinuxSettings(ctx, tconn, settings.ManageSharedFolders)
	if err != nil {
		return errors.Wrap(err, "failed to open Manage shared folders")
	}
	defer st.Close(ctx)

	shared, err := st.GetSharedFolders(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to find the shared folders list")
	}
	want := []string{sharedfolders.MyFiles}
	if diff := cmp.Diff(want, shared); diff != "" {
		return errors.Errorf("failed to verify shared folders list, got %s, want %s", shared, want)
	}

	// Check the file list in the container.
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		list, err := cont.GetFileList(ctx, sharedfolders.MountPath)
		if err != nil {
			return err
		}
		want := []string{"fonts", sharedfolders.MountFolderMyFiles}
		if diff := cmp.Diff(want, list); diff != "" {
			return errors.Errorf("failed to verify file list in /mnt/chromeos, got %s, want %s", list, want)
		}
		return nil
	}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to verify file list in container")
	}

	return nil
}
