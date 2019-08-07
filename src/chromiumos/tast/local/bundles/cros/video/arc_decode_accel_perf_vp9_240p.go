// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package video

import (
	"context"

	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/bundles/cros/video/decode"
	"chromiumos/tast/local/media/caps"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         ARCDecodeAccelPerfVP9240P,
		Desc:         "Runs arcvideodecoder_test on ARC++ to measure the performance with an 240p VP9 video test-25fps.vp9",
		Contacts:     []string{"johnylin@chromium.org", "chromeos-video-eng@google.com"},
		Attr:         []string{"group:crosbolt", "crosbolt_perbuild"},
		SoftwareDeps: []string{"android", "chrome", caps.HWDecodeVP9},
		Data:         []string{"test-25fps.vp9", "test-25fps.vp9.json"},
		Pre:          arc.Booted(),
	})
}

func ARCDecodeAccelPerfVP9240P(ctx context.Context, s *testing.State) {
	decode.RunARCVideoPerfTest(ctx, s, "test-25fps.vp9")
}
