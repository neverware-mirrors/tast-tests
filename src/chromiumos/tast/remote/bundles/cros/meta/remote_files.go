// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package meta

import (
	"context"
	"path/filepath"

	"chromiumos/tast/fsutil"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:     RemoteFiles,
		Desc:     "Helper test that uses data and output files",
		Contacts: []string{"tast-owners@google.com"},
		// This test is called by remote tests in the meta package.
		Attr: []string{"disabled"},
		Data: []string{"remote_files_data.txt"},
	})
}

func RemoteFiles(ctx context.Context, s *testing.State) {
	const fn = "remote_files_data.txt"
	s.Log("Copying ", fn)
	if err := fsutil.CopyFile(s.DataPath(fn), filepath.Join(s.OutDir(), fn)); err != nil {
		s.Fatal("Failed copying file: ", err)
	}
}
