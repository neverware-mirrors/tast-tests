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
		Func:     DecodeAccelVDVP9ResolutionSwitch,
		Desc:     "Runs Chrome video_decode_accelerator_tests with a VP9 resolution switching video on a media::VideoDecoder (see go/vd-migration)",
		Contacts: []string{"dstaessens@chromium.org", "chromeos-video-eng@google.com"},
		Attr:     []string{"group:mainline", "informational"},
		// TODO(b/137916185): Remove dependency on android capability. It's used here
		// to guarantee import-mode support, which is required by the new VD's.
		SoftwareDeps: []string{"android", "chrome", caps.HWDecodeVP9},
		Data:         []string{"resolution_change_500frames.vp9.ivf", "resolution_change_500frames.vp9.ivf.json"},
	})
}

func DecodeAccelVDVP9ResolutionSwitch(ctx context.Context, s *testing.State) {
	decode.RunAccelVideoTest(ctx, s, "resolution_change_500frames.vp9.ivf", decode.VD)
}
