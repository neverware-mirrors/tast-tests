// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"context"
	"time"

	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/bundles/cros/arc/c2e2etest"
	"chromiumos/tast/local/bundles/cros/arc/video"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/local/media/videotype"
	"chromiumos/tast/testing"
	"chromiumos/tast/testing/hwdep"
)

const (
	// Enable to cache the extracted raw video to speed up the test.
	veaCacheExtractedVideo = false
	// Enable to download the video file encoded by the test.
	veaPullEncodedVideo = false
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         VideoEncodeAccel,
		Desc:         "Verifies ARC++ and ARCVM hardware encode acceleration by running the arcvideoencoder_test binary",
		Contacts:     []string{"dstaessens@chromium.org", "chromeos-video-eng@google.com"},
		Data:         []string{c2e2etest.X86ApkName, c2e2etest.ArmApkName},
		SoftwareDeps: []string{"chrome", caps.HWEncodeH264},
		HardwareDeps: hwdep.D(hwdep.SkipOnPlatform(video.EncoderBlocklist...)),
		Pre:          arc.BootedWithVideoLogging(),
		// TODO(yusukes): Change the timeout back to 4 min when we revert arc.go's BootTimeout to 120s.
		Timeout: 5 * time.Minute,
		Params: []testing.Param{{
			Name: "h264_192p_i420",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Bear192P,
				PixelFormat: videotype.I420,
			},
			ExtraData:         []string{video.Bear192P.Name},
			ExtraSoftwareDeps: []string{"android_p"},
			ExtraAttr:         []string{"group:mainline", "informational"},
		}, {
			// TODO(b/140082257): enable once the virtio video encoder is ready
			Name: "h264_192p_i420_vm",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Bear192P,
				PixelFormat: videotype.I420,
			},
			ExtraData:         []string{video.Bear192P.Name},
			ExtraSoftwareDeps: []string{"android_vm"},
		}, {
			Name: "h264_360p_i420",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Tulip360P,
				PixelFormat: videotype.I420},
			ExtraData:         []string{video.Tulip360P.Name},
			ExtraSoftwareDeps: []string{"android_p"},
			ExtraAttr:         []string{"group:mainline", "informational"},
		}, {
			// TODO(b/140082257): enable once the virtio video encoder is ready
			Name: "h264_360p_i420_vm",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Tulip360P,
				PixelFormat: videotype.I420},
			ExtraData:         []string{video.Tulip360P.Name},
			ExtraSoftwareDeps: []string{"android_vm"},
		}, {
			Name: "h264_720p_i420",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Tulip720P,
				PixelFormat: videotype.I420},
			ExtraData:         []string{video.Tulip720P.Name},
			ExtraSoftwareDeps: []string{"android_p"},
			ExtraAttr:         []string{"group:mainline", "informational"},
		}, {
			// TODO(b/140082257): enable once the virtio video encoder is ready
			Name: "h264_720p_i420_vm",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Tulip720P,
				PixelFormat: videotype.I420},
			ExtraData:         []string{video.Tulip720P.Name},
			ExtraSoftwareDeps: []string{"android_vm"},
		}, {
			Name: "h264_1080p_i420",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Crowd1080P,
				PixelFormat: videotype.I420},
			ExtraData:         []string{video.Crowd1080P.Name},
			ExtraSoftwareDeps: []string{"android_p"},
			ExtraAttr:         []string{"group:mainline", "informational"},
		}, {
			// TODO(b/140082257): enable once the virtio video encoder is ready
			Name: "h264_1080p_i420_vm",
			Val: video.EncodeTestOptions{
				Profile:     videotype.H264Prof,
				Params:      video.Crowd1080P,
				PixelFormat: videotype.I420},
			ExtraData:         []string{video.Crowd1080P.Name},
			ExtraSoftwareDeps: []string{"android_vm"},
		}},
	})
}

func VideoEncodeAccel(ctx context.Context, s *testing.State) {
	video.RunARCVideoTest(ctx, s, s.PreValue().(arc.PreData).ARC,
		s.Param().(video.EncodeTestOptions), veaPullEncodedVideo, veaCacheExtractedVideo)
}
