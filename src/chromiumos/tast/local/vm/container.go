// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vm

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/godbus/dbus"

	cpb "chromiumos/system_api/vm_cicerone_proto" // protobufs for container management
	"chromiumos/tast/errors"
	"chromiumos/tast/local/dbusutil"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/testing"
)

const (
	liveContainerImageServerFormat    = "https://storage.googleapis.com/cros-containers/%d"         // simplestreams image server being served live
	stagingContainerImageServerFormat = "https://storage.googleapis.com/cros-containers-staging/%d" // simplestreams image server for staging

	testContainerUsername = "testuser"            // default container username during testing
	testImageAlias        = "debian/stretch/test" // default container alias

	ciceroneName      = "org.chromium.VmCicerone"
	ciceronePath      = dbus.ObjectPath("/org/chromium/VmCicerone")
	ciceroneInterface = "org.chromium.VmCicerone"
)

// ContainerType represents the container image type to be downloaded.
type ContainerType int

const (
	// LiveImageServer indicates that the current live container image should be downloaded.
	LiveImageServer ContainerType = iota
	// StagingImageServer indicates that the current staging container image should be downloaded.
	StagingImageServer
)

// Container encapsulates a container running in a VM.
type Container struct {
	// VM is the VM in which this container is running.
	VM            *VM
	containerName string // name of the container
	username      string // username of the container's primary user
	ciceroneObj   dbus.BusObject
}

// DefaultContainer returns a container object with default settings.
func DefaultContainer(vmInstance *VM) *Container {
	return &Container{
		VM:            vmInstance,
		containerName: DefaultContainerName,
		username:      testContainerUsername,
	}
}

// Create will create a Linux container in an existing VM. It returns without waiting for the creation to complete.
// One must listen on cicerone D-Bus signals to know the creation is done.
// TODO(851207): Make a minimal Linux container for testing so this completes
// fast enough to use in bvt.
func (c *Container) Create(ctx context.Context, t ContainerType) error {
	var err error
	if _, c.ciceroneObj, err = dbusutil.Connect(ctx, ciceroneName, ciceronePath); err != nil {
		return err
	}

	milestone, err := getMilestone()
	if err != nil {
		return err
	}
	var server string
	switch t {
	case LiveImageServer:
		server = fmt.Sprintf(liveContainerImageServerFormat, milestone)
	case StagingImageServer:
		server = fmt.Sprintf(stagingContainerImageServerFormat, milestone)
	}

	resp := &cpb.CreateLxdContainerResponse{}
	if err = dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".CreateLxdContainer",
		&cpb.CreateLxdContainerRequest{
			VmName:        c.VM.name,
			ContainerName: DefaultContainerName,
			OwnerId:       c.VM.Concierge.ownerID,
			ImageServer:   server,
			ImageAlias:    testImageAlias,
		}, resp); err != nil {
		return err
	}

	switch resp.GetStatus() {
	case cpb.CreateLxdContainerResponse_UNKNOWN, cpb.CreateLxdContainerResponse_FAILED:
		return errors.Errorf("failed to create container: %v", resp.GetFailureReason())
	case cpb.CreateLxdContainerResponse_EXISTS:
		return errors.New("container already exists")
	}
	return nil
}

// start launches a Linux container in an existing VM.
func (c *Container) start(ctx context.Context) error {
	starting, err := dbusutil.NewSignalWatcherForSystemBus(ctx, ciceroneDBusMatchSpec("LxdContainerStarting"))
	if err != nil {
		return err
	}
	// Always close the LxdContainerStarting watcher regardless of success.
	defer starting.Close(ctx)

	resp := &cpb.StartLxdContainerResponse{}
	if err := dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".StartLxdContainer",
		&cpb.StartLxdContainerRequest{
			VmName:        c.VM.name,
			ContainerName: c.containerName,
			OwnerId:       c.VM.Concierge.ownerID,
			Async:         true,
		}, resp); err != nil {
		return err
	}

	switch resp.GetStatus() {
	case cpb.StartLxdContainerResponse_RUNNING:
		return errors.New("container is already running")
	case cpb.StartLxdContainerResponse_STARTING, cpb.StartLxdContainerResponse_REMAPPING:
	default:
		return errors.Errorf("failed to start container: %v", resp.GetFailureReason())
	}

	sigResult := &cpb.LxdContainerStartingSignal{}
	for sigResult.VmName != c.VM.name ||
		sigResult.ContainerName != c.containerName ||
		sigResult.OwnerId != c.VM.Concierge.ownerID {
		if err := waitForDBusSignal(ctx, starting, nil, sigResult); err != nil {
			return err
		}
	}

	if sigResult.Status != cpb.LxdContainerStartingSignal_STARTED {
		return errors.Errorf("container failed to start: %v", resp.GetFailureReason())
	}

	testing.ContextLogf(ctx, "Started container %q in VM %q", c.containerName, c.VM.name)
	return nil
}

