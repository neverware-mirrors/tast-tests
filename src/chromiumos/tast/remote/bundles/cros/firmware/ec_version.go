// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"

	"chromiumos/tast/remote/firmware"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     ECVersion,
		Desc:     "Verify that the EC version can be retrieved from ectool",
		Contacts: []string{"cros-fw-engprod@google.com"},
		Attr:     []string{"group:mainline", "informational"},
	})
}

func ECVersion(ctx context.Context, s *testing.State) {
	version, err := firmware.ECVersion(ctx, s.DUT())
	if err != nil {
		s.Fatal("Failed to determine ec version: ", err)
	}
	s.Log("EC version: ", version)
}