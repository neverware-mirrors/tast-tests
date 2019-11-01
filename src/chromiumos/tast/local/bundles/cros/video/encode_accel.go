// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"
	"time"

	"chromiumos/tast/local/bundles/cros/video/encode"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/local/media/videotype"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         EncodeAccel,
		Desc:         "Verifies hardware encode acceleration by running the video_encode_accelerator_unittest binary",
		Contacts:     []string{"akahuang@chromium.org", "johnylin@chromium.org", "chromeos-video-eng@google.com"},
		Attr:         []string{"group:mainline"},
		SoftwareDeps: []string{"chrome"},
		// TODO(crbug.com/979497): Reduce to appropriate timeout after checking the exact execution time of h264_2160p_i420.
		Timeout: 10 * time.Minute,
		Params: []testing.Param{{
			Name: "h264_180p_i420",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip180P,
				PixelFormat: videotype.I420,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip180P.Name},
			// Marked informational due to failures on ToT.
			// TODO(crbug.com/1009297): Promote to critical again.
			ExtraAttr: []string{"informational"},
		}, {
			Name: "h264_192p_i420",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.I420,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Bear192P.Name},
			// Marked informational due to failures on ToT.
			// TODO(crbug.com/1009297): Promote to critical again.
			ExtraAttr: []string{"informational"},
		}, {
			Name: "h264_360p_i420",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip360P,
				PixelFormat: videotype.I420,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip360P.Name},
			// Marked informational due to failures on ToT.
			// TODO(crbug.com/1009297): Promote to critical again.
			ExtraAttr: []string{"informational"},
		}, {
			Name: "h264_720p_i420",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip720P,
				PixelFormat: videotype.I420,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip720P.Name},
			ExtraAttr:         []string{"informational"},
		}, {
			Name: "h264_1080p_i420",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Crowd1080P,
				PixelFormat: videotype.I420,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Crowd1080P.Name},
			ExtraAttr:         []string{"informational"},
		}, {
			Name: "h264_2160p_i420",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Crowd2160P,
				PixelFormat: videotype.I420,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264_4K},
			ExtraData:         []string{encode.Crowd2160P.Name},
			ExtraAttr:         []string{"informational"},
		}, {
			Name: "h264_192p_nv12",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.NV12,
				InputMode:   encode.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"informational"},
		}, {
			Name: "h264_192p_nv12_dmabuf",
			Val: encode.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.NV12,
				InputMode:   encode.DMABuf},
			// Although the ability to android is unrelated to this test ability,
			// we would like to run this test on ARC++ enabled boards.
			// TODO(hiroh): Remove "android" deps once Chrome VEAs and Chrome OS
			// supports DMABUF-backed video frame on all boards.
			ExtraSoftwareDeps: []string{"android", caps.HWEncodeH264},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"informational"},
		}},
	})
}

func EncodeAccel(ctx context.Context, s *testing.State) {
	encode.RunAllAccelVideoTests(ctx, s, s.Param().(encode.TestOptions))
}
