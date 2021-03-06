// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package network

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/bundles/cros/network/stringset"
	"chromiumos/tast/local/shill"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/local/upstart"
	"chromiumos/tast/testing"
	"chromiumos/tast/testing/hwdep"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         UdevRename,
		Desc:         "Verifies that network interfaces remain intact after udev restart and WiFi driver rebind",
		Contacts:     []string{"yenlinlai@google.com", "chromeos-kernel-wifi@google.com"},
		Attr:         []string{"group:mainline"},
		SoftwareDeps: []string{"wifi", "shill-wifi"},
		// TODO(b/149247291): remove the dependency once elm/hana upreved kernel to 4.19 or above.
		// TODO(crbug.com/1115620): remove "Elm" and "Hana" after unibuild migration completed.
		HardwareDeps: hwdep.D(hwdep.SkipOnPlatform("oak"), hwdep.SkipOnPlatform("elm"), hwdep.SkipOnPlatform("hana")),
	})
}

func restartWifiInterface(ctx context.Context) error {
	manager, err := shill.NewManager(ctx)
	if err != nil {
		return errors.Wrap(err, "failed creating shill manager proxy")
	}

	iface, err := shill.WifiInterface(ctx, manager, 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "could not find interface")
	}

	devicePath := fmt.Sprintf("/sys/class/net/%s/device", iface)
	deviceRealPath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return errors.Wrapf(err, "could not evaluate symlink on payload %s", devicePath)
	}

	// The driver path is the directory where we can bind and release the device.
	driverPath := filepath.Join(devicePath, "driver")
	driverRealPath, err := filepath.EvalSymlinks(driverPath)
	if err != nil {
		return errors.Wrapf(err, "could not evaluate symlink on path %s", driverPath)
	}

	// Function to find device paths for the brcmfmac (Broadcom FullMAC) driver.
	// In general, one device is associated with a driver. However, for brcmfmac driver,
	// it associates with two devices. We have to unbind/bind both.
	brcmfmacDevicePaths := func(driverPath string) ([]string, error) {
		paths, err := filepath.Glob(filepath.Join(driverPath, "*"))
		if err != nil {
			return nil, err
		}
		if len(paths) <= 1 {
			return nil, errors.Errorf("found %d brcmfmac driver devices, expected at least 2", len(paths))
		}

		var ret []string
		for _, p := range paths {
			// Only consider links to devices, and not paths like '/sys/bus/.../unbind'.
			if rp, err := filepath.EvalSymlinks(p); err == nil && strings.HasPrefix(rp, "/sys/devices") {
				ret = append(ret, rp)
			}
		}
		return ret, nil
	}

	devPaths := []string{deviceRealPath}
	// Special case for brcmfmac (Broadcom FullMAC) driver.
	// Note that in older kernels, e.g. 3.14, the driver name of Broadcom FullMAC is "brcmfmac_sdio";
	// however, in recent kernels, e.g. 4.19, it is named as "brcmfmac". So we use prefix match here.
	if strings.HasPrefix(filepath.Base(driverRealPath), "brcmfmac") {
		devPaths, err = brcmfmacDevicePaths(driverPath)
		if err != nil {
			errors.Wrap(err, "brcmfmac device paths error")
		}
		testing.ContextLog(ctx, "Devices associated with brcmfmac driver: ", devPaths)
	}

	for _, devPath := range devPaths {
		testing.ContextLogf(ctx, "Rebind device %s to driver %s", devPath, driverRealPath)
		devName := filepath.Base(devPath)
		if err := ioutil.WriteFile(filepath.Join(driverRealPath, "unbind"), []byte(devName), 0200); err != nil {
			return errors.Wrapf(err, "could not unbind %s driver", iface)
		}
		if err := ioutil.WriteFile(filepath.Join(driverRealPath, "bind"), []byte(devName), 0200); err != nil {
			return errors.Wrapf(err, "could not bind %s driver", iface)
		}
	}
	return nil
}