// StartAndWait starts up an already created container and waits for that startup to complete
// before returning. The directory dir may be used to store logs on failure.
func (c *Container) StartAndWait(ctx context.Context, dir string) error {
	started, err := dbusutil.NewSignalWatcherForSystemBus(ctx, ciceroneDBusMatchSpec("ContainerStarted"))
	if err != nil {
		return err
	}
	// Always close the ContainerStarted watcher regardless of success.
	defer started.Close(ctx)

	if err = c.start(ctx); err != nil {
		return err
	}

	if err = c.SetUpUser(ctx); err != nil {
		if err := c.DumpLog(ctx, dir); err != nil {
			testing.ContextLog(ctx, "Failure dumping container log: ", err)
		}
		return err
	}

	testing.ContextLog(ctx, "Waiting for ContainerStarted D-Bus signal")
	sigResult := &cpb.ContainerStartedSignal{}
	for sigResult.VmName != c.VM.name ||
		sigResult.ContainerName != c.containerName ||
		sigResult.OwnerId != c.VM.Concierge.ownerID {
		if err := waitForDBusSignal(ctx, started, nil, sigResult); err != nil {
			return err
		}
	}
	return nil
}

// GetUsername returns the default user in a container.
func (c *Container) GetUsername(ctx context.Context) (string, error) {
	resp := &cpb.GetLxdContainerUsernameResponse{}
	if err := dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".GetLxdContainerUsername",
		&cpb.GetLxdContainerUsernameRequest{
			VmName:        c.VM.name,
			ContainerName: c.containerName,
			OwnerId:       c.VM.Concierge.ownerID,
		}, resp); err != nil {
		return "", err
	}

	if resp.GetStatus() != cpb.GetLxdContainerUsernameResponse_SUCCESS {
		return "", errors.Errorf("failed to get username: %v", resp.GetFailureReason())
	}

	return resp.GetUsername(), nil
}

// SetUpUser sets up the default user in a container.
func (c *Container) SetUpUser(ctx context.Context) error {
	resp := &cpb.SetUpLxdContainerUserResponse{}
	if err := dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".SetUpLxdContainerUser",
		&cpb.SetUpLxdContainerUserRequest{
			VmName:            c.VM.name,
			ContainerName:     c.containerName,
			OwnerId:           c.VM.Concierge.ownerID,
			ContainerUsername: c.username,
		}, resp); err != nil {
		return err
	}

	if resp.GetStatus() != cpb.SetUpLxdContainerUserResponse_SUCCESS &&
		resp.GetStatus() != cpb.SetUpLxdContainerUserResponse_EXISTS {
		return errors.Errorf("failed to set up user: %v", resp.GetFailureReason())
	}

	testing.ContextLogf(ctx, "Set up user %q in container %q", c.username, c.containerName)
	return nil
}

