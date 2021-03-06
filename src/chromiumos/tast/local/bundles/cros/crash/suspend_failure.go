// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crash

import (
	"context"
	"strings"

	"chromiumos/tast/local/crash"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: SuspendFailure,
		Desc: "Verify suspend failures are logged as expected",
		Contacts: []string{
			"dbasehore@google.com",
			"mutexlox@google.com",
			"cros-telemetry@google.com",
		},
		Attr: []string{"group:mainline", "informational"},
	})
}

func SuspendFailure(ctx context.Context, s *testing.State) {
	const suspendFailureName = "suspend-failure"

	if err := crash.SetUpCrashTest(ctx, crash.WithMockConsent()); err != nil {
		s.Fatal("SetUpCrashTest failed: ", err)
	}
	defer crash.TearDownCrashTest(ctx)

	// Restart anomaly detector to clear its cache of recently seen suspend
	// failures and ensure this one is logged.
	if err := crash.RestartAnomalyDetector(ctx); err != nil {
		s.Fatal("Failed to restart anomaly detector: ", err)
	}

	s.Log("Inducing artificial suspend permission failure")
	perm, err := testexec.CommandContext(ctx, "stat", "--format=%a", "/sys/power/state").Output(testexec.DumpLogOnError)
	if err != nil {
		s.Fatal("Failed to run 'stat -c='%a' /sys/power/state': ", err)
	}

	// Remove all permissions on /sys/power/state to induce the failure on write
	err = testexec.CommandContext(ctx, "chmod", "000", "/sys/power/state").Run()
	if err != nil {
		s.Fatal("Failed to set permissions on /sys/power/state: ", err)
	}

	// Error is expected here. Set a 60 second wakeup just in case suspend
	// somehow works here.
	err = testexec.CommandContext(ctx, "powerd_dbus_suspend", "--timeout=30", "--wakeup_timeout=60").Run()
	if err == nil {
		s.Error("powerd_dbus_suspend didn't fail when we expect it to")
	}

	// Restart powerd since it's still trying to suspend (which we don't want to
	// happen right now).
	err = testexec.CommandContext(ctx, "restart", "powerd").Run()
	if err != nil {
		s.Error("Failed to restart powerd: ", err)
		// If we fail to restart powerd, we'll shut down in ~100 seconds, so
		// just reboot.
		testexec.CommandContext(ctx, "reboot").Run()
	}

	err = testexec.CommandContext(ctx, "chmod", strings.TrimRight(string(perm), "\r\n"), "/sys/power/state").Run()
	if err != nil {
		s.Errorf("Failed to reset permissions (%v) on /sys/power/state: %v", perm, err)
		// We're messed up enough that rebooting the machine to reset the file
		// permissions on /sys/power/state is best here.
		testexec.CommandContext(ctx, "reboot").Run()
	}

	expectedRegexes := []string{`suspend_failure\.\d{8}\.\d{6}\.\d+\.0\.log`,
		`suspend_failure\.\d{8}\.\d{6}\.\d+\.0\.meta`}

	files, err := crash.WaitForCrashFiles(ctx, []string{crash.SystemCrashDir}, expectedRegexes)
	if err != nil {
		s.Fatal("Couldn't find expected files: ", err)
	}
	if err := crash.RemoveAllFiles(ctx, files); err != nil {
		s.Log("Couldn't clean up files: ", err)
	}
}
