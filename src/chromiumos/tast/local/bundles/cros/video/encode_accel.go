// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"
	"time"

	"chromiumos/tast/local/bundles/cros/video/encode"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/local/media/encoding"
	"chromiumos/tast/local/media/videotype"
	"chromiumos/tast/testing"
)

// Enable to cache the extracted raw video to speed up the test.
const ecCacheExtractedVideo = false

func init() {
	testing.AddTest(&testing.Test{
		Func:         EncodeAccel,
		Desc:         "Verifies hardware encode acceleration by running the video_encode_accelerator_unittest binary",
		Contacts:     []string{"akahuang@chromium.org", "johnylin@chromium.org", "hiroh@chromium.org", "chromeos-video-eng@google.com"},
		SoftwareDeps: []string{"chrome"},
		// TODO(crbug.com/979497): Reduce to appropriate timeout after checking the exact execution time of h264_2160p_i420.
		Timeout: 10 * time.Minute,
		Params: []testing.Param{{
			Name: "h264_180p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip180P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip180P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_192p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_360p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip360P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip360P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_720p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip720P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip720P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_1080p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Crowd1080P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Crowd1080P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_2160p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Crowd2160P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264_4K},
			ExtraData:         []string{encode.Crowd2160P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_192p_nv12",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeH264},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_192p_nv12_dmabuf",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.DMABuf},
			// TODO(hiroh): Remove "arc" deps once all boards support DMABUF-backed video frames.
			ExtraSoftwareDeps: []string{"arc", caps.HWEncodeH264},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_360p_nv12_dmabuf",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip360P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.DMABuf},
			// TODO(hiroh): Remove "arc" deps once all boards support DMABUF-backed video frames.
			ExtraSoftwareDeps: []string{"arc", caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip360P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_720p_nv12_dmabuf",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Tulip720P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.DMABuf},
			// TODO(hiroh): Remove "arc" deps once all boards support DMABUF-backed video frames.
			ExtraSoftwareDeps: []string{"arc", caps.HWEncodeH264},
			ExtraData:         []string{encode.Tulip720P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "h264_1080p_nv12_dmabuf",
			Val: encoding.TestOptions{
				Profile:     videotype.H264Prof,
				Params:      encode.Crowd1080P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.DMABuf},
			// TODO(hiroh): Remove "arc" deps once all boards support DMABUF-backed video frames.
			ExtraSoftwareDeps: []string{"arc", caps.HWEncodeH264},
			ExtraData:         []string{encode.Crowd1080P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_180p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Tulip180P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			ExtraData:         []string{encode.Tulip180P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_192p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_360p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Tulip360P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			ExtraData:         []string{encode.Tulip360P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_720p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Tulip720P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			ExtraData:         []string{encode.Tulip720P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_1080p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Crowd1080P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			ExtraData:         []string{encode.Crowd1080P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_2160p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Crowd2160P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8_4K},
			ExtraData:         []string{encode.Crowd2160P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_192p_nv12",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_192p_nv12_dmabuf",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.NV12,
				InputMode:   encoding.DMABuf},
			// Although the ability to android is unrelated to this test ability,
			// we would like to run this test on ARC++ enabled boards.
			// TODO(hiroh): Remove "arc" deps once Chrome VEAs and
			// Chrome OS supports DMABUF-backed video frame on all boards.
			ExtraSoftwareDeps: []string{"arc", caps.HWEncodeVP8},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp8_361p_i420_odd",
			Val: encoding.TestOptions{
				Profile:     videotype.VP8Prof,
				Params:      encode.Crowd361P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP8OddDimension},
			ExtraData:         []string{encode.Crowd361P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_180p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Tulip180P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9},
			ExtraData:         []string{encode.Tulip180P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_192p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Bear192P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9},
			ExtraData:         []string{encode.Bear192P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_360p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Tulip360P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9},
			ExtraData:         []string{encode.Tulip360P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_720p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Tulip720P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9},
			ExtraData:         []string{encode.Tulip720P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_1080p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Crowd1080P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9},
			ExtraData:         []string{encode.Crowd1080P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_2160p_i420",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Crowd2160P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9_4K},
			ExtraData:         []string{encode.Crowd2160P.Name},
			ExtraAttr:         []string{"group:graphics", "graphics_video", "graphics_nightly"},
		}, {
			Name: "vp9_361p_i420_odd",
			Val: encoding.TestOptions{
				Profile:     videotype.VP9Prof,
				Params:      encode.Crowd361P,
				PixelFormat: videotype.I420,
				InputMode:   encoding.SharedMemory},
			ExtraSoftwareDeps: []string{caps.HWEncodeVP9OddDimension},
			ExtraData:         []string{encode.Crowd361P.Name},
			// Disabled because the Intel encoder driver always aligns visible size by 16.
			// TODO(b/139846661): Enable once the Intel encoder driver issue is fixed.
		}},
	})
}

func EncodeAccel(ctx context.Context, s *testing.State) {
	encode.RunAllAccelVideoTests(ctx, s, s.Param().(encoding.TestOptions), ecCacheExtractedVideo)
}
