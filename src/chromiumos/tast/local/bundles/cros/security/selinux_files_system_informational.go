// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package security

import (
	"context"
	"os"
	"syscall"

	"chromiumos/tast/local/bundles/cros/security/selinux"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         SELinuxFilesSystemInformational,
		Desc:         "Checks that SELinux file labels are set correctly for system files (new testcases, flaky testcases)",
		Contacts:     []string{"fqj@chromium.org", "kroot@chromium.org", "chromeos-security@google.com"},
		Attr:         []string{"informational"},
		SoftwareDeps: []string{"selinux"},
	})
}

func SELinuxFilesSystemInformational(ctx context.Context, s *testing.State) {
	type rwFilter int
	const (
		readonly rwFilter = iota
		writable
	)
	systemCPUFilter := func(writableFilter rwFilter) selinux.FileLabelCheckFilter {
		return func(p string, fi os.FileInfo) (skipFile, skipSubdir selinux.FilterResult) {
			mode := fi.Mode()
			// Domain has search to both sysfs and sysfs_devices_system_cpu.
			if mode.IsDir() {
				return selinux.Skip, selinux.Check
			}

			isWritable := mode.IsRegular() && ((mode.Perm() & (syscall.S_IWUSR | syscall.S_IWGRP | syscall.S_IWOTH)) > 0)
			// Writable files
			if isWritable != (writableFilter == writable) {
				return selinux.Skip, selinux.Check
			}

			return selinux.Check, selinux.Check
		}
	}

	// Files to be tested.
	testArgs := []selinux.FileTestCase{
		{Path: "/sys/devices/system/cpu", Context: "sysfs", Recursive: true, Filter: systemCPUFilter(writable), Log: false},
		{Path: "/sys/devices/system/cpu", Context: "sysfs_devices_system_cpu", Recursive: true, Filter: systemCPUFilter(readonly), Log: false},
		{Path: "/var/log/arc.log", Context: "cros_arc_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/authpolicy.log", Context: "cros_authpolicy_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/boot.log", Context: "cros_boot_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/eventlog.txt", Context: "cros_var_log_eventlog", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/messages", Context: "cros_syslog", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/mount-encrypted.log", Context: "cros_var_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/net.log", Context: "cros_net_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/secure", Context: "cros_secure_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/tlsdate.log", Context: "cros_tlsdate_log", Recursive: false, Filter: nil, Log: true},
		{Path: "/var/log/asan", Context: "cros_var_log_asan", Recursive: true, Filter: nil, Log: true},
	}

	selinux.FilesTestInternal(ctx, s, testArgs)
}
