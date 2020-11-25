// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/godbus/dbus"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/metrics"
	"chromiumos/tast/local/crash"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/crostini/ui/terminalapp"
	"chromiumos/tast/local/dbusutil"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/testing"
)

const (
	smallFile       = "/tmp/small_file"
	bigFile         = "/tmp/big_file"
	smallFileDest   = "/home/testuser/small_file"
	bigFileDest     = "/home/testuser/big_file"
	smallUUID       = "fd8a2552-6822-490a-ae67-d43c6ff6e8eb"
	bigUUID         = "ddddffbb-c479-4872-9421-bcf9f1764ed7"
	uuidReplacement = "00000000-0000-0000-0000-000000000000"

	anomalyEventServiceName              = "org.chromium.AnomalyEventService"
	anomalyEventServicePath              = dbus.ObjectPath("/org/chromium/AnomalyEventService")
	anomalyEventServiceInterface         = "org.chromium.AnomalyEventServiceInterface"
	anomalyGuestFileCorruptionSignalName = "GuestFileCorruption"
	fsCorruptionHistogram                = "Crostini.FilesystemCorruption"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: FsCorruption,
		Desc: "Check that fs corruption is detected correctly",
		Contacts: []string{
			// Crostini
			"sidereal@google.com",
			"cros-containers-dev@google.com",
			// Telemetry
			"mutexlox@google.com",
			"cros-telemetry@google.com",
		},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Attr:         []string{"group:mainline", "informational"},
		Vars:         []string{"keepState"},
		Params: []testing.Param{
			// Parameters generated by params_test.go. DO NOT EDIT.
			{
				Name:              "amd64_stable",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           10 * time.Minute,
			}, {
				Name:              "amd64_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           10 * time.Minute,
			}, {
				Name:              "arm_stable",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniStable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           10 * time.Minute,
			}, {
				Name:              "arm_unstable",
				ExtraAttr:         []string{"informational"},
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           10 * time.Minute,
			},
		},
	})
}

func createTestFile(ctx context.Context, container *vm.Container, hostPath, guestPath string, data []byte) error {
	if err := ioutil.WriteFile(hostPath, data, 0755); err != nil {
		return errors.Wrap(err, "failed to write to test file")
	}
	if err := container.PushFile(ctx, hostPath, guestPath); err != nil {
		return errors.Wrap(err, "failed to push test file to container")
	}
	return nil
}

func createTestFiles(ctx context.Context, container *vm.Container) error {
	// Small files will get incorporated into the filesystem metadata. Metadata and file data corruption produce different log messages, so we do both.
	if err := createTestFile(ctx, container, smallFile, smallFileDest, []byte(smallUUID)); err != nil {
		return err
	}

	data := make([]byte, 1024*1024)
	copy(data, []byte(bigUUID))
	if err := createTestFile(ctx, container, bigFile, bigFileDest, data); err != nil {
		return err
	}

	return nil
}

func scanBuffer(baseOffset int64, buffer, pattern []byte) []int64 {
	var result []int64
	for {
		idx := bytes.Index(buffer, pattern)
		if idx == -1 {
			break
		}

		newOffset := baseOffset + int64(idx)
		result = append(result, newOffset)

		buffer = buffer[idx+1:]
		baseOffset += int64(idx) + 1
	}
	return result
}

// getOffsets finds all locations of two patterns (smallPattern and
// bigPattern) in the given filepath. Returns two slices of offsets
// into the file, one for each pattern, and an error (if any).
//
// We can't use things like Scanner here because we're scanning a
// binary file with no limit on the spacing between \n characters, and
// Scanner expects to be operating on a UTF-8 character encoding with
// reasonably sized lines. Instead we have a hand written loop.
func getOffsets(ctx context.Context, filepath string, smallPattern, bigPattern []byte) (_, _ []int64, resultError error) {
	const oneMiB = 1024 * 1024

	// Buffer size = 1MiB + a small amount of overlap in case one of the UUIDs is split across a boundary
	overlap := len(smallPattern)
	if len(bigPattern) > overlap {
		overlap = len(bigPattern)
	}
	buffer := make([]byte, oneMiB+overlap-1)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open VM disk for reading")
	}
	defer func() {
		if err := file.Close(); err != nil {
			if resultError == nil {
				resultError = err
			} else {
				testing.ContextLog(ctx, "Failed to close file: ", err)
			}
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to stat VM disk")
	}
	size := stat.Size()

	var smallOffsets, bigOffsets []int64
	for i := int64(0); i < size; i += oneMiB {
		read, err := file.ReadAt(buffer, i)
		if err != nil && err != io.EOF {
			return nil, nil, errors.Wrapf(err, "failed to read a full buffer from VM disk at offset %d, got %d bytes, expected %d", i, read, len(buffer))
		}
		if err == io.EOF {
			// We read less then the full buffer because we hit EOF, so truncate the buffer
			buffer = buffer[:read]
		}

		smallOffsets = append(smallOffsets, scanBuffer(i, buffer[:oneMiB+len(smallPattern)-1], smallPattern)...)
		bigOffsets = append(bigOffsets, scanBuffer(i, buffer[:oneMiB+len(bigPattern)-1], bigPattern)...)
	}

	return smallOffsets, bigOffsets, nil
}