// PushFile copies a local file to the container's filesystem.
func (c *Container) PushFile(ctx context.Context, localPath, containerPath string) error {
	testing.ContextLogf(ctx, "Copying local file %v to container %v", localPath, containerPath)
	// We base64 encode this and write it through terminal commands. We need to
	// base64 encode it since we are using the vsh command underneath which is a
	// terminal and binary control characters may interfere with its operation.
	fileData, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	// Remove the target file in case it already exists. "-f" to ignore nonexistent files.
	cmd := c.Command(ctx, "rm", "-f", containerPath)
	if err = cmd.Run(); err != nil {
		return err
	}

	// TODO(cylee): Workaround to break payload in pieces since shell has argument length limit.
	// Use sftp or scp instead (https://crbug.com/923721).
	const maxArgLen = 4000
	dataLen := len(base64Data)
	for start, end := 0, 0; start < dataLen; start = end {
		end = start + maxArgLen
		if end > dataLen {
			end = dataLen
		}
		cmd := c.Command(ctx, "sh", "-c", "echo '"+base64Data[start:end]+"' | base64 --decode >>"+containerPath)
		if err = cmd.Run(); err != nil {
			cmd.DumpLog(ctx)
			return err
		}
	}
	return nil
}

// LinuxPackageInfo queries the container for information about a Linux package
// file. The packageID returned corresponds to the package ID for an installed
// package based on the PackageKit specification which is of the form
// 'package_id;version;arch;repository'.
func (c *Container) LinuxPackageInfo(ctx context.Context, path string) (packageID string, err error) {
	resp := &cpb.LinuxPackageInfoResponse{}
	if err := dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".GetLinuxPackageInfo",
		&cpb.LinuxPackageInfoRequest{
			VmName:        c.VM.name,
			ContainerName: c.containerName,
			OwnerId:       c.VM.Concierge.ownerID,
			FilePath:      path,
		}, resp); err != nil {
		return "", err
	}

	if !resp.GetSuccess() {
		return "", errors.Errorf("failed to get Linux package info: %v", resp.GetFailureReason())
	}

	return resp.GetPackageId(), nil
}

// InstallPackage installs a Linux package file into the container.
func (c *Container) InstallPackage(ctx context.Context, path string) error {
	progress, err := dbusutil.NewSignalWatcherForSystemBus(ctx, ciceroneDBusMatchSpec("InstallLinuxPackageProgress"))
	if err != nil {
		return err
	}
	// Always close the InstallLinuxPackageProgress watcher regardless of success.
	defer progress.Close(ctx)

	resp := &cpb.InstallLinuxPackageResponse{}
	if err = dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".InstallLinuxPackage",
		&cpb.LinuxPackageInfoRequest{
			VmName:        c.VM.name,
			ContainerName: c.containerName,
			OwnerId:       c.VM.Concierge.ownerID,
			FilePath:      path,
		}, resp); err != nil {
		return err
	}

	if resp.Status != cpb.InstallLinuxPackageResponse_STARTED {
		return errors.Errorf("failed to start Linux package install: %v", resp.FailureReason)
	}

	// Wait for the signal for install completion which will signify success or
	// failure.
	testing.ContextLog(ctx, "Waiting for InstallLinuxPackageProgress D-Bus signal")
	sigResult := &cpb.InstallLinuxPackageProgressSignal{}
	for {
		if err := waitForDBusSignal(ctx, progress, nil, sigResult); err != nil {
			return err
		}
		if sigResult.VmName == c.VM.name &&
			sigResult.ContainerName == c.containerName &&
			sigResult.OwnerId == c.VM.Concierge.ownerID {
			if sigResult.Status == cpb.InstallLinuxPackageProgressSignal_SUCCEEDED {
				return nil
			}
			if sigResult.Status == cpb.InstallLinuxPackageProgressSignal_FAILED {
				return errors.Errorf("failure with Linux package install: %v", sigResult.FailureDetails)
			}
		}
	}
}

// UninstallPackageOwningFile uninstalls the package owning a particular desktop
// file in the container.
func (c *Container) UninstallPackageOwningFile(ctx context.Context, desktopFileID string) error {
	progress, err := dbusutil.NewSignalWatcherForSystemBus(ctx, ciceroneDBusMatchSpec("UninstallPackageProgress"))
	if err != nil {
		return err
	}
	// Always close the UninstallPackageProgress watcher regardless of success.
	defer progress.Close(ctx)

	resp := &cpb.UninstallPackageOwningFileResponse{}
	if err = dbusutil.CallProtoMethod(ctx, c.ciceroneObj, ciceroneInterface+".UninstallPackageOwningFile",
		&cpb.UninstallPackageOwningFileRequest{
			VmName:        c.VM.name,
			ContainerName: c.containerName,
			OwnerId:       c.VM.Concierge.ownerID,
			DesktopFileId: desktopFileID,
		}, resp); err != nil {
		return err
	}

	if resp.Status != cpb.UninstallPackageOwningFileResponse_STARTED {
		return errors.Errorf("failed to start package uninstall: %v", resp.FailureReason)
	}

	// Wait for the signal for uninstall completion which will signify success or
	// failure.
	testing.ContextLog(ctx, "Waiting for UninstallPackageProgress D-Bus signal")
	sigResult := &cpb.UninstallPackageProgressSignal{}
	for {
		if err := waitForDBusSignal(ctx, progress, nil, sigResult); err != nil {
			return err
		}
		if sigResult.VmName == c.VM.name && sigResult.ContainerName == c.containerName &&
			sigResult.OwnerId == c.VM.Concierge.ownerID {
			if sigResult.Status == cpb.UninstallPackageProgressSignal_SUCCEEDED {
				return nil
			}
			if sigResult.Status == cpb.UninstallPackageProgressSignal_FAILED {
				return errors.Errorf("failure with package uninstall: %v", sigResult.FailureDetails)
			}
		}
	}
}

// containerCommand returns a testexec.Cmd with a vsh command that will run in
// the specified container.
func containerCommand(ctx context.Context, vmName, containerName, ownerID string, vshArgs ...string) *testexec.Cmd {
	args := append([]string{"--vm_name=" + vmName,
		"--target_container=" + containerName,
		"--owner_id=" + ownerID,
		"--"},
		vshArgs...)
	cmd := testexec.CommandContext(ctx, "vsh", args...)
	// Add a dummy buffer for stdin to force allocating a pipe. vsh uses
	// epoll internally and generates a warning (EPERM) if stdin is /dev/null.
	cmd.Stdin = &bytes.Buffer{}
	return cmd
}

// DefaultContainerCommand returns a testexec.Cmd with a vsh command that will run in
// the default termina/penguin container.
func DefaultContainerCommand(ctx context.Context, ownerID string, vshArgs ...string) *testexec.Cmd {
	return containerCommand(ctx, DefaultVMName, DefaultContainerName, ownerID, vshArgs...)
}

// Command returns a testexec.Cmd with a vsh command that will run in this
// container.
func (c *Container) Command(ctx context.Context, vshArgs ...string) *testexec.Cmd {
	return containerCommand(ctx, c.VM.name, c.containerName, c.VM.Concierge.ownerID, vshArgs...)
}

// DumpLog dumps the logs from the container to a local output file named
// container_log.txt in dir (typically the test's output dir).
// It does this by executing journalctl in the container and grabbing the output.
func (c *Container) DumpLog(ctx context.Context, dir string) error {
	f, err := os.Create(filepath.Join(dir, "container_log.txt"))
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO(jkardatzke): Remove stripping off the color codes that show up in
	// journalctl once crbug.com/888102 is fixed.
	cmd := c.Command(ctx, "sh", "-c",
		"sudo journalctl --no-pager | tr -cd '[:space:][:print:]'")
	cmd.Stdout = f
	return cmd.Run()
}

