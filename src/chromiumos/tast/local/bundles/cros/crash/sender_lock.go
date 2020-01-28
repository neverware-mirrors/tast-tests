// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crash

import (
	"context"
	"os"

	"golang.org/x/sys/unix"

	"chromiumos/tast/local/bundles/cros/crash/sender"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: SenderLock,
		Desc: "Check that only one crash_sender runs at a time",
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

func SenderLock(ctx context.Context, s *testing.State) {
	if err := sender.SetUp(ctx, s.PreValue().(*chrome.Chrome)); err != nil {
		s.Fatal("Setup failed: ", err)
	}
	defer sender.TearDown()

	const basename = "some_program.1.2.3"
	if _, err := sender.AddFakeMinidumpCrash(ctx, basename); err != nil {
		s.Fatal("Failed to add a fake minidump crash: ", err)
	}

	// Obtain the crash_sender lock. This should prevent crash_sender from running.
	const lockPath = "/run/lock/crash_sender"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.Fatal("Failed to obtain crash_sender lock: ", err)
	}
	defer f.Close()
	if err := unix.FcntlFlock(f.Fd(), unix.F_SETLK, &unix.Flock_t{Type: unix.F_WRLCK}); err != nil {
		s.Fatal("Failed to obtain crash_sender lock: ", err)
	}

	if _, err := sender.Run(ctx); err == nil {
		s.Fatal("crash_sender succeeded unexpectedly")
	}
	s.Log("crash_sender failed as expected")
}
