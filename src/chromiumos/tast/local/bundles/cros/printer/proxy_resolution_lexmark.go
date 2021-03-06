// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package printer

import (
	"context"

	"chromiumos/tast/local/bundles/cros/printer/proxyippprint"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: ProxyResolutionLexmark,
		Desc: "Verifies that Lexmark printers add the appropriate options for the IPP printer-resolution attribute",
		Contacts: []string{
			"batrapranav@chromium.org",
			"cros-printing-dev@chromium.org",
		},
		SoftwareDeps: []string{"chrome", "cros_internal", "cups", "plugin_vm"},
		Data: []string{
			"to_print.pdf",
			"printer_Lexmark.ppd",
		},
		Attr: []string{"group:mainline", "informational"},
		Pre:  chrome.LoggedIn(),
		Params: []testing.Param{{
			Name: "600dpi",
			Val: &proxyippprint.Params{
				PpdFile:      "printer_Lexmark.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_resolution_lexmark_600dpi_golden.ps",
				Options:      []proxyippprint.Option{proxyippprint.WithResolution("600dpi")},
			},
			ExtraData: []string{"printer_resolution_lexmark_600dpi_golden.ps"},
		}, {
			Name: "1200dpi",
			Val: &proxyippprint.Params{
				PpdFile:      "printer_Lexmark.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_resolution_lexmark_1200dpi_golden.ps",
				Options:      []proxyippprint.Option{proxyippprint.WithResolution("1200dpi")},
			},
			ExtraData: []string{"printer_resolution_lexmark_1200dpi_golden.ps"},
		}, {
			Name: "2400x600dpi",
			Val: &proxyippprint.Params{
				PpdFile:      "printer_Lexmark.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_resolution_lexmark_2400x600dpi_golden.ps",
				Options:      []proxyippprint.Option{proxyippprint.WithResolution("2400x600dpi")},
			},
			ExtraData: []string{"printer_resolution_lexmark_2400x600dpi_golden.ps"},
		}},
	})
}

func ProxyResolutionLexmark(ctx context.Context, s *testing.State) {
	testOpt := s.Param().(*proxyippprint.Params)

	proxyippprint.Run(ctx, s, testOpt)
}
