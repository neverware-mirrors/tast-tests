// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chromiumos/tast/common/perf"
	"chromiumos/tast/ctxutil"
	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/bundles/cros/arc/c2e2etest"
	"chromiumos/tast/local/bundles/cros/arc/video"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/local/media/encoding"
	"chromiumos/tast/local/media/videotype"
	"chromiumos/tast/local/power"
	"chromiumos/tast/local/power/setup"
	"chromiumos/tast/testing"
	"chromiumos/tast/testing/hwdep"
)

const (
	// Enable to cache the extracted raw video to speed up the test.
	pvepCacheExtractedVideo = false
	pvepIterationCount      = 30
	pvepIterationDuration   = 10 * time.Second
	pvepWarmupDuration      = 10 * time.Second
	pvepTestSlack           = 5 * time.Minute
	pvepPowerTestDuration   = pvepIterationCount*pvepIterationDuration + pvepWarmupDuration + pvepTestSlack
)

func init() {
	testing.AddTest(&testing.Test{
		Func: PowerVideoEncodePerf,
		Desc: "Measures the battery drain during hardware accelerated video encoding",
		Contacts: []string{
			"dstaessens@chromium.org",
			"arcvm-eng@google.com",
		},
		SoftwareDeps: []string{"chrome", caps.HWEncodeH264},
		HardwareDeps: hwdep.D(hwdep.SkipOnPlatform(video.EncoderBlocklist...)),
		Data:         []string{c2e2etest.X86ApkName, c2e2etest.ArmApkName},
		Pre:          arc.Booted(),
		Timeout:      pvepPowerTestDuration,
		Params: []testing.Param{{
			Name: "h264_1080p_i420",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Crowd1080P,
				PixelFormat: videotype.I420,
			},
			ExtraData:         []string{video.Crowd1080P.Name},
			ExtraSoftwareDeps: []string{"android_p"},
			ExtraAttr:         []string{"group:crosbolt", "crosbolt_perbuild"},
		}, {
			Name: "h264_1080p_i420_vm",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Crowd1080P,
				PixelFormat: videotype.I420,
			},
			ExtraData:         []string{video.Crowd1080P.Name},
			ExtraSoftwareDeps: []string{"android_vm"},
		}},
	})
}

func PowerVideoEncodePerf(ctx context.Context, s *testing.State) {
	const (
		// arcFilePath must be on the sdcard because of android permissions
		arcFilePath = "/sdcard/Download/c2_e2e_test/"
		logFileName = "gtest_logs.txt"
	)

	// Give cleanup actions a minute to run, even if we fail by exceeding our deadline.
	cleanupCtx := ctx
	ctx, cancel := ctxutil.Shorten(ctx, time.Minute)
	defer cancel()

	cr := s.PreValue().(arc.PreData).Chrome
	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create Test API connection: ", err)
	}

	a := s.PreValue().(arc.PreData).ARC
	opts := s.Param().(video.EncodeTestOptions)

	// Only H.264 is currently supported.
	if opts.Profile != videotype.H264Prof {
		s.Fatalf("Profile (%d) is not supported", opts.Profile)
	}

	// Extract video to create the raw video stream that will be encoded.
	rawVideoPath, err := encoding.PrepareYUV(ctx, s.DataPath(opts.Params.Name), opts.PixelFormat, opts.Params.Size)
	if err != nil {
		s.Fatal("Failed to prepare YUV file: ", err)
	}
	if !pvepCacheExtractedVideo {
		defer os.Remove(rawVideoPath)
	}

	// Set up device for measuring power drain.
	sup, cleanup := setup.New("video power")
	defer func() {
		if err := cleanup(cleanupCtx); err != nil {
			s.Error("Cleanup failed: ", err)
		}
	}()
	sup.Add(setup.PowerTest(ctx, tconn, setup.PowerTestOptions{
		Wifi: setup.DisableWifiInterfaces, Battery: setup.ForceBatteryDischarge}))

	// Push raw video file to ARC.
	sup.Add(setup.AdbMkdir(ctx, a, arcFilePath))
	if err := a.PushFile(ctx, rawVideoPath, arcFilePath); err != nil {
		s.Fatal("Failed to push raw video file to ARC: ", err)
	}

	// Install and prepare video encoder test.
	apkName, err := c2e2etest.ApkNameForArch(ctx, a)
	if err != nil {
		s.Fatal("Failed to get apk: ", err)
	}

	sup.Add(setup.InstallApp(ctx, a, s.DataPath(apkName), c2e2etest.Pkg))
	for _, p := range c2e2etest.RequiredPermissions() {
		sup.Add(setup.GrantAndroidPermission(ctx, a, c2e2etest.Pkg, p))
	}

	// Wait until CPU is cooled down.
	if _, err := power.WaitUntilCPUCoolDown(ctx, power.CoolDownPreserveUI); err != nil {
		s.Fatal("CPU failed to cool down: ", err)
	}

	// Run video encoder test in loop.
	encodeOutFile := filepath.Join(arcFilePath, strings.TrimSuffix(opts.Params.Name, ".vp9.webm")+".h264")
	streamDataArgs := encoding.CreateStreamDataArg(opts.Params, opts.Profile,
		opts.PixelFormat, arcFilePath+"/"+filepath.Base(rawVideoPath), encodeOutFile)
	testArgs := []string{
		streamDataArgs,
		"--run_at_fps",
		"--num_encoded_frames=1000000",
		"--gtest_filter=C2VideoEncoderE2ETest.TestSimpleEncode",
	}
	if opts.EncoderType == video.SoftwareEncoder {
		testArgs = append(testArgs, "--use_sw_encoder")
	}
	intentExtras := []string{
		"--ez", "do-encode", "true",
		"--esa", "test-args", strings.Join(testArgs, ","),
		"--es", "log-file", filepath.Join(arcFilePath, logFileName)}

	sup.Add(setup.StartActivity(ctx, tconn, a, c2e2etest.Pkg, c2e2etest.ActivityName,
		setup.Prefixes("-W", "-n"), setup.Suffixes(intentExtras...)))
	if err := sup.Check(ctx); err != nil {
		s.Fatal("Setup failed: ", err)
	}

	// Setup test metrics.
	metrics, err := perf.NewTimeline(ctx, power.TestMetrics(), perf.Interval(pvepIterationDuration))
	if err != nil {
		s.Fatal("Failed to build metrics: ", err)
	}
	s.Log("Finished setup")

	if err := metrics.Start(ctx); err != nil {
		s.Fatal("Failed to start metrics: ", err)
	}

	s.Log("Warmup: waiting a bit before starting the measurement")
	if err := testing.Sleep(ctx, pvepWarmupDuration); err != nil {
		s.Fatal("Failed to sleep: ", err)
	}

	// Measure the power drain during the specified interval.
	s.Log("Starting measurement")
	if err := metrics.StartRecording(ctx); err != nil {
		s.Fatal("Failed to start recording: ", err)
	}

	sleepDuration := pvepIterationCount * pvepIterationDuration
	s.Logf("Sleeping for %v while measuring power drain", sleepDuration)
	if err := testing.Sleep(ctx, sleepDuration); err != nil {
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