// CreateDefaultContainer prepares a VM and container with default settings and
// either the live or staging container versions. The directory dir may be used
// to store logs on failure.
func CreateDefaultContainer(ctx context.Context, dir, user string, t ContainerType) (*Container, error) {
	concierge, err := NewConcierge(ctx, user)
	if err != nil {
		return nil, err
	}

	vmInstance := NewDefaultVM(concierge)

	if err := vmInstance.Start(ctx); err != nil {
		return nil, err
	}

	created, err := dbusutil.NewSignalWatcherForSystemBus(ctx, ciceroneDBusMatchSpec("LxdContainerCreated"))
	if err != nil {
		return nil, err
	}
	// Always close the InstallLinuxPackageProgress watcher regardless of success.
	defer created.Close(ctx)

	c := DefaultContainer(vmInstance)
	if err := c.Create(ctx, t); err != nil {
		return nil, err
	}
	// Container is being created, wait for signal.
	createdSig := &cpb.LxdContainerCreatedSignal{}
	testing.ContextLogf(ctx, "Waiting for LxdContainerCreated signal for container %q, VM %q", c.containerName, vmInstance.name)
	if err := waitForDBusSignal(ctx, created, nil, createdSig); err != nil {
		return nil, errors.Wrap(err, "failed to get LxdContainerCreatedSignal")
	}
	if createdSig.GetVmName() != vmInstance.name {
		return nil, errors.Errorf("unexpected container creation signal for VM %q", createdSig.GetVmName())
	} else if createdSig.GetContainerName() != c.containerName {
		return nil, errors.Errorf("unexpected container creation signal for container %q", createdSig.GetContainerName())
	}
	if createdSig.GetStatus() != cpb.LxdContainerCreatedSignal_CREATED {
		return nil, errors.Errorf("failed to create container: status: %d reason: %v", createdSig.GetStatus(), createdSig.GetFailureReason())
	}

	if err := c.StartAndWait(ctx, dir); err != nil {
		return nil, err
	}

	return c, nil
}

func ciceroneDBusMatchSpec(memberName string) dbusutil.MatchSpec {
	return dbusutil.MatchSpec{
		Type:      "signal",
		Path:      ciceronePath,
		Interface: ciceroneInterface,
		Member:    memberName,
	}
}

// ContainerCreationWatcher is a wrapper of SignalWatcher to trace container creation progress.
type ContainerCreationWatcher struct {
	cont    *Container
	watcher *dbusutil.SignalWatcher
}

// NewContainerCreationWatcher returns a ContainerCreationWatcher.
func NewContainerCreationWatcher(ctx context.Context, cont *Container) (*ContainerCreationWatcher, error) {
	watcher, err := dbusutil.NewSignalWatcherForSystemBus(ctx,
		ciceroneDBusMatchSpec("LxdContainerDownloading"), ciceroneDBusMatchSpec("LxdContainerCreated"))
	if err != nil {
		return nil, err
	}
	return &ContainerCreationWatcher{cont, watcher}, nil
}

// Close cleans up the SignalWatcher.
func (c *ContainerCreationWatcher) Close(ctx context.Context) {
	c.watcher.Close(ctx)
}

// isWatchingContainer returns whether the signal is for the container we are watching.
func (c *ContainerCreationWatcher) isWatchingContainer(vmName, containerName, ownerID string) bool {
	return vmName == c.cont.VM.name && containerName == c.cont.containerName && ownerID == c.cont.VM.Concierge.ownerID
}

// WaitForDownload waits for cicerone to send a container download notification.
// If pct is negative, this method returns after the next notification is received.
// Otherwise, it returns only after a notification with percent pct in [0, 100] is received.
// An error is returned if ctx's deadline is reached.
func (c *ContainerCreationWatcher) WaitForDownload(ctx context.Context, pct int32) error {
	spec := ciceroneDBusMatchSpec("LxdContainerDownloading")
	sigResult := &cpb.LxdContainerDownloadingSignal{}
	for {
		if err := waitForDBusSignal(ctx, c.watcher, &spec, sigResult); err != nil {
			return err
		}
		if c.isWatchingContainer(sigResult.VmName, sigResult.ContainerName, sigResult.OwnerId) {
			if pct < 0 || sigResult.DownloadProgress == pct {
				return nil
			}
		}
	}
}

// WaitForCreationComplete waits for the container to be created.
func (c *ContainerCreationWatcher) WaitForCreationComplete(ctx context.Context) error {
	spec := ciceroneDBusMatchSpec("LxdContainerCreated")
	sigResult := &cpb.LxdContainerCreatedSignal{}
	for {
		if err := waitForDBusSignal(ctx, c.watcher, &spec, sigResult); err != nil {
			return err
		}
		if c.isWatchingContainer(sigResult.VmName, sigResult.ContainerName, sigResult.OwnerId) {
			if sigResult.GetStatus() == cpb.LxdContainerCreatedSignal_CREATED {
				return nil
			}
		}
	}
}