// udevEventMonitor waits until any udev event is emitted or error when timeout reached.
// A channel for the caller to wait for result is returned.
func udevEventMonitor(ctx context.Context, timeout time.Duration) <-chan error {
	ret := make(chan error, 1)
	scanDone := make(chan struct{})

	// Spawn udevadm monitor.
	cmd := testexec.CommandContext(ctx, "udevadm", "monitor", "-u")
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		ret <- errors.Wrap(err, "failed to get stdout reader")
		return ret
	}
	if err := cmd.Start(); err != nil {
		ret <- errors.Wrap(err, "failed to spawn \"udevadm monitor\"")
		return ret
	}

	// Spawn reader routine.
	go func() {
		defer close(scanDone)
		scanner := bufio.NewScanner(cmdOut)
		for scanner.Scan() {
			line := scanner.Text()
			// Check if it's a udev event by the line prefix.
			if strings.HasPrefix(line, "UDEV  [") {
				ret <- nil
				return
			}
		}
		ret <- errors.New("udev event not captured")
	}()

	// Spawn watchdog for "udevadm monitor" which terminates and cleans up
	// the bg process. When we reach timeout, try to send SIGTERM first so
	// that if the process has stdout buffer, it can still flush it. If we
	// finish the scan, we kill and wait to clean up. If the process reaches
	// hard deadline (ctx.Done) and kill-ed by cmd package, we can just do
	// the same as previous case.
	go func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			if err := cmd.Signal(syscall.SIGTERM); err != nil {
				testing.ContextLog(ctx, "Failed to terminate udevadm monitor")
			}
		case <-scanDone:
		}

		<-scanDone
		if err := cmd.Kill(); err != nil {
			testing.ContextLog(ctx, "Failed to kill udevadm monitor")
		}
		cmd.Wait()
	}()

	return ret
}

func restartUdev(ctx context.Context) error {
	const service = "udev"
	if _, state, _, err := upstart.JobStatus(ctx, service); err != nil {
		return errors.Wrapf(err, "could not query status of service %s", service)
	} else if state != upstart.RunningState {
		return errors.Errorf("%s not running", service)
	}

	if err := upstart.StopJob(ctx, service); err != nil {
		return errors.Errorf("%s failed to stop", service)
	}

	// Make sure udev finished its job and stopped.
	testexec.CommandContext(ctx, "udevadm", "settle").Run()

	// TODO(yenlinlai): Currently we don't yet have a good way to wait from restarting
	// udev until having all rules processed. "udevadm settle" may not properly wait if
	// udev has not gotten into event processing loop. Some examples can be found in
	// crrev.com/c/1725184.
	// Our current work-around is to watch the first output of "udevadm monitor -u" as
	// the ready signal. However, there's still some possible race if udev finishes all
	// update before "udevadm monitor" starts. In this case, it may not catch any event
	// but wait until timeout error.

	// Spawn udevadm monitor, continue when error cause we want to start udev.
	done := udevEventMonitor(ctx, 10*time.Second)

	if err := upstart.StartJob(ctx, service); err != nil {
		return errors.Errorf("%s failed to start", service)
	}

	return <-done
}

// deviceRestarter is a function type that defines a first class function that would restart
// a device or series of devices. restartUdev() and restartWifiInterface() match the
// function prototype.
type deviceRestarter func(ctx context.Context) error

func interfaceNames() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(ifaces))
	for i := range ifaces {
		names[i] = ifaces[i].Name
	}
	sort.Strings(names)
	return names, nil
}

// expectIface expects actual interfaces is the same as expected.
func expectIface(expect, actual []string) error {
	es := stringset.New(expect)
	as := stringset.New(actual)
	if es.Equal(as) {
		return nil
	}
	var errs []string

	// wanted: interfaces in expect not in actual.
	if wanted := es.Diff(as); len(wanted) > 0 {
		errs = append(errs, fmt.Sprintf("wanted:%v", wanted.Elements()))
	}
	// unexpected: interfaces in actual not in expect.
	if unexpected := as.Diff(es); len(unexpected) > 0 {
		errs = append(errs, fmt.Sprintf("unexpected:%v", unexpected.Elements()))
	}
	// matched: interfaces in both actual and expect.
	if matched := es.Intersect(as); len(matched) > 0 {
		errs = append(errs, fmt.Sprintf("matched:%v", matched.Elements()))
	}
	return errors.New("failed expecting network interfaces: " + strings.Join(errs, ", "))
}

func testUdevDeviceList(ctx context.Context, fn deviceRestarter) error {
	iflistPre, err := interfaceNames()
	if err != nil {
		return err
	}
	if err := fn(ctx); err != nil {
		return err
	}

	// Wait for event processing.
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()
	if err := testexec.CommandContext(timeoutCtx, "udevadm", "settle").Run(testexec.DumpLogOnError); err != nil {
		return errors.Wrap(err, "device could not settle in time after restart")
	}

	if err := testing.Poll(ctx, func(ctx context.Context) error {
		iflistPost, err := interfaceNames()
		if err != nil {
			return err
		}
		if err := expectIface(iflistPre, iflistPost); err != nil {
			return err
		}
		return nil
	}, &testing.PollOptions{Timeout: 10 * time.Second}); err != nil {
		return err
	}
	return nil
}

func UdevRename(ctx context.Context, s *testing.State) {
	if err := testUdevDeviceList(ctx, restartUdev); err != nil {
		s.Error("Restarting udev: ", err)
	}

	if err := testUdevDeviceList(ctx, restartWifiInterface); err != nil {
		s.Error("Restarting wireless interface: ", err)
	}
}
