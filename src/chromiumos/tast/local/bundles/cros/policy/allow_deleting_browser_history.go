// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package policy

import (
	"context"
	"strconv"
	"time"

	"chromiumos/tast/common/policy"
	"chromiumos/tast/local/bundles/cros/policy/pre"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/policyutil"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func: AllowDeletingBrowserHistory,
		Desc: "Behavior of AllowDeletingBrowserHistory policy, checking the correspoding checkbox states (restriction and checked) after setting the policy",
		Contacts: []string{
			"gabormagda@google.com", // Test author
			"chromeos-commercial-stability@google.com",
		},
		SoftwareDeps: []string{"chrome"},
		Attr:         []string{"group:mainline", "informational"},
		Pre:          pre.User,
	})
}

// AllowDeletingBrowserHistory tests the AllowDeletingBrowserHistory policy.
func AllowDeletingBrowserHistory(ctx context.Context, s *testing.State) {
	cr := s.PreValue().(*pre.PreData).Chrome
	fdms := s.PreValue().(*pre.PreData).FakeDMS

	for _, param := range []struct {
		name           string
		wantRestricted bool                                // wantRestricted is the wanted restriction state of the checkboxes in Browsing history.
		wantChecked    bool                                // wantChecked is the wanted checked state of the checkboxes in Browsing history.
		value          *policy.AllowDeletingBrowserHistory // value is the value of the policy.
	}{
		{
			name:           "unset",
			wantRestricted: false,
			wantChecked:    true,
			value:          &policy.AllowDeletingBrowserHistory{Stat: policy.StatusUnset},
		},
		{
			name:           "allow",
			wantRestricted: false,
			wantChecked:    true,
			value:          &policy.AllowDeletingBrowserHistory{Val: true},
		},
		{
			name:           "deny",
			wantRestricted: true,
			wantChecked:    false,
			value:          &policy.AllowDeletingBrowserHistory{Val: false},
		},
	} {
		s.Run(ctx, param.name, func(ctx context.Context, s *testing.State) {
			// Perform cleanup.
			if err := policyutil.ResetChrome(ctx, fdms, cr); err != nil {
				s.Fatal("Failed to clean up: ", err)
			}

			// Update policies.
			if err := policyutil.ServeAndRefresh(ctx, fdms, cr, []policy.Policy{param.value}); err != nil {
				s.Fatal("Failed to update policies: ", err)
			}

			// Connect to Test API to use it with the ui library.
			tconn, err := cr.TestAPIConn(ctx)
			if err != nil {
				s.Fatal("Failed to create Test API connection: ", err)
			}

			// Open settings page where the affected checkboxes can be found.
			conn, err := cr.NewConn(ctx, "chrome://settings/clearBrowserData")
			if err != nil {
				s.Fatal("Failed to connect to the settings page: ", err)
			}
			defer conn.Close()

			// Loop for different checkboxes.
			for _, cb := range []struct {
				ref  string // ref is the shortened name of the checkbox that can be used in logging.
				name string // name is a unique part of the checkbox name in the UI tree.
				tab  string // tab is the name of the tab in the UI tree that should be selected to find the checkbox.
			}{
				{
					ref:  "Browsing history",
					name: "Browsing history Clears history and autocompletions in the address bar.",
					tab:  "Basic",
				},
				{
					ref:  "Browsing history",
					name: "Browsing history None",
					tab:  "Advanced",
				},
				{
					ref:  "Download history",
					name: "Download history None",
					tab:  "Advanced",
				},
			} {
				// Select the tab if it is not selected already.
				tabNode, err := ui.FindWithTimeout(ctx, tconn, ui.FindParams{
					Role: ui.RoleTypeTab,
					Name: cb.tab,
				}, 15*time.Second)
				if err != nil {
					s.Fatalf("Finding %s tab failed: %v", cb.tab, err)
				}
				defer tabNode.Release(ctx)

				if tabNode.ClassName != "tab selected" {
					if err := tabNode.LeftClick(ctx); err != nil {
						s.Fatalf("Failed to click on %s tab: %v", cb.tab, err)
					}

					if err := ui.WaitUntilExists(ctx, tconn, ui.FindParams{
						Role:      ui.RoleTypeTab,
						Name:      cb.tab,
						ClassName: "tab selected",
					}, 15*time.Second); err != nil {
						s.Fatalf("Waiting for %s tab failed: %v", cb.tab, err)
					}
				}

				// Find the checkbox node.
				cbNode, err := ui.FindWithTimeout(ctx, tconn, ui.FindParams{
					Role: ui.RoleTypeCheckBox,
					Name: cb.name,
				}, 15*time.Second)
				if err != nil {
					s.Fatalf("Finding %q checkbox failed: %v", cb.ref, err)
				}
				defer cbNode.Release(ctx)

				// Check the checked state of the checkbox.
				if checked, err := cbNode.Attribute(ctx, "checked"); err != nil {
					s.Fatalf("Failed to get the checked attribute of the %q checkbox in the %s tab: %v", cb.ref, cb.tab, err)
				} else if checkedStr, ok := checked.(string); !ok {
					s.Fatalf("The checked attribute of the %q checkbox is not a string: %v", cb.ref, checkedStr)
				} else if checkedStr != strconv.FormatBool(param.wantChecked) {
					s.Errorf("Unexpected %q checkbox checked state in the %s tab: got %s; want %t", cb.ref, cb.tab, checkedStr, param.wantChecked)
				}

				// Check the restriction setting of the checkbox.
				if restriction, err := cbNode.Attribute(ctx, "restriction"); err != nil {
					// The restriction attribute is optional. If it does not exists, the checkbox is not restricted.
					if param.wantRestricted {
						s.Errorf("Unexpected %q checkbox restriction in the %s tab: got false; want %t", cb.ref, cb.tab, param.wantRestricted)
					}
				} else if restrictionStr, ok := restriction.(string); !ok {
					s.Fatalf("The restriction attribute of the %q checkbox is not a string: %v", cb.ref, restriction)
				} else if restricted := (restrictionStr == "disabled" || restrictionStr == "readOnly"); restricted != param.wantRestricted {
					s.Log("The restriction attribute is: ", restrictionStr)
					s.Errorf("Unexpected %q checkbox restriction in the %s tab: got %t; want %t", cb.ref, cb.tab, restricted, param.wantRestricted)
				}
			}
		})
	}
}