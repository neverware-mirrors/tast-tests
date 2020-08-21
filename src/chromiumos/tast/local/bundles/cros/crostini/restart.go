// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"strconv"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome/ui/faillog"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/terminalapp"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         Restart,
		Desc:         "Tests that we can shut down and restart crostini (where the VM image is a build artifact)",
		Contacts:     []string{"hollingum@chromium.org", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Vars:         []string{"crostini.Restart.numRestarts", "keepState"},
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

// varInt returns the value for the named variable, or defaultVal if it is
// not supplied or unparseable.
func varInt(s *testing.State, name string, defaultVal int) int {
	if str, ok := s.Var(name); ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			return val
		}
		s.Logf("Cannot parse argument %s %s: %v", name, str, err)
	}
	return defaultVal
}

func Restart(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	cont := pre.Container
	tconn := pre.TestAPIConn
	cr := pre.Chrome
	keyboard := pre.Keyboard
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	defer faillog.DumpUITreeOnError(ctx, s.OutDir(), s.HasError, tconn)

	numRestarts := varInt(s, "crostini.Restart.numRestarts", 2)

	startupTime, err := startTime(ctx, cont)
	if err != nil {
		s.Fatal("Failed to get startup time: ", err)
	}

	for i := 0; i < numRestarts; i++ {
		terminalApp, err := terminalapp.Launch(ctx, tconn)
		if err != nil {
			s.Fatal("Failed to lauch terminal: ", err)
		}

		s.Logf("Restart #%d, startup time was %v", i+1, startupTime)
		if err := terminalApp.RestartCrostini(ctx, keyboard, cont, cr.User()); err != nil {
			s.Fatal("Failed to restart crostini: ", err)
		}

		// Compare start times.
		newStartupTime, err := startTime(ctx, cont)
		if err != nil {
			s.Fatal("Failed to get new startup time: ", err)
		}
		if !newStartupTime.After(startupTime) {
			s.Errorf("Restarted container didnt have a later startup time, %v vs %v", startupTime, newStartupTime)
		}
		startupTime = newStartupTime
	}
}

func startTime(ctx context.Context, cont *vm.Container) (time.Time, error) {
	out, err := cont.Command(ctx, "uptime", "--since").Output(testexec.DumpLogOnError)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to run uptime cmd")
	}
	t, err := time.Parse("2006-01-02 15:04:05\n", string(out))
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to parse uptime")
	}
	return t, nil
}
