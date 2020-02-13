// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crash contains utilties common to tests that use crash_reporter and
// crash_sender.
package crash

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/fsutil"
	"chromiumos/tast/local/set"
	"chromiumos/tast/testing"
)

const (
	crashTestInProgressDir = "/run/crash_reporter"
	// crashTestInProgressFile is a special control file that tells crash_reporter
	// to act normally during a crash test. Usually, crash_reporter is being told
	// (by /mnt/stateful_partition/etc/collect_chrome_crashes) to be more
	// aggressive about gathering crash data so that we can debug other, non-
	// crash_reporter tests more easily.
	crashTestInProgressFile = "crash-test-in-progress"
	// anomalyDetectorReadyFile is an indicator that the anomaly detector
	// has started and can detect any new anomalies.
	anomalyDetectorReadyFile = "anomaly-detector-ready"
	// mockConsentFile is a special control file that tells crash_reporter and
	// crash_sender to act as if the user has given consent for crash collection
	// and uploading.
	mockConsentFile = "mock-consent"
	// senderPausePath is the path to the file whose existence indicates that
	// crash_sender should be paused.
	senderPausePath = "/var/lib/crash_sender_paused"
	// senderProcName is the name of the crash_sender process.
	senderProcName = "crash_sender"
	// SystemCrashDir is the directory where system crash reports go.
	SystemCrashDir = "/var/spool/crash"
	// systemCrashStash is a directory to stash pre-existing system crashes during crash tests.
	systemCrashStash = "/var/spool/crash.real"
	// LocalCrashDir is the directory where user crash reports go.
	LocalCrashDir = "/home/chronos/crash"
	// localCrashStash is a directory to stash pre-existing user crashes during crash tests.
	localCrashStash = "/home/chronos/crash.real"
	// UserCrashDir is the directory where crash reports of currently logged in user go.
	UserCrashDir = "/home/chronos/user/crash"
	// userCrashStash is a directory to stash pre-existing crash reports of currently logged in user during crash tests.
	userCrashStash = "/home/chronos/user/crash.real"
	// userCrashDirs is used for finding the directory name containing a hash for current logged-in user,
	// in order to compare it with crash reporter log.
	userCrashDirs = "/home/chronos/u-*/crash"

	// BIOSExt is the extension for bios crash files.
	BIOSExt = ".bios_log"
	// CoreExt is the extension for core files.
	CoreExt = ".core"
	// MinidumpExt is the extension for minidump crash files.
	MinidumpExt = ".dmp"
	// LogExt is the extension for log files containing additional information that are written by crash_reporter.
	LogExt = ".log"
	// InfoExt is the extention for info files.
	InfoExt = ".info"
	// ProclogExt is the extention for proclog files.
	ProclogExt = ".proclog"
	// KCrashExt is the extension for log files created by kernel warnings and crashes.
	KCrashExt = ".kcrash"
	// GPUStateExt is the extension for GPU state files written by crash_reporter.
	GPUStateExt = ".i915_error_state.log.xz"
	// MetadataExt is the extension for metadata files written by crash collectors and read by crash_sender.
	MetadataExt = ".meta"
	// CompressedTxtExt is an extension on the compressed log files written by crash_reporter.
	CompressedTxtExt = ".txt.gz"
	// CompressedLogExt is an extension on the compressed log files written by crash_reporter.
	CompressedLogExt = ".log.gz"
	// DevCoredumpExt is an extension for device coredump files.
	DevCoredumpExt = ".devcore"

	// ChromeVerboseConsentFlags provides the flags to enable verbose logging about consent.
	ChromeVerboseConsentFlags = "--vmodule=stats_reporting_controller=1,autotest_private_api=1"
)

// DefaultDirs returns all standard directories to which crashes are written.
func DefaultDirs() []string {
	return []string{SystemCrashDir, LocalCrashDir, UserCrashDir}
}

