// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package graphics

import (
	"context"
	"time"

	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/graphics/trace"
	"chromiumos/tast/local/graphics/trace/comm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         TraceReplayExtended,
		Desc:         "Repeatedly replay a 3D graphics trace file in Crostini VM for a fixed duration",
		Contacts:     []string{"chromeos-gfx@google.com", "ryanneph@google.com", "ddmail@google.com", "tutankhamen@google.com", "ihf@google.com"},
		SoftwareDeps: []string{"chrome", "crosvm_gpu", "vm_host"},
		Attr:         []string{},
		Vars:         []string{"keepState", "graphics.TraceReplayExtended.resultDir", "graphics.TraceReplayExtended.signalRunningFile", "graphics.TraceReplayExtended.signalCheckpointFile"},
		Params: []testing.Param{
			// Parameters generated by trace_replay_extended_test.go. DO NOT EDIT.
			{
				Name:              "glxgears_1minute_amd64",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           45 * time.Minute,
				Val: comm.TestGroupConfig{
					Labels: []string{"short"},
					Repository: comm.RepositoryInfo{
						RootURL: "gs://chromiumos-test-assets-public/tast/cros/graphics/traces/repo",
						Version: 1,
					},
					ExtendedDuration: 1 * 60,
				},
			}, {
				Name:              "glxgears_1minute_arm",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           45 * time.Minute,
				Val: comm.TestGroupConfig{
					Labels: []string{"short"},
					Repository: comm.RepositoryInfo{
						RootURL: "gs://chromiumos-test-assets-public/tast/cros/graphics/traces/repo",
						Version: 1,
					},
					ExtendedDuration: 1 * 60,
				},
			},
		},
	})
}

// TraceReplayExtended replays a graphics trace repeatedly inside a crostini container.
func TraceReplayExtended(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	config := s.Param().(comm.TestGroupConfig)
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))
	testVars := comm.TestVars{PowerTestVars: comm.GetPowerTestVars(s)}
	if err := trace.RunTraceReplayTest(ctx, s.OutDir(), s.CloudStorage(), pre.Container, &config, &testVars); err != nil {
		s.Fatal("Trace replay test failed: ", err)
	}
}
