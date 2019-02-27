// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"

	"chromiumos/tast/local/bundles/cros/video/encode"
	"chromiumos/tast/local/bundles/cros/video/lib/caps"
	"chromiumos/tast/local/bundles/cros/video/lib/videotype"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         EncodeAccelVP9720PI420,
		Desc:         "Runs Chrome video_encode_accelerator_unittest from 720p I420 raw frames to VP9 stream",
		Contacts:     []string{"hiroh@chromium.org", "chromeos-video-eng@google.com"},
		Attr:         []string{"informational"},
		SoftwareDeps: []string{caps.HWEncodeVP9},
		Data:         []string{encode.Tulip720P.Name},
	})
}

func EncodeAccelVP9720PI420(ctx context.Context, s *testing.State) {
	encode.RunAllAccelVideoTests(ctx, s, encode.TestOptions{
		Profile:     videotype.VP9Prof,
		Params:      encode.Tulip720P,
		PixelFormat: videotype.I420,
		InputMode:   encode.SharedMemory})
}
