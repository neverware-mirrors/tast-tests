// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package trace provides common code to replay graphics trace files.
package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"chromiumos/tast/common/perf"
	"chromiumos/tast/ctxutil"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/graphics/trace/comm"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/local/vm"
	"chromiumos/tast/lsbrelease"
	"chromiumos/tast/testing"
)

const (
	outDirName           = "trace"
	glxInfoFile          = "glxinfo.txt"
	replayAppName        = "trace_replay"
	replayAppPathAtHost  = "/usr/local/graphics"
	replayAppPathAtGuest = "/tmp/graphics"
	fileServerPort       = 8085
)

func logContainerInfo(ctx context.Context, cont *vm.Container, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := cont.Command(ctx, "glxinfo")
	cmd.Stdout, cmd.Stderr = f, f
	return cmd.Run()
}

func getSystemInfo(sysInfo *comm.SystemInfo) error {
	lsbReleaseData, err := lsbrelease.Load()
	if err != nil {
		return errors.Wrap(err, "unable to retreive lsbrelease information")
	}

	sysInfo.Board = lsbReleaseData[lsbrelease.Board]
	sysInfo.ChromeOSVersion = lsbReleaseData[lsbrelease.Version]
	return nil
}

func getOutboundIP() (string, error) {
	// 8.8.8.8 is used as a known WAN ip address to get the proper network interface
	// for outbound connections. No actual connection is estabilished.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

func setTCPPortState(ctx context.Context, port int, open bool) error {
	const iptablesApp = "iptables"
	iptablesArgs := []string{"INPUT", "-p", "tcp", "--dport", strconv.Itoa(port), "--syn", "-j", "ACCEPT"}
	checkCmd := testexec.CommandContext(ctx, iptablesApp, append([]string{"-C"}, iptablesArgs...)...)
	err := checkCmd.Run()
	exitCode, ok := testexec.ExitCode(err)
	if !ok {
		return err
	}
	var iptablesActionArg string
	if open == true && exitCode != 0 {
		iptablesActionArg = "-I"
	} else if open == false && exitCode == 0 {
		iptablesActionArg = "-D"
	} else {
		return nil
	}
	toggleCmd := testexec.CommandContext(ctx, iptablesApp, append([]string{iptablesActionArg}, iptablesArgs...)...)
	return toggleCmd.Run()
}

// File server routine. It serves all the artifact requests request from the guest.
type fileServer struct {
	CloudStorage *testing.CloudStorage
	Repository   *comm.RepositoryInfo
}

func validateRequestedFilePath(filePath string) bool {
	// detect dot segments in filePath using path.Join()
	return path.Join("/", filePath)[1:] == filePath
}

func (s *fileServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	testing.ContextLog(ctx, "[Proxy Server] Serving request: ", req.URL.RawQuery)
	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		testing.ContextLogf(ctx, "[Proxy Server] Error: Unable to parse request query <%s>: %s", req.URL.RawQuery, err.Error())
		wr.WriteHeader(http.StatusBadRequest)
		return
	}

	requestedFilePath := query["d"][0]
	// The requested path is specified relative to the repository root URL to restrict access to arbitrary
	// files via this request.
	if !validateRequestedFilePath(requestedFilePath) {
		testing.ContextLogf(ctx, "[Proxy Server] Error: Unable to validate the requested path  <%s>", requestedFilePath)
		wr.WriteHeader(http.StatusUnauthorized)
		return
	}

	requestURL, e := url.Parse(s.Repository.RootURL)
	if e != nil {
		testing.ContextLogf(ctx, "[Proxy Server] Error: Unable to parse the repository URL <%s>: %s", s.Repository.RootURL, err.Error())
		wr.WriteHeader(http.StatusBadRequest)
		return
	}

	requestURL.Path = path.Join(requestURL.Path, requestedFilePath)
	testing.ContextLog(ctx, "[Proxy Server] Downloading: ", requestURL.String())
	r, err := s.CloudStorage.Open(ctx, requestURL.String())
	if err != nil {
		testing.ContextLogf(ctx, "[Proxy Server] Error: Unable to open <%v>: %v", requestURL, err)
		wr.WriteHeader(http.StatusNotFound)
		return
	}
	defer r.Close()

	wr.Header().Set("Content-Disposition", "attachment; filename="+path.Base(requestedFilePath))
	wr.WriteHeader(http.StatusOK)

	copied, err := io.Copy(wr, r)
	if err != nil {
		testing.ContextLog(ctx, "[Proxy Server] Error: io.Copy() failed: ", err)
	}
	testing.ContextLogf(ctx, "[Proxy Server] Request served successfully. %d byte(s) copied", copied)
}

func startFileServer(ctx context.Context, addr string, cloudStorage *testing.CloudStorage, repository *comm.RepositoryInfo) *http.Server {
	handler := &fileServer{CloudStorage: cloudStorage, Repository: repository}
	testing.ContextLog(ctx, "Starting server at "+addr)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			testing.ContextLog(ctx, "Error: ListenAndServe() failed: ", err)
		}
	}()
	return server
}

