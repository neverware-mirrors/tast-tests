// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crash

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/sysutil"
)

// moveAllCrashesTo moves crashes from |source| to |target|. This allows us to
// start crash tests with an empty spool directory, reducing risk of flakes if
// the dir is already full when the test starts.
func moveAllCrashesTo(source, target string) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return errors.Wrapf(err, "couldn't make stash crash dir %s", target)
	}
	files, err := ioutil.ReadDir(source)
	if err != nil {
		// Bubble this up so caller can check whether IsNotExist and behave accordingly.
		return err
	}
	for _, f := range files {
		if err := os.Rename(filepath.Join(source, f.Name()), filepath.Join(target, f.Name())); err != nil {
			return errors.Wrapf(err, "couldn't move file: %v", f.Name())
		}
	}
	return nil
}

// SetUpCrashTest indicates that we are running a test that involves the crash
// reporting system (crash_reporter, crash_sender, or anomaly_detector). The
// test should "defer TearDownCrashTest()" after calling this. If developer image
// behavior is required for the test, call SetUpDevImageCrashTest instead.
func SetUpCrashTest() error {
	return setUpCrashTestWithDirectories(crashTestInProgressDir, SystemCrashDir, systemCrashStash,
		LocalCrashDir, localCrashStash, false)
}

// SetUpDevImageCrashTest stashes away existing crash files to prevent tests which
// generate crashes from failing due to full crash directories. This function does
// not indicate to the DUT that a crash test is in progress, allowing the test to
// complete with standard developer image behavior. The test should
// "defer TearDownCrashTest()" after calling this
func SetUpDevImageCrashTest() error {
	return setUpCrashTestWithDirectories(crashTestInProgressDir, SystemCrashDir, systemCrashStash,
		LocalCrashDir, localCrashStash, true)
}

// setUpCrashTestWithDirectories is a helper function for SetUpCrashTest. We need
// this as a separate function for testing.
func setUpCrashTestWithDirectories(inProgDir, sysCrashDir, sysCrashStash, userCrashDir, userCrashStash string, isDevImageTest bool) (retErr error) {
	// Move all crashes into stash directory so a full directory won't stop
	// us from saving a new crash report
	if err := moveAllCrashesTo(sysCrashDir, sysCrashStash); err != nil && !os.IsNotExist(err) {
		return err
	}
	defer func() {
		if retErr != nil {
			cleanUpStashDir(sysCrashStash, sysCrashDir)
		}
	}()

	if err := moveAllCrashesTo(userCrashDir, userCrashStash); err != nil && !os.IsNotExist(err) {
		return err
	}
	defer func() {
		if retErr != nil {
			cleanUpStashDir(userCrashStash, userCrashDir)
		}
	}()

	// If the test is meant to run with developer image behavior, return here to
	// avoid creating the directory that indicates a crash test is in progress.
	if isDevImageTest {
		return nil
	}

	if err := os.MkdirAll(inProgDir, 0755); err != nil {
		return errors.Wrapf(err, "could not make directory %v", inProgDir)
	}

	filePath := filepath.Join(inProgDir, crashTestInProgressFile)
	if err := ioutil.WriteFile(filePath, nil, 0644); err != nil {
		return errors.Wrapf(err, "could not create %v", filePath)
	}
	return nil
}

func cleanUpStashDir(stashDir, realDir string) error {
	// Stash dir should exist, so error if it doesn't.
	if err := moveAllCrashesTo(stashDir, realDir); err != nil {
		return err
	}
	if err := os.Remove(stashDir); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "couldn't remove stash dir: %v", stashDir)
		}
	}
	return nil
}

// TearDownCrashTest undoes the work of SetUpCrashTest.
func TearDownCrashTest() error {
	if err := tearDownCrashTestWithDirectories(crashTestInProgressDir, SystemCrashDir, systemCrashStash,
		LocalCrashDir, localCrashStash); err != nil {
		return err
	}
	// The user crash directory should always be owned by chronos not root. The
	// unit tests don't run as root and can't chown, so skip this in tests.
	if err := os.Chown(LocalCrashDir, int(sysutil.ChronosUID), int(sysutil.ChronosGID)); err != nil {
		return errors.Wrapf(err, "couldn't chown %s", LocalCrashDir)
	}
	return nil
}

// tearDownCrashTestWithDirectories is a helper function for TearDownCrashTest. We need
// this as a separate function for testing.
func tearDownCrashTestWithDirectories(inProgDir, sysCrashDir, sysCrashStash, userCrashDir, userCrashStash string) error {
	var firstErr error
	if err := cleanUpStashDir(sysCrashStash, sysCrashDir); err != nil {
		firstErr = err
	}
	if err := cleanUpStashDir(userCrashStash, userCrashDir); err != nil && firstErr == nil {
		firstErr = err
	}

	filePath := filepath.Join(inProgDir, crashTestInProgressFile)
	if err := os.Remove(filePath); err != nil && firstErr == nil {
		if os.IsNotExist(err) {
			// Something else already removed the file or it was never created (See SetUpDevImageCrashTest).
			// Well, whatever, we're in the correct state now (the file is gone).
			return nil
		}
		return errors.Wrapf(err, "could not remove %v", filePath)
	}
	return firstErr
}
