// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crash

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"chromiumos/tast/local/bundles/cros/crash/sender"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: SenderPaused,
		Desc: "Check that crash_sender is paused during crash tests",
		Contacts: []string{
			"mutexlox@chromium.org",
			"iby@chromium.org",
			"cros-monitoring-forensics@google.com",
			"nya@chromium.org", // ported to Tast
		},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome", "metrics_consent"},
		Pre:          chrome.LoggedIn(),
	})
}

func SenderPaused(ctx context.Context, s *testing.State) {
	if err := sender.SetUp(ctx, s.PreValue().(*chrome.Chrome)); err != nil {
		s.Fatal("Setup failed: ", err)
	}
	defer sender.TearDown()

	const basename = "some_program.1.2.3"
	exp, err := sender.AddFakeMinidumpCrash(ctx, basename)
	if err != nil {
		s.Fatal("Failed to add a fake minidump crash: ", err)
	}

	// Run crash_sender without ignoring the pause file. This should fail because
	// sender.SetUp creates the pause file.
	if _, err := sender.RunNoIgnorePauseFile(ctx); err == nil {
		s.Fatal("crash_sender succeeded unexpectedly")
	}
	s.Log("crash_sender failed as expected")

	// Run crash_sender again, ignoring the pause file this time. This should pass.
	got, err := sender.Run(ctx)
	if err != nil {
		s.Fatal("Failed to run crash_sender: ", err)
	}
	want := []*sender.SendResult{{
		Success: true,
		Data:    *exp,
	}}
	if diff := cmp.Diff(got, want, cmpopts.IgnoreFields(sender.SendResult{}, "Schedule")); diff != "" {
		s.Log("Results mismatch (-got +want): ", diff)
		s.Errorf("crash_sender sent unexpected %d results; see logs for diff", len(got))
	}
}
