// Copyright 2017 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main implements an executable containing local tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"chromiumos/tast/control"
	"chromiumos/tast/local/crash"
	"chromiumos/tast/local/logs"
	"chromiumos/tast/runner"

	// These packages register their tests via init functions.
	_ "chromiumos/tast/local/tests/example"
	_ "chromiumos/tast/local/tests/power"
	_ "chromiumos/tast/local/tests/security"
	_ "chromiumos/tast/local/tests/ui"
)

const (
	systemLogDir      = "/var/log"  // directory where system logs are located
	testTimeout       = time.Minute // max running time for each test
	maxCrashesPerExec = 3           // max crashes to collect per executable
)

// getInitialLogSizes returns the starting sizes of log files.
// If mw is nil, messages are logged to stdout instead.
func getInitialLogSizes(mw *control.MessageWriter) logs.InodeSizes {
	runner.Log(mw, "Getting original log inode sizes")
	ls, warnings, err := logs.GetLogInodeSizes(systemLogDir)
	for p, err := range warnings {
		runner.Log(mw, fmt.Sprintf("Failed to check log %s: %v", p, err))
	}
	if err != nil {
		runner.Log(mw, fmt.Sprintf("Failed to get original log inode sizes: %v", err))
	}
	return ls
}

// copyLogUpdates copies updated portions of system logs to a temporary dir.
// sizes contains the original log sizes and is used to identify old content that won't be copied.
// If mw is nil, messages are logged to stdout instead.
// The directory containing the log updates is returned.
func copyLogUpdates(sizes logs.InodeSizes, mw *control.MessageWriter) (outDir string) {
	runner.Log(mw, "Copying log updates")
	if sizes == nil {
		runner.Log(mw, "Don't have original log sizes")
		return
	}

	var err error
	if outDir, err = ioutil.TempDir("", "local_tests_logs."); err != nil {
		runner.Log(mw, fmt.Sprintf("Failed to create log output dir: %v", err))
		return
	}

	warnings, err := logs.CopyLogFileUpdates(systemLogDir, outDir, sizes)
	for p, werr := range warnings {
		runner.Log(mw, fmt.Sprintf("Failed to copy log %s: %v", p, werr))
	}
	if err != nil {
		runner.Log(mw, fmt.Sprintf("Failed to copy log updates: %v", err))
	}
	return outDir
}

// copyNewMinidumps copies new minidump crash reports into a temporary dir.
// oldDumps contains paths of dump files that existed before the test run started.
func copyNewMinidumps(oldDumps []string, mw *control.MessageWriter) (outDir string) {
	runner.Log(mw, "Copying crashes")
	_, newDumps, err := crash.GetCrashes(crash.DefaultCrashDir)
	if err != nil {
		runner.Log(mw, fmt.Sprintf("Failed to get new crashes: %v", err))
		return
	}
	if outDir, err = ioutil.TempDir("", "local_tests_crashes."); err != nil {
		runner.Log(mw, fmt.Sprintf("Failed to create minidump output dir: %v", err))
		return
	}

	warnings, err := crash.CopyNewFiles(outDir, newDumps, oldDumps, maxCrashesPerExec)
	for p, werr := range warnings {
		runner.Log(mw, fmt.Sprintf("Failed to copy minidump %s: %v", p, werr))
	}
	if err != nil {
		runner.Log(mw, fmt.Sprintf("Failed to copy minidumps: %v", err))
	}
	return outDir
}

func main() {
	cfg := runner.RunConfig{
		Ctx:                context.Background(),
		DefaultTestTimeout: testTimeout,
	}

	flag.StringVar(&cfg.DataDir, "datadir", "/usr/local/share/tast/data",
		"directory where data files are located")
	listData := flag.Bool("listdata", false, "print data files needed for tests and exit")
	listTests := flag.Bool("listtests", false, "print matching tests and exit")
	report := flag.Bool("report", false, "report progress for calling process")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <pattern> <pattern> ...\n"+
			"Runs local tests matched by zero or more patterns.\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *report {
		cfg.MessageWriter = control.NewMessageWriter(os.Stdout)
	}

	var err error
	if cfg.Tests, err = runner.TestsToRun(flag.Args()); err != nil {
		runner.Abort(cfg.MessageWriter, err.Error())
	}

	if *listData {
		if err := listDataFiles(os.Stdout, cfg.Tests); err != nil {
			runner.Abort(cfg.MessageWriter, err.Error())
		}
		os.Exit(0)
	}
	if *listTests {
		if err := runner.PrintTests(os.Stdout, cfg.Tests); err != nil {
			runner.Abort(cfg.MessageWriter, err.Error())
		}
		os.Exit(0)
	}

	if cfg.BaseOutDir, err = ioutil.TempDir("", "local_tests_data."); err != nil {
		runner.Abort(cfg.MessageWriter, err.Error())
	}

	// Perform the test run.
	var logSizes logs.InodeSizes
	var oldMinidumps []string
	if *report {
		cfg.MessageWriter.WriteMessage(&control.RunStart{time.Now(), len(cfg.Tests)})
		logSizes = getInitialLogSizes(cfg.MessageWriter)
		if _, oldMinidumps, err = crash.GetCrashes(crash.DefaultCrashDir); err != nil {
			runner.Log(cfg.MessageWriter, fmt.Sprintf("Failed to get existing minidumps: %v", err))
		}
	}
	numFailed, err := runner.RunTests(cfg)
	if err != nil {
		runner.Abort(cfg.MessageWriter, err.Error())
	}
	if *report {
		logDir := copyLogUpdates(logSizes, cfg.MessageWriter)
		crashDir := copyNewMinidumps(oldMinidumps, cfg.MessageWriter)
		cfg.MessageWriter.WriteMessage(&control.RunEnd{
			Time:     time.Now(),
			LogDir:   logDir,
			CrashDir: crashDir,
			OutDir:   cfg.BaseOutDir})
	}

	// Exit with a nonzero exit code if we were run manually and saw at least one test fail.
	if !*report && numFailed > 0 {
		os.Exit(1)
	}
}