func pushTraceReplayApp(ctx context.Context, cont *vm.Container) error {
	replayAppAtHost := path.Join(replayAppPathAtHost, replayAppName)
	if _, err := os.Stat(replayAppAtHost); os.IsNotExist(err) {
		return errors.Wrapf(err, "unable to locate replay app at host: <%s> not found", replayAppAtHost)
	}

	if err := cont.Command(ctx, "mkdir", "-p", replayAppPathAtGuest).Run(); err != nil {
		return errors.Wrapf(err, "unable to create directory <%s> inside the container", replayAppPathAtGuest)
	}

	replayAppAtGuest := path.Join(replayAppPathAtGuest, replayAppName)
	testing.ContextLogf(ctx, "Copying %s to the container <%s>", replayAppName, replayAppPathAtGuest)
	if err := cont.PushFile(ctx, replayAppAtHost, replayAppAtGuest); err != nil {
		return errors.Wrap(err, "unable to copy replay app into the guest container")
	}
	return nil
}

// RunTraceReplayTest starts a VM and replays all the traces in the test config.
func RunTraceReplayTest(ctx context.Context, resultDir string, cloudStorage *testing.CloudStorage, cont *vm.Container, group *comm.TestGroupConfig) error {
	// Guest is unable to use VM network interface to access it's host because of security reason,
	// and the only to make such connectivity is to use host's outbound network interface.
	outboundIP, err := getOutboundIP()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve outbound IP address")
	}
	testing.ContextLog(ctx, "Outbound IP address: ", outboundIP)

	if err := setTCPPortState(ctx, fileServerPort, true); err != nil {
		return errors.Wrap(err, "unable to open TCP port")
	}

	defer func() {
		if err := setTCPPortState(ctx, fileServerPort, false); err != nil {
			testing.ContextLog(ctx, "Unable to close TCP port: ", err)
		}
	}()

	outDir := filepath.Join(resultDir, outDirName)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create output dir <%v>", outDir)
	}
	file := filepath.Join(outDir, glxInfoFile)
	testing.ContextLog(ctx, "Logging container graphics environment to ", glxInfoFile)
	if err := logContainerInfo(ctx, cont, file); err != nil {
		testing.ContextLog(ctx, "Warning: Unable to log container information: ", err)
	}

	if err := getSystemInfo(&group.Host); err != nil {
		return err
	}

	if err := pushTraceReplayApp(ctx, cont); err != nil {
		return err
	}

	serverAddr := fmt.Sprintf("%s:%d", outboundIP, fileServerPort)
	server := startFileServer(ctx, serverAddr, cloudStorage, &group.Repository)
	defer func() {
		if err := server.Shutdown(ctx); err != nil {
			testing.ContextLog(ctx, "Unable to shutdown file server: ", err)
		}
	}()
	shortCtx, shortCancel := ctxutil.Shorten(ctx, 30*time.Second)
	defer shortCancel()
	perfValues := perf.NewValues()

	group.ProxyServer = comm.ProxyServerInfo{
		URL: "http://" + serverAddr,
	}

	replayArgs, err := json.Marshal(*group)
	if err != nil {
		return err
	}

	testing.ContextLog(shortCtx, "Running replay with args: "+string(replayArgs))
	replayCmd := cont.Command(shortCtx, path.Join(replayAppPathAtGuest, replayAppName), string(replayArgs))
	replayOutput, err := replayCmd.Output()
	if err != nil {
		return err
	}

	testing.ContextLog(shortCtx, "Replay output: "+string(replayOutput))

	var testResult comm.TestGroupResult
	if err := json.Unmarshal(replayOutput, &testResult); err != nil {
		return errors.Wrapf(err, "unable to parse test group result output: %q", string(replayOutput))
	}

	type getFieldValueFn func(val comm.ReplayResult) float64
	getValues := func(vals []comm.ReplayResult, fn getFieldValueFn) []float64 {
		var values []float64
		for _, val := range vals {
			values = append(values, fn(val))
		}
		return values
	}

	failedEntries := 0
	for _, resultEntry := range testResult.Entries {
		if resultEntry.Message != "" {
			testing.ContextLog(shortCtx, resultEntry.Message)
		}
		if resultEntry.Result != comm.TestResultSuccess {
			failedEntries++
			continue
		}
		perfValues.Set(perf.Metric{
			Name:      resultEntry.Name,
			Variant:   "time",
			Unit:      "sec",
			Direction: perf.SmallerIsBetter,
			Multiple:  true,
		}, getValues(resultEntry.Values, func(r comm.ReplayResult) float64 {
			return float64(r.DurationInSeconds)
		})...)
		perfValues.Set(perf.Metric{
			Name:      resultEntry.Name,
			Variant:   "frames",
			Unit:      "frame",
			Direction: perf.BiggerIsBetter,
			Multiple:  true,
		}, getValues(resultEntry.Values, func(r comm.ReplayResult) float64 {
			return float64(r.TotalFrames)
		})...)
		perfValues.Set(perf.Metric{
			Name:      resultEntry.Name,
			Variant:   "fps",
			Unit:      "fps",
			Direction: perf.BiggerIsBetter,
			Multiple:  true,
		}, getValues(resultEntry.Values, func(r comm.ReplayResult) float64 {
			return float64(r.AverageFPS)
		})...)
	}

	if err := perfValues.Save(resultDir); err != nil {
		return errors.Wrap(err, "unable to save performance values")
	}

	if testResult.Result != comm.TestResultSuccess {
		return errors.Errorf("%s", testResult.Message)
	}

	return nil
}
