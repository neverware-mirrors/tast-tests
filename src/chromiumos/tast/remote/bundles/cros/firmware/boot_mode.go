// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"

	fwCommon "chromiumos/tast/common/firmware"
	"chromiumos/tast/remote/firmware"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         BootMode,
		Desc:         "Verifies that remote tests can boot the DUT into, and confirm that the DUT is in, the different firmware modes (normal, dev, and recovery)",
		Contacts:     []string{"cros-fw-engprod@google.com"},
		Data:         firmware.ConfigDatafiles(),
		ServiceDeps:  []string{"tast.cros.firmware.UtilsService"},
		SoftwareDeps: []string{"crossystem"},
		Attr:         []string{"group:mainline", "informational"},
		Params: []testing.Param{{
			Name: "normal",
			Val:  []fwCommon.BootMode{fwCommon.BootModeNormal, fwCommon.BootModeNormal},
		}, {
			Name: "rec",
			Val:  []fwCommon.BootMode{fwCommon.BootModeNormal, fwCommon.BootModeRecovery, fwCommon.BootModeNormal},
		}, {
			Name: "dev",
			Val:  []fwCommon.BootMode{fwCommon.BootModeNormal, fwCommon.BootModeDev, fwCommon.BootModeNormal},
		}},
		Vars: []string{"servo"},
	})
}

func BootMode(ctx context.Context, s *testing.State) {
	modes := s.Param().([]fwCommon.BootMode)
	h := firmware.NewHelper(s.DUT(), s.RPCHint(), s.DataPath(firmware.ConfigDir), s.RequiredVar("servo"))
	defer h.Close(ctx)
	ms, err := firmware.NewModeSwitcher(ctx, h)
	if err != nil {
		s.Fatal("creating mode switcher: ", err)
	}

	// Ensure that DUT starts in the initial mode.
	if curr, err := h.Reporter.CurrentBootMode(ctx); err != nil {
		s.Fatal("Checking boot mode at beginning of test: ", err)
	} else if curr != modes[0] {
		s.Logf("DUT started in boot mode %s. Setting up initial boot mode %s", curr, modes[0])
		if err = ms.RebootToMode(ctx, modes[0]); err != nil {
			s.Fatalf("Failed to reboot to initial mode %s: %+v", modes[0], err)
		}
	}

	// Transition through the boot modes enumerated in ms, verifying boot mode at each step along the way.
	var fromMode, toMode fwCommon.BootMode
	for i := 0; i < len(modes)-1; i++ {
		fromMode, toMode = modes[i], modes[i+1]
		s.Logf("Beginning transition %d of %d: %s -> %s", i+1, len(modes)-1, fromMode, toMode)
		if err := ms.RebootToMode(ctx, toMode); err != nil {
			s.Fatalf("Error during transition from %s to %s: %+v", fromMode, toMode, err)
		}
	}
}
