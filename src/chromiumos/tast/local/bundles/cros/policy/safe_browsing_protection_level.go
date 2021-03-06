// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package policy

import (
	"context"
	"time"

	"chromiumos/tast/common/policy"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/chrome/ui/faillog"
	"chromiumos/tast/local/policyutil"
	"chromiumos/tast/local/policyutil/pre"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: SafeBrowsingProtectionLevel,
		Desc: "Checks if Google Chrome's Safe Browsing feature is enabled and the mode it operates in",
		Contacts: []string{
			"mohamedaomar@google.com", // Test author
			"chromeos-commercial-stability@google.com",
		},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome"},
		Pre:          pre.User,
	})
}

func SafeBrowsingProtectionLevel(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*pre.PreData).Chrome
	fdms := s.PreValue().(*pre.PreData).FakeDMS

	// Connect to Test API to use it with the UI library.
	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create Test API connection: ", err)
	}

	for _, param := range []struct {
		name           string
		wantRestricted bool   // wantRestricted is the wanted restriction state of the checkboxes in Safe Browsing settings page.
		selectedOption string // selectedOption is the selected safety level in Safe Browsing settings page.
		value          *policy.SafeBrowsingProtectionLevel
	}{
		{
			name:           "unset",
			wantRestricted: false,
			selectedOption: "Standard protection",
			value:          &policy.SafeBrowsingProtectionLevel{Stat: policy.StatusUnset},
		},
		{
			name:           "no_protection",
			wantRestricted: true,
			selectedOption: "No protection (not recommended)",
			value:          &policy.SafeBrowsingProtectionLevel{Val: 0},
		},
		{
			name:           "standard_protection",
			wantRestricted: true,
			selectedOption: "Standard protection",
			value:          &policy.SafeBrowsingProtectionLevel{Val: 1},
		},
		{
			name:           "enhanced_protection",
			wantRestricted: true,
			selectedOption: "Enhanced protection",
			value:          &policy.SafeBrowsingProtectionLevel{Val: 2},
		},
	} {
		s.Run(ctx, param.name, func(ctx context.Context, s *testing.State) {
			defer faillog.DumpUITreeOnErrorToFile(ctx, s.OutDir(), s.HasError, tconn, "ui_tree_"+param.name+".txt")

			// Perform cleanup.
			if err := policyutil.ResetChrome(ctx, fdms, cr); err != nil {
				s.Fatal("Failed to clean up: ", err)
			}

			// Update policies.
			if err := policyutil.ServeAndVerify(ctx, fdms, cr, []policy.Policy{param.value}); err != nil {
				s.Fatal("Failed to update policies: ", err)
			}

			// Open the security settings page where the affected radio buttons can be found.
			conn, err := cr.NewConn(ctx, "chrome://settings/security")
			if err != nil {
				s.Fatal("Failed to connect to the settings page: ", err)
			}
			defer conn.Close()

			// Find the radio group node.
			rgNode, err := ui.FindWithTimeout(ctx, tconn, ui.FindParams{Role: ui.RoleTypeRadioGroup}, 15*time.Second)
			if err != nil {
				s.Fatal("Finding radio group failed: ", err)
			}
			defer rgNode.Release(ctx)

			// Find the selected radio button under the radio group.
			srbNode, err := rgNode.FindSelectedRadioButton(ctx)
			if err != nil {
				s.Fatal("Finding the selected radio button failed: ", err)
			}
			defer srbNode.Release(ctx)

			if isRestricted, err := srbNode.Matches(ctx, ui.FindParams{Attributes: map[string]interface{}{"restriction": ui.RestrictionDisabled}}); err != nil {
				s.Fatal("Checking restriction attribute failed: ", err)
			} else if isRestricted != param.wantRestricted {
				s.Errorf("Failed to verify restriction behavior; got %t, want %t", isRestricted, param.wantRestricted)
			}

			if srbNode.Name != param.selectedOption {
				s.Errorf("Unexpected selected safety level; got %q, want %q", srbNode.Name, param.selectedOption)
			}
		})
	}
}