// isCrashFile returns true if filename could be the name of a file generated by
// crashes or crash_reporter.
func isCrashFile(filename string) bool {
	knownExts := []string{
		BIOSExt,
		CoreExt,
		MinidumpExt,
		LogExt,
		ProclogExt,
		InfoExt,
		KCrashExt,
		GPUStateExt,
		MetadataExt,
		CompressedTxtExt,
		CompressedLogExt,
		DevCoredumpExt,
	}
	for _, ext := range knownExts {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// GetCrashes returns the paths of all files in dirs generated in response to crashes.
// Nonexistent directories are skipped.
func GetCrashes(dirs ...string) ([]string, error) {
	var crashFiles []string
	for _, dir := range dirs {
		files, err := ioutil.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		for _, fi := range files {
			if isCrashFile(fi.Name()) {
				crashFiles = append(crashFiles, filepath.Join(dir, fi.Name()))
			}
		}
	}
	return crashFiles, nil
}

// GetCrashDir gives the path to the crash directory for given username.
func GetCrashDir(username string) (string, error) {
	if username == "root" || username == "crash" {
		return SystemCrashDir, nil
	}
	p, err := filepath.Glob(userCrashDirs)
	if err != nil {
		// This only happens when userCrashDirs is malformed.
		return "", errors.Wrapf(err, "failed to list up files with pattern [%s]", userCrashDirs)
	}
	if len(p) == 0 {
		return LocalCrashDir, nil
	}
	if len(p) > 1 {
		return "", errors.Errorf("Wrong number of users logged in; got %d, want 1 or 0", len(p))
	}
	return p[0], nil
}

// WaitForCrashFiles waits for each regex in regexes to match a file in dirs that is not also in oldFiles.
// One might use it by
// 1. Getting a list of already-extant files in a directory.
// 2. Doing some operation that will create new files in that directory (e.g. inducing a crash).
// 3. Calling this method to wait for the expected files to appear.
// On success, WaitForCrashFiles returns a map from a regex to a list of files that matched that regex.
// If any regex was not matched, instead returns an error.
//
// When it comes to deleting files, tests should:
//   * Remove matching files that they expect to generate
//   * Leave matching files they do not expect to generate
// If there are more matches than expected and the test can't tell which are expected, it shouldn't delete any.
func WaitForCrashFiles(ctx context.Context, dirs, oldFiles, regexes []string) (map[string][]string, error) {
	var files map[string][]string
	err := testing.Poll(ctx, func(c context.Context) error {
		var newFiles []string
		for _, dir := range dirs {
			dirFiles, err := GetCrashes(dir)
			if err != nil {
				return testing.PollBreak(errors.Wrap(err, "failed to get new crashes"))
			}
			newFiles = append(newFiles, dirFiles...)
		}
		diffFiles := set.DiffStringSlice(newFiles, oldFiles)

		// Reset files each time the poll function is invoked, to avoid
		// repeatedly adding the same file
		files = make(map[string][]string)

		// track regexes that weren't matched.
		var missing []string
		for _, re := range regexes {
			match := false
			for _, f := range diffFiles {
				var err error
				match, err = regexp.MatchString(re, f)
				if err != nil {
					return testing.PollBreak(errors.Wrapf(err, "invalid regexp %s", re))
				}
				if match {
					files[re] = append(files[re], f)
					break
				}
			}
			if !match {
				missing = append(missing, re)
			}
		}
		if len(missing) != 0 {
			return errors.Errorf("no file matched %s (found %s)", strings.Join(missing, ", "), strings.Join(diffFiles, ", "))
		}
		return nil
	}, &testing.PollOptions{Timeout: 15 * time.Second})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// MoveFilesToOut moves all given files to s.OutDir(). Useful when further
// investigation of some files is needed to debug a test failure.
func MoveFilesToOut(ctx context.Context, outDir string, files ...string) error {
	var firstErr error
	for _, f := range files {
		base := filepath.Base(f)
		testing.ContextLogf(ctx, "Saving %s", base)
		if err := fsutil.MoveFile(f, filepath.Join(outDir, base)); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			testing.ContextLogf(ctx, "Couldn't save %s: %v", base, err)
		}
	}
	return firstErr
}

// RemoveAllFiles removes all files in the values of |map|.
func RemoveAllFiles(ctx context.Context, files map[string][]string) error {
	var firstErr error
	for _, v := range files {
		for _, f := range v {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				if firstErr == nil {
					firstErr = err
				}
				testing.ContextLogf(ctx, "Couldn't clean up %s: %v", f, err)
			}
		}
	}
	return firstErr
}
