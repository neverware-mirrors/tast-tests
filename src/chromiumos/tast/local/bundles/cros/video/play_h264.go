// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"

	"chromiumos/tast/local/bundles/cros/video/play"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: PlayH264,
		Desc: "Checks H264 video playback is working",
		Attr: []string{"informational"},
		// "chrome_internal" is needed because H.264 is a proprietary codec.
		SoftwareDeps: []string{"chrome_login", "chrome_internal"},
		Data:         []string{"bear_h264_320x180.mp4", "video.html"},
	})
}

// PlayH264 plays bear_h264_320x180.mp4 with Chrome.
func PlayH264(ctx context.Context, s *testing.State) {
	play.TestPlay(ctx, s, "bear_h264_320x180.mp4", play.NoCheckHistogram)
}
