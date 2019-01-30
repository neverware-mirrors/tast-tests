// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"

	"chromiumos/tast/local/bundles/cros/video/decode"
	"chromiumos/tast/local/bundles/cros/video/lib/caps"
	"chromiumos/tast/local/bundles/cros/video/lib/videotype"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     DecodeAccelVP8Import,
		Desc:     "Run Chrome video_decode_accelerator_unittest with a VP8 video in IMPORT mode",
		Contacts: []string{"keiichiw@chromium.org", "chromeos-video-eng@google.com"},
		Attr:     []string{"informational"},
		// VDA unittest cannot run with IMPORT mode on devices where ARC++ is disabled. (cf. crbug.com/881729)
		SoftwareDeps: []string{"android", caps.HWDecodeVP8},
		Data:         decode.DataFiles(videotype.VP8Prof, decode.ImportBuffer),
	})
}

// DecodeAccelVP8Import runs video_decode_accelerator_unittest in IMPORT mode with test-25fps.vp8.
func DecodeAccelVP8Import(ctx context.Context, s *testing.State) {
	decode.RunAllAccelVideoTest(ctx, s, decode.Test25FPSVP8, decode.ImportBuffer)
}
