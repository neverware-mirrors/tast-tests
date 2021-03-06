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
		Func: ShowLogoutButtonInTray,
		Desc: "Behavior of ShowLogoutButtonInTray policy, check if a logout button is shown based on the value of the policy",
		Contacts: []string{
			"alexanderhartl@google.com", // Test author
			"chromeos-commercial-stability@google.com",
		},
		SoftwareDeps: []string{"chrome"},
		Attr:         []string{"group:mainline"},
		Pre:          pre.User,
	})
}

func ShowLogoutButtonInTray(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*pre.PreData).Chrome
	fdms := s.PreValue().(*pre.PreData).FakeDMS

	// Connect to Test API to use it with the UI library.
	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create Test API connection: ", err)
	}

	for _, param := range []struct {
		name       string
		wantButton bool                           // wantButton is the expected existence of the "Sign out" button.
		policy     *policy.ShowLogoutButtonInTray // policy is the policy we test.
	}{
		{
			name:       "unset",
			wantButton: false,
			policy:     &policy.ShowLogoutButtonInTray{Stat: policy.StatusUnset},
		},
		{
			name:       "don't show",
			wantButton: false,
			policy:     &policy.ShowLogoutButtonInTray{Val: false},
		},
		{
			name:       "show",
			wantButton: true,
			policy:     &policy.ShowLogoutButtonInTray{Val: true},
		},
	} {
		s.Run(ctx, param.name, func(ctx context.Context, s *testing.State) {
			defer faillog.DumpUITreeOnErrorToFile(ctx, s.OutDir(), s.HasError, tconn, "ui_tree_"+param.name+".txt")

			// Perform cleanup.
			if err := policyutil.ResetChrome(ctx, fdms, cr); err != nil {
				s.Fatal("Failed to clean up: ", err)
			}

			// Update policies.
			if err := policyutil.ServeAndVerify(ctx, fdms, cr, []policy.Policy{param.policy}); err != nil {
				s.Fatal("Failed to update policies: ", err)
			}

			// Confirm the status of the Sign out button node.
			if err := ui.WaitUntilExistsStatus(ctx, tconn, ui.FindParams{
				Role: ui.RoleTypeButton,
				Name: "Sign out",
			}, param.wantButton, 15*time.Second); err != nil {
				s.Error("Could not confirm the desired status of the Sign out button: ", err)
			}
		})
	}
}