func writeAtSync(ctx context.Context, filepath string, b []byte, offsets []int64) (resultError error) {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to open VM disk for editing")
	}
	defer func() {
		if err := file.Close(); err != nil {
			if resultError == nil {
				resultError = err
			} else {
				testing.ContextLog(ctx, "Failed to close file: ", err)
			}
		}
	}()

	for _, offset := range offsets {
		if _, err := file.WriteAt(b, offset); err != nil {
			return errors.Wrapf(err, "failed to write to file at offset %d", offset)
		}
	}
	if err := file.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync write to disk")
	}

	return nil
}

func waitForSignal(ctx context.Context, signalWatcher *dbusutil.SignalWatcher) (*dbus.Signal, error) {
	select {
	case signal := <-signalWatcher.Signals:
		return signal, nil
	case <-ctx.Done():
		return nil, errors.New("Context deadline expired")
	}
}

func checkHistogram(ctx context.Context, tconn *chrome.TestConn, baseline int64) (int64, error) {
	hist, err := metrics.GetHistogram(ctx, tconn, fsCorruptionHistogram)
	if err != nil {
		return 0, err
	}
	if hist.Sum <= baseline {
		return hist.Sum, errors.Errorf("expected total of more then %v histogram values, got %v", baseline, hist.Sum)
	}
	return hist.Sum, nil
}

func readContainerFile(ctx context.Context, container *vm.Container, path string) error {
	output, err := container.Command(ctx, "cat", path).CombinedOutput()
	if err == nil {
		testing.ContextLogf(ctx, "Read succesfully from path %s in the container", path)
		return nil
	}

	// We expect, at least some of the time, to get an error while
	// running this command because the file is corrupted. But we
	// might also get an error from a different source. So we try
	// to filter out the expected error from other errors by
	// matching on the output in stderr, the exit code, and the
	// concrete type of the error.

	expectedOutput := fmt.Sprintf("cat: %s: Input/output error\n", path)
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if string(output) == expectedOutput && exitError.ExitCode() == 1 {
			testing.ContextLogf(ctx, "Got expected error while reading file: %v, combined output: %q", exitError, output)
			return nil
		}
		return errors.Wrapf(exitError, "got unexpected error while reading file, combined output: %q", output)
	}
	return errors.Wrap(err, "got unexpected error while reading file")
}

// testOverwriteAtOffsets overwrites the VM disk that stores
// container at the locations in offsets with uuidReplacement. It
// then restarts the VM and container and checks that the filesystem
// corruption is detected. Finally, it stops the VM and restores the VM
// disk from backupPath. outDir is passed to
// vm.RestartDefaultVMContainer and may be used to store logs from
// container startup on failure.
//
// We assume that the VM is initially stopped and that the disk is in
// a good (uncorrupted) state. If this function returns successfully,
// the VM will be stopped, but the disk is in an unspecified state.
func testOverwriteAtOffsets(ctx context.Context, tconn *chrome.TestConn, offsets []int64, container *vm.Container, diskPath, backupPath, outDir string) (resultError error) {
	match := dbusutil.MatchSpec{
		Type:      "signal",
		Path:      anomalyEventServicePath,
		Interface: anomalyEventServiceInterface,
		Member:    anomalyGuestFileCorruptionSignalName,
	}
	signalWatcher, err := dbusutil.NewSignalWatcherForSystemBus(ctx, match)
	if err != nil {
		return errors.Wrap(err, "failed to listed for DBus signals")
	}
	defer func() {
		if err := signalWatcher.Close(ctx); err != nil {
			if resultError == nil {
				resultError = err
			} else {
				testing.ContextLog(ctx, "Failed to close signal watcher: ", err)
			}
		}
	}()

	// Make edit to disk at these offsets.
	testing.ContextLog(ctx, "Making changes at offsets ", offsets)
	if err := writeAtSync(ctx, diskPath, []byte(uuidReplacement), offsets); err != nil {
		return errors.Wrap(err, "failed to make disk edit")
	}

	// Make sure to stop the VM before returning from this
	// function. We do this by a direct call to concierge since we
	// want to ensure it isn't running even when there may be an
	// error in the startup process.
	defer func() {
		if err := container.VM.Stop(ctx); err != nil {
			if resultError == nil {
				resultError = err
			} else {
				testing.ContextLog(ctx, "Failed to stop VM: ", err)
			}
		}
	}()

	testing.ContextLog(ctx, "Restarting VM")
	var vmRunning bool
	// Discard the error, as this may fail due to corruption.
	if terminal, _ := terminalapp.Launch(ctx, tconn); terminal != nil {
		// If we got a terminal object from Launch, we need to
		// call Close to free its internal UI node.
		if err := terminal.Close(ctx); err != nil {
			return err
		}
		vmRunning = true
	}

	// Filesystem corruption doesn't get detected until some process tries to read from the corrupted location. For metadata, this usually happens during container startup, but if the container started successfully then we read from both files just to be sure.
	if vmRunning {
		testing.ContextLog(ctx, "Attempting to read corrupted files")
		if err := readContainerFile(ctx, container, bigFileDest); err != nil {
			return errors.Wrap(err, "failed to read big file")
		}
		if err := readContainerFile(ctx, container, smallFileDest); err != nil {
			return errors.Wrap(err, "failed to read small file")
		}
	} else {
		testing.ContextLog(ctx, "Terminal failed to start, not trying to read files")
	}

	testing.ContextLog(ctx, "Waiting for signal from anomaly_detector")
	signalCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := waitForSignal(signalCtx, signalWatcher); err != nil {
		return errors.Wrap(err, "didn't get expected DBus signal")
	}
	testing.ContextLog(ctx, "Got expected signal from anomaly_detector")

	return nil
}

