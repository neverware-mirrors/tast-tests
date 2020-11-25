// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"strings"
	"time"

	"chromiumos/tast/local/coords"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/shutil"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         DisplayDensity,
		Desc:         "Runs a crostini application from the terminal in high/low DPI modes and compares sizes",
		Contacts:     []string{"smbarber@chromium.org", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		Vars:         []string{"keepState"},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Params: []testing.Param{
			// Parameters generated by display_density_test.go. DO NOT EDIT.
			{
				Name:              "wayland_stretch_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_stretch_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_stretch_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_stretch_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_buster_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_buster_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_buster_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "wayland_buster_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.WaylandDemoConfig(),
			}, {
				Name:              "x11_stretch_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_stretch_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_stretch_amd64.tar.xz", "crostini_test_container_rootfs_stretch_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_stretch_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_stretch_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_stretch_arm.tar.xz", "crostini_test_container_rootfs_stretch_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentStretch(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_buster_amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_buster_amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_buster_arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			}, {
				Name:              "x11_buster_arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
				Val:               crostini.X11DemoConfig(),
			},
		},
	})
}

func DisplayDensity(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	tconn := pre.TestAPIConn
	cont := pre.Container
	keyboard := pre.Keyboard
	conf := s.Param().(crostini.DemoConfig)
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	type density int

	const (
		lowDensity density = iota
		highDensity
	)

	demoWindowSize := func(densityConfiguration density) (coords.Size, error) {
		windowName := conf.Name
		var subCommandArgs []string
		if densityConfiguration == lowDensity {
			windowName = windowName + "_low_density"
			// TODO(hollingum): Find a better way to pass environment vars to a container command (rather than invoking sh).
			subCommandArgs = append(subCommandArgs, "DISPLAY=${DISPLAY_LOW_DENSITY}", "WAYLAND_DISPLAY=${WAYLAND_DISPLAY_LOW_DENSITY}")
		}
		subCommandArgs = append(subCommandArgs, conf.AppPath, "--width=100", "--height=100", "--title="+windowName)

		cmd := cont.Command(ctx, "sh", "-c", strings.Join(subCommandArgs, " "))
		s.Logf("Running %q", shutil.EscapeSlice(cmd.Args))
		if err := cmd.Start(); err != nil {
			return coords.Size{}, err
		}
		defer cmd.Wait(testexec.DumpLogOnError)
		defer cmd.Kill()

		var sz coords.Size
		var err error
		if sz, err = crostini.PollWindowSize(ctx, tconn, windowName, 10*time.Second); err != nil {
			return coords.Size{}, err
		}
		s.Logf("Window %q size is %v", windowName, sz)

		s.Logf("Closing %q with keypress", windowName)
		err = keyboard.Accel(ctx, "Enter")

		return sz, err
	}

	sizeHighDensity, err := demoWindowSize(highDensity)
	if err != nil {
		s.Fatal("Failed getting high-density window size: ", err)
	}

	sizeLowDensity, err := demoWindowSize(lowDensity)
	if err != nil {
		s.Fatal("Failed getting low-density window size: ", err)
	}

	if err := crostini.VerifyWindowDensities(ctx, tconn, sizeHighDensity, sizeLowDensity); err != nil {
		s.Fatal("Failed during window density comparison: ", err)
	}
}
