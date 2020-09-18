// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package example

import (
	"context"

	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

type manyParamsParams struct {
	Options []chrome.Option
	Expr    string
}

func init() {
	testing.AddTest(&testing.Test{
		Func:     ManyParams,
		Desc:     "Example to generate many test parameters automatically",
		Contacts: []string{"tast-owners@google.com"},
		Params: []testing.Param{
			// Parameters generated by many_params_test.go. DO NOT EDIT.
			{
				Name:              "noarc_url",
				ExtraSoftwareDeps: []string{"chrome"},
				Val: manyParamsParams{
					Expr: "location.href",
				},
			},
			{
				Name:              "noarc_state",
				ExtraSoftwareDeps: []string{"chrome"},
				Val: manyParamsParams{
					Expr: "document.readyState",
				},
			},
			{
				Name:              "arc_url",
				ExtraSoftwareDeps: []string{"chrome", "android"},
				Val: manyParamsParams{
					Options: []chrome.Option{
						chrome.ARCEnabled(),
					},
					Expr: "location.href",
				},
			},
			{
				Name:              "arc_state",
				ExtraSoftwareDeps: []string{"chrome", "android"},
				Val: manyParamsParams{
					Options: []chrome.Option{
						chrome.ARCEnabled(),
					},
					Expr: "document.readyState",
				},
			},
		},
	})
}

func ManyParams(ctx context.Context, s *testing.State) {
	p := s.Param().(manyParamsParams)

	cr, err := chrome.New(ctx, p.Options...)
	if err != nil {
		s.Fatal("Failed to start Chrome: ", err)
	}
	defer cr.Close(ctx)

	conn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to obtain test API connection: ", err)
	}

	var out string
	if err := conn.Eval(ctx, p.Expr, &out); err != nil {
		s.Fatal("Eval failed: ", err)
	}

	s.Logf("Eval(%q) = %q", p.Expr, out)
}