func launchAndReleaseTerminal(ctx context.Context, tconn *chrome.TestConn) error {
	terminal, err := terminalapp.Launch(ctx, tconn)
	if terminal != nil {
		err = terminal.Close(ctx)
	}
	return err
}

// FsCorruption sets up the VM and then introduces corruption into its disk to check that this is detected correctly.
func FsCorruption(ctx context.Context, s *testing.State) {
	data := s.PreValue().(crostini.PreData)
	tconn := data.TestAPIConn
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	if err := crash.SetUpCrashTest(ctx, crash.WithMockConsent()); err != nil {
		s.Fatal("Failed to set up crash test: ", err)
	}
	defer func() {
		if err := crash.TearDownCrashTest(ctx); err != nil {
			s.Error("Failed to tear down crash test fixture: ", err)
		}
	}()

	s.Log("Writing test file to container")
	if err := createTestFiles(ctx, data.Container); err != nil {
		s.Fatal("Failed to create test files: ", err)
	}

	disk, err := data.Container.VM.Concierge.GetVMDiskInfo(ctx, vm.DefaultVMName)
	if err != nil {
		s.Fatal("Failed to get VM disk info: ", err)
	}

	// Stop the VM so it isn't running while we edit its disk.
	s.Log("Stopping VM")
	if err := data.Container.VM.Stop(ctx); err != nil {
		s.Fatal("Failed to stop VM: ", err)
	}
	// Restart everything before finishing so the precondition will be in a good state.
	defer func() {
		if err := launchAndReleaseTerminal(ctx, tconn); err != nil {
			s.Error("Failed to restart crostini terminal: ", err)
		}
	}()

	s.Log("Searching for pattern in disk image")
	smallOffsets, bigOffsets, err := getOffsets(ctx, disk.GetPath(), []byte(smallUUID), []byte(bigUUID))
	if err != nil || len(smallOffsets) == 0 || len(bigOffsets) == 0 {
		s.Fatalf("Failed to get file offsets: %v, %v, %v", err, smallOffsets, bigOffsets)
	}

	// BTRFS filesystems are modified on every mount, so we make a backup here of the disk so we can start each corruption from a known state.
	s.Log("Backing up the current disk image")
	backupPath := disk.GetPath() + ".bak"
	cmd := testexec.CommandContext(ctx, "cp", "--sparse=always", "--backup=off", "--preserve=all", disk.GetPath(), backupPath)
	if err := cmd.Run(testexec.DumpLogOnError); err != nil {
		testexec.CommandContext(ctx, "rm", "--force", backupPath).Run()
		s.Fatal("Failed to back up VM disk for editing: ", err)
	}

	// Always restore the backup disk before ending the test.
	defer func() {
		cmd := testexec.CommandContext(ctx, "mv", "--force", backupPath, disk.GetPath())
		if err := cmd.Run(testexec.DumpLogOnError); err != nil {
			s.Error("Failed to restore VM disk from backup: ", err)
		}
	}()

	histogramCount, err := checkHistogram(ctx, tconn, -1)
	if err != nil {
		s.Fatal("Failed to get baseline for histogram: ", err)
	}

	if err := testOverwriteAtOffsets(ctx, tconn, bigOffsets, data.Container, disk.GetPath(), backupPath, s.OutDir()); err != nil {
		s.Fatal("Didn't get an error signal for big file: ", err)
	}
	histogramCount, err = checkHistogram(ctx, tconn, histogramCount)
	if err != nil {
		s.Fatal("Failed to check histogram: ", err)
	}

	// Restore the disk from backup because the previous test will have corrutped it.
	cmd = testexec.CommandContext(ctx, "cp", "--sparse=always", "--backup=off", "--preserve=all", backupPath, disk.GetPath())
	if err := cmd.Run(testexec.DumpLogOnError); err != nil {
		s.Fatal("Failed to restore disk from backup between tests: ", err)
	}

	if err := testOverwriteAtOffsets(ctx, tconn, smallOffsets, data.Container, disk.GetPath(), backupPath, s.OutDir()); err != nil {
		s.Fatal("Didn't get an error signal for small file: ", err)
	}
	if _, err := checkHistogram(ctx, tconn, histogramCount); err != nil {
		s.Fatal("Failed to check histogram: ", err)
	}
}
