// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package printer

import (
	"context"

	"chromiumos/tast/local/bundles/cros/printer/usbprintertests"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         IPPUSBPPDNoCopies,
		Desc:         "Verifies that the 'copies-supported' attribute of the printer is used to populate the cupsManualCopies and cupsMaxCopies values in the corresponding generated PPD",
		Contacts:     []string{"valleau@chromium.org"},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome", "cups"},
		Data:         []string{"ippusb_no_copies.json"},
		Pre:          chrome.LoggedIn(),
	})
}

// IPPUSBPPDNoCopies tests that the "cupsManualCopies" and "cupsMaxCopies" PPD
// fields will be correctly populated when configuring an IPP-over-USB printer
// which does not provide a value for the "copies-supported" attribute.
func IPPUSBPPDNoCopies(ctx context.Context, s *testing.State) {
	const descriptors = "/usr/local/etc/virtual-usb-printer/ippusb_printer.json"
	usbprintertests.RunIPPUSBPPDTest(ctx, s, descriptors, s.DataPath("ippusb_no_copies.json"), map[string]string{
		"*cupsManualCopies": "True",
		"*cupsMaxCopies":    "1",
	})
}
