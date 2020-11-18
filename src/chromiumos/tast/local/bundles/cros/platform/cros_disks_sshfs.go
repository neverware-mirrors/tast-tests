// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package platform

import (
	"context"

	"chromiumos/tast/local/bundles/cros/platform/crosdisks"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: CrosDisksSSHFS,
		Desc: "Verifies CrosDisks mounts a network server over ssh",
		Contacts: []string{
			"dats@chromium.org",
			"joelhockey@chromium.org",
		},
		Attr: []string{"group:mainline", "informational"},
	})
}

func CrosDisksSSHFS(ctx context.Context, s *testing.State) {
	crosdisks.RunSSHFSTests(ctx, s)
}
