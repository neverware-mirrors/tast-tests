// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package input

import (
	"context"
	"math/big"
	"os"
)

// GamepadEventWriter supports injecting events into a virtual gamepad device.
type GamepadEventWriter struct {
	rw   *RawEventWriter
	virt *os.File // if non-nil, used to hold a virtual device open
	dev  string   // path to underlying device in /dev/input
}

// Gamepad creates a virtual gamepad device and returns and EventWriter that injects events into it.
func Gamepad(ctx context.Context) (*GamepadEventWriter, error) {
	gw := &GamepadEventWriter{}
	const usbBus = 0x3 // BUS_USB from input.h
	var err error
	if gw.dev, gw.virt, err = createVirtual(
		gw.DeviceName(),
		devID{usbBus, gw.VendorID(), gw.ProductID(), 0x0111}, 0, 0x1b,
		map[EventType]*big.Int{
			EV_KEY: makeBigInt([]uint64{0x3fff000000000000, 0, 0, 0, 0}),
			EV_ABS: big.NewInt(0x26081000003003f),
			EV_MSC: big.NewInt(0x10)}); err != nil {
		return nil, err
	}

	if gw.rw, err = Device(ctx, gw.dev); err != nil {
		gw.Close()
		return nil, err
	}

	return gw, nil
}

// Close closes the gamepad device.
func (gw *GamepadEventWriter) Close() error {
	var firstErr error
	if gw.rw != nil {
		firstErr = gw.rw.Close()
	}
	if gw.virt != nil {
		if err := gw.virt.Close(); firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Device returns the path of the underlying device, e.g. "/dev/input/event3".
func (gw *GamepadEventWriter) Device() string { return gw.dev }

// VendorID returns the vendor ID of the virtual gamepad device.
func (gw *GamepadEventWriter) VendorID() uint16 { return 0x054c }

// ProductID returns the product ID of the virtual gamepad device.
func (gw *GamepadEventWriter) ProductID() uint16 { return 0x09cc }

// DeviceName returns the device name of the virtual gamepad device.
func (gw *GamepadEventWriter) DeviceName() string {
	return "Wireless Controller"
}

// sendKey writes a EV_KEY event containing the specified code and value, followed by a EV_SYN event.
func (gw *GamepadEventWriter) sendKey(ctx context.Context, ec EventCode, val int32) error {
	if err := gw.rw.Event(EV_KEY, ec, val); err != nil {
		return err
	}
	return gw.rw.Sync()
}

// TapButton presses and releases the gamepad button specified by EventCode.
func (gw *GamepadEventWriter) TapButton(ctx context.Context, ec EventCode) error {
	if err := gw.sendKey(ctx, ec, 1); err != nil {
		return err
	}
	return gw.sendKey(ctx, ec, 0)
}
