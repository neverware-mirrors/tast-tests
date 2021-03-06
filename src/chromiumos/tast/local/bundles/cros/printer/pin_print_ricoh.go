// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package printer

import (
	"context"

	"chromiumos/tast/local/bundles/cros/printer/ippprint"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: PinPrintRicoh,
		Desc: "Verifies that printers with Ricoh Pin printing support add the appropriate options for a variety of attributes",
		Contacts: []string{
			"bmalcolm@chromium.org",
			"cros-printing-dev@chromium.org",
		},
		SoftwareDeps: []string{"cros_internal", "cups"},
		Data: []string{
			"to_print.pdf",
			"printer_Ricoh_JobPassword.ppd",
			"printer_Ricoh_LockedPrintPassword.ppd",
			"printer_Ricoh_password.ppd",
		},
		Attr: []string{"group:mainline"},
		Params: []testing.Param{{
			Name: "jobpassword_no_pin",
			Val: &ippprint.Params{
				PpdFile:      "printer_Ricoh_JobPassword.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_pin_print_ricoh_JobPassword_no_pin_golden.ps",
			},
			ExtraData: []string{"printer_pin_print_ricoh_JobPassword_no_pin_golden.ps"},
		}, {
			Name: "jobpassword_pin",
			Val: &ippprint.Params{
				PpdFile:      "printer_Ricoh_JobPassword.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_pin_print_ricoh_JobPassword_pin_golden.ps",
				Options:      []ippprint.Option{ippprint.WithJobPassword("1234")},
			},
			ExtraData: []string{"printer_pin_print_ricoh_JobPassword_pin_golden.ps"},
		}, {
			Name: "lockedprintpassword_no_pin",
			Val: &ippprint.Params{
				PpdFile:      "printer_Ricoh_LockedPrintPassword.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_pin_print_ricoh_LockedPrintPassword_no_pin_golden.ps",
			},
			ExtraData: []string{"printer_pin_print_ricoh_LockedPrintPassword_no_pin_golden.ps"},
		}, {
			Name: "lockedprintpassword_pin",
			Val: &ippprint.Params{
				PpdFile:      "printer_Ricoh_LockedPrintPassword.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_pin_print_ricoh_LockedPrintPassword_pin_golden.ps",
				Options:      []ippprint.Option{ippprint.WithJobPassword("1234")},
			},
			ExtraData: []string{"printer_pin_print_ricoh_LockedPrintPassword_pin_golden.ps"},
		}, {
			Name: "password_no_pin",
			Val: &ippprint.Params{
				PpdFile:      "printer_Ricoh_password.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_pin_print_ricoh_password_no_pin_golden.ps",
			},
			ExtraData: []string{"printer_pin_print_ricoh_password_no_pin_golden.ps"},
		}, {
			Name: "password_pin",
			Val: &ippprint.Params{
				PpdFile:      "printer_Ricoh_password.ppd",
				PrintFile:    "to_print.pdf",
				ExpectedFile: "printer_pin_print_ricoh_password_pin_golden.ps",
				Options:      []ippprint.Option{ippprint.WithJobPassword("1234")},
			},
			ExtraData: []string{"printer_pin_print_ricoh_password_pin_golden.ps"},
		}},
	})
}

func PinPrintRicoh(ctx context.Context, s *testing.State) {
	testOpt := s.Param().(*ippprint.Params)

	ippprint.Run(ctx, s, testOpt)
}
