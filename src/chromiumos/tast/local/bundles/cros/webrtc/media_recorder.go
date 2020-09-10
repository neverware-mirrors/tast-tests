// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webrtc

import (
	"context"
	"time"

	"chromiumos/tast/local/bundles/cros/webrtc/mediarecorder"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/local/media/pre"
	"chromiumos/tast/local/media/videotype"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: MediaRecorder,
		Desc: "Verifies that MediaRecorder uses video encode acceleration",
		Contacts: []string{
			"mcasas@chromium.org",
			"chromeos-gfx-video@google.com",
			"chromeos-video-eng@google.com",
		},

		SoftwareDeps: []string{"chrome"},
		Data:         []string{"loopback_media_recorder.html"},
		Params: []testing.Param{{
			Name:              "h264",
			Val:               videotype.H264,
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_perbuild"},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264, "proprietary_codecs"},
			Pre:               pre.ChromeVideoWithFakeWebcam(),
		}, {
			Name:              "vp8",
			Val:               videotype.VP8,
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_perbuild"},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			Pre:               pre.ChromeVideoWithFakeWebcam(),
		}, {
			Name:              "vp9",
			Val:               videotype.VP9,
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_perbuild"},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9},
			Pre:               pre.ChromeVideoWithFakeWebcam(),
		}, {
			Name:              "vp8_cam",
			Val:               videotype.VP8,
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
			ExtraSoftwareDeps: []string{caps.BuiltinCamera, caps.HWEncodeVP8},
			Pre:               pre.ChromeCameraPerf(),
		}},
	})
}

// MediaRecorder verifies that a video encode accelerator was used.
func MediaRecorder(ctx context.Context, s *testing.State) {
	const (
		// Let the MediaRecorder accumulate a few milliseconds, otherwise we might
		// receive just bits and pieces of the container header.
		recordDuration = 100 * time.Millisecond
	)

	if err := mediarecorder.VerifyMediaRecorderUsesEncodeAccelerator(ctx, s.PreValue().(*chrome.Chrome), s.DataFileSystem(), s.Param().(videotype.Codec), recordDuration); err != nil {
		s.Error("Failed to run VerifyMediaRecorderUsesEncodeAccelerator: ", err)
	}
}
