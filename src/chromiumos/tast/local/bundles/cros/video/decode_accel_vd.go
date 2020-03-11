// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"

	"chromiumos/tast/local/bundles/cros/video/decode"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         DecodeAccelVD,
		Desc:         "Verifies hardware decode acceleration of media::VideoDecoders by running the video_decode_accelerator_tests binary (see go/vd-migration)",
		Contacts:     []string{"acourbot@chromium.org", "dstaessens@chromium.org", "chromeos-video-eng@google.com"},
		SoftwareDeps: []string{"chrome", "cros_video_decoder"},
		Params: []testing.Param{{
			Name:              "h264",
			Val:               "test-25fps.h264",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeH264},
			ExtraData:         []string{"test-25fps.h264", "test-25fps.h264.json"},
		}, {
			Name:              "vp8",
			Val:               "test-25fps.vp8",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeVP8},
			ExtraData:         []string{"test-25fps.vp8", "test-25fps.vp8.json"},
		}, {
			Name:              "vp9",
			Val:               "test-25fps.vp9",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9},
			ExtraData:         []string{"test-25fps.vp9", "test-25fps.vp9.json"},
		}, {
			Name: "vp9_2",
			Val:  "test-25fps.vp9_2",
			// TODO(crbug.com/911754): reenable this test once HDR VP9.2 is implemented.
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9_2},
			ExtraData:         []string{"test-25fps.vp9_2", "test-25fps.vp9_2.json"},
		}, {
			Name:              "h264_resolution_switch",
			Val:               "switch_1080p_720p_240frames.h264",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeH264},
			ExtraData:         []string{"switch_1080p_720p_240frames.h264", "switch_1080p_720p_240frames.h264.json"},
		}, {
			Name:              "vp8_resolution_switch",
			Val:               "resolution_change_500frames.vp8.ivf",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeVP8},
			ExtraData:         []string{"resolution_change_500frames.vp8.ivf", "resolution_change_500frames.vp8.ivf.json"},
		}, {
			Name:              "vp9_resolution_switch",
			Val:               "resolution_change_500frames.vp9.ivf",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9},
			ExtraData:         []string{"resolution_change_500frames.vp9.ivf", "resolution_change_500frames.vp9.ivf.json"},
		}, {
			// This test uses a video that makes use of the VP9 show-existing-frame feature and is used in Android CTS:
			// https://android.googlesource.com/platform/cts/+/master/tests/tests/media/res/raw/vp90_2_17_show_existing_frame.vp9
			Name:              "vp9_show_existing_frame",
			Val:               "vda_sanity-vp90_2_17_show_existing_frame.vp9",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9},
			ExtraData:         []string{"vda_sanity-vp90_2_17_show_existing_frame.vp9", "vda_sanity-vp90_2_17_show_existing_frame.vp9.json"},
		}, {
			// H264 stream in which a profile changes from Baseline to Main.
			Name:              "h264_profile_change",
			Val:               "test-25fps_basemain.h264",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeH264},
			ExtraData:         []string{"test-25fps_basemain.h264", "test-25fps_basemain.h264.json"},
		}, {
			// Run with HW decoder using VA-API only because only the HW decoder can decode SVC stream correctly today.
			// Decode VP9 spatial-SVC stream. Precisely the structure in the stream is called k-SVC, where spatial-layers are at key-frame only.
			// The structure is used in Hangouts Meet. go/vp9-svc-hangouts for detail.
			Name:              "vp9_keyframe_spatial_layers",
			Val:               "keyframe_spatial_layers_180p_360p.vp9.ivf",
			ExtraAttr:         []string{"group:mainline", "informational"},
			ExtraSoftwareDeps: []string{caps.HWDecodeVP9, "vaapi"},
			ExtraData:         []string{"keyframe_spatial_layers_180p_360p.vp9.ivf", "keyframe_spatial_layers_180p_360p.vp9.ivf.json"},
		}},
	})
}

func DecodeAccelVD(ctx context.Context, s *testing.State) {
	decode.RunAccelVideoTest(ctx, s, s.Param().(string), decode.VD)
}
