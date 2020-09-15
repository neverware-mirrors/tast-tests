// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"chromiumos/tast/common/perf"
	"chromiumos/tast/ctxutil"
	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/bundles/cros/arc/c2e2etest"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/local/power"
	"chromiumos/tast/local/power/setup"
	"chromiumos/tast/testing"
)

const (
	// arcFilePath must be on the sdcard because of android permissions
	arcFilePath = "/sdcard/Download/c2_e2e_test/"

	iterationCount    = 30
	iterationDuration = 10 * time.Second
	warmupDuration    = 10 * time.Second
	// Mostly consumed by boot timeout and WaitUntilIdle
	testSlack         = 5 * time.Minute
	powerTestDuration = iterationCount*iterationDuration + warmupDuration + testSlack

	logFileName = "gtest_logs.txt"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: PowerVideoDecodePerf,
		Desc: "Measures the battery drain during hardware accelerated video playback",
		Contacts: []string{
			"stevensd@chromium.org",
			"arcvm-eng@google.com",
		},
		SoftwareDeps: []string{"chrome"},
		Data:         []string{c2e2etest.X86ApkName, c2e2etest.ArmApkName},
		Pre:          arc.BootedWithDisableSyncFlags(),
		Attr:         []string{"group:crosbolt", "crosbolt_nightly"},
		Timeout:      powerTestDuration,
		Params: []testing.Param{{
			Name:              "h264_1080p_30fps",
			Val:               "1080p_30fps_300frames.h264",
			ExtraSoftwareDeps: []string{caps.HWDecodeH264, "android_p"},
			ExtraData:         []string{"1080p_30fps_300frames.h264", "1080p_30fps_300frames.h264.json"},
		}, {
			Name:              "h264_1080p_30fps_vm",
			Val:               "1080p_30fps_300frames.h264",
			ExtraSoftwareDeps: []string{caps.HWDecodeH264, "android_vm"},
			ExtraData:         []string{"1080p_30fps_300frames.h264", "1080p_30fps_300frames.h264.json"},
		}, {
			Name:              "vp8_1080p_30fps",
			Val:               "1080p_30fps_300frames.vp8.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP8, "android_p"},
			ExtraData:         []string{"1080p_30fps_300frames.vp8.ivf", "1080p_30fps_300frames.vp8.ivf.json"},
		}, {
			Name:              "vp8_1080p_30fps_vm",
			Val:               "1080p_30fps_300frames.vp8.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP8, "android_vm"},
			ExtraData:         []string{"1080p_30fps_300frames.vp8.ivf", "1080p_30fps_300frames.vp8.ivf.json"},
		}, {
			Name:              "vp9_1080p_30fps",
			Val:               "1080p_30fps_300frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9, "android_p"},
			ExtraData:         []string{"1080p_30fps_300frames.vp9.ivf", "1080p_30fps_300frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_1080p_30fps_vm",
			Val:               "1080p_30fps_300frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9, "android_vm"},
			ExtraData:         []string{"1080p_30fps_300frames.vp9.ivf", "1080p_30fps_300frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_1080p_60fps",
			Val:               "1080p_60fps_600frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_60, "android_p"},
			ExtraData:         []string{"1080p_60fps_600frames.vp9.ivf", "1080p_60fps_600frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_1080p_60fps_vm",
			Val:               "1080p_60fps_600frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_60, "android_vm"},
			ExtraData:         []string{"1080p_60fps_600frames.vp9.ivf", "1080p_60fps_600frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_2160p_30fps",
			Val:               "2160p_30fps_300frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_4K, "android_p"},
			ExtraData:         []string{"2160p_30fps_300frames.vp9.ivf", "2160p_30fps_300frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_2160p_30fps_vm",
			Val:               "2160p_30fps_300frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_4K, "android_vm"},
			ExtraData:         []string{"2160p_30fps_300frames.vp9.ivf", "2160p_30fps_300frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_2160p_60fps",
			Val:               "2160p_60fps_600frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_4K60, "android_p"},
			ExtraData:         []string{"2160p_60fps_600frames.vp9.ivf", "2160p_60fps_600frames.vp9.ivf.json"},
		}, {
			Name:              "vp9_2160p_60fps_vm",
			Val:               "2160p_60fps_600frames.vp9.ivf",
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_4K60, "android_vm"},
			ExtraData:         []string{"2160p_60fps_600frames.vp9.ivf", "2160p_60fps_600frames.vp9.ivf.json"},
		}},
	})
}

func PowerVideoDecodePerf(ctx context.Context, s *testing.State) {
	// Give cleanup actions a minute to run, even if we fail by exceeding our
	// deadline.
	cleanupCtx := ctx
	ctx, cancel := ctxutil.Shorten(ctx, time.Minute)
	defer cancel()

	cr := s.PreValue().(arc.PreData).Chrome

	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create Test API connection: ", err)
	}

	a := s.PreValue().(arc.PreData).ARC
	testVideoFile := s.Param().(string)

	// Parse JSON metadata.
	md, err := c2e2etest.LoadMetadata(s.DataPath(testVideoFile) + ".json")
	if err != nil {
		s.Fatal("Failed to get metadata: ", err)
	}

	apkName, err := c2e2etest.ApkNameForArch(ctx, a)
	if err != nil {
		s.Fatal("Failed to get apk: ", err)
	}

	testVideoDataArg, err := md.StreamDataArg(filepath.Join(arcFilePath, testVideoFile))
	if err != nil {
		s.Fatal("Failed to construct --test_video_data: ", err)
	}

	testArgs := []string{
		testVideoDataArg,
		"--loop",
		"--gtest_filter=C2VideoDecoderSurfaceE2ETest.TestFPS",
	}
	intentExtras := []string{
		"--esa", "test-args", strings.Join(testArgs, ","),
		"--es", "log-file", filepath.Join(arcFilePath, logFileName)}

	sup, cleanup := setup.New("video power")
	defer func() {
		if err := cleanup(cleanupCtx); err != nil {
			s.Error("Cleanup failed: ", err)
		}
	}()

	sup.Add(setup.PowerTest(ctx, tconn, setup.ForceBatteryDischarge))
	sup.Add(setup.InstallApp(ctx, a, s.DataPath(apkName), c2e2etest.Pkg))
	for _, p := range c2e2etest.RequiredPermissions() {
		sup.Add(setup.GrantAndroidPermission(ctx, a, c2e2etest.Pkg, p))
	}

	sup.Add(setup.AdbMkdir(ctx, a, arcFilePath))
	if err := a.PushFile(ctx, s.DataPath(testVideoFile), arcFilePath); err != nil {
		s.Fatal("Failed to push video stream to ARC: ", err)
	}

	// Wait until CPU is cooled down.
	if _, err := power.WaitUntilCPUCoolDown(ctx, power.CoolDownPreserveUI); err != nil {
		s.Fatal("CPU failed to cool down: ", err)
	}

	sup.Add(setup.StartActivity(ctx, tconn, a, c2e2etest.Pkg, c2e2etest.ActivityName, setup.Prefixes("-n"), setup.Suffixes(intentExtras...)))

	if err := sup.Check(ctx); err != nil {
		s.Fatal("Setup failed: ", err)
	}

	metrics, err := perf.NewTimeline(ctx, power.TestMetrics(), perf.Interval(iterationDuration))
	if err != nil {
		s.Fatal("Failed to build metrics: ", err)
	}
	s.Log("Finished setup")

	if err := metrics.Start(ctx); err != nil {
		s.Fatal("Failed to start metrics: ", err)
	}

	s.Log("Warmup: waiting a bit before starting the measurement")
	if err := testing.Sleep(ctx, warmupDuration); err != nil {
		s.Fatal("Failed to sleep: ", err)
	}

	s.Log("Starting measurement")
	if err := metrics.StartRecording(ctx); err != nil {
		s.Fatal("Failed to start recording: ", err)
	}

	if err := testing.Sleep(ctx, iterationCount*iterationDuration); err != nil {
		s.Fatal("Failed to sleep: ", err)
	}

	p, err := metrics.StopRecording()
	if err != nil {
		s.Fatal("Error while recording power metrics: ", err)
	}

	if err := p.Save(s.OutDir()); err != nil {
		s.Error("Failed saving perf data: ", err)
	}
}
