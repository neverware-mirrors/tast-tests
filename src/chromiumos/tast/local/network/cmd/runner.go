// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cmd contains the interface for running commands in packages such as iw and ping.
package cmd

import (
	"context"

	"chromiumos/tast/common/network/cmd"
	"chromiumos/tast/local/testexec"
)

// LocalCmdRunner is the object used for running local commands.
type LocalCmdRunner struct{}

var _ cmd.Runner = (*LocalCmdRunner)(nil)

// Run runs a command and waits for its completion.
func (r *LocalCmdRunner) Run(ctx context.Context, cmd string, args ...string) error {
	return testexec.CommandContext(ctx, cmd, args...).Run(testexec.DumpLogOnError)
}

// Output runs a command, waits for its completion and returns stdout output of the command.
func (r *LocalCmdRunner) Output(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	return testexec.CommandContext(ctx, cmd, args...).Output(testexec.DumpLogOnError)
}