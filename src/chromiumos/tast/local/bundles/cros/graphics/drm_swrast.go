// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package graphics

import (
	"context"
	"time"

	"chromiumos/tast/local/bundles/cros/graphics/drm"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: DRMSwrast,
		Desc: "Verifies DRM swrast_test runs successfully",
		Contacts: []string{
			"andrescj@chromium.org",
			"chromeos-gfx@google.com",
			"hidehiko@chromium.org", // Tast port.
		},
		Attr: []string{"informational"},
	})
}

func DRMSwrast(ctx context.Context, s *testing.State) {
	if err := drm.SetUp(ctx); err != nil {
		s.Fatal("Failed to set up the DRM test: ", err)
	}
	defer drm.TearDown(ctx)

	drm.RunTest(ctx, s, 20*time.Second, "/usr/local/bin/swrast_test")
}
