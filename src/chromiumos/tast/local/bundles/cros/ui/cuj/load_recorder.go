// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cuj

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"

	"chromiumos/tast/common/perf"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

const checkInterval = 300 * time.Millisecond
const waitUntilReady = 3 * time.Second

const totalEntryName = "total"

type loadEntry struct {
	privateMemory int64
	cpuPercent    float64
}

// loadRecorder records the load (CPU time and memory) of the processes
// asynchronously and creates their perf reports.
type loadRecorder struct {
	procNames map[int32]string
	procs     []*process.Process

	cancel context.CancelFunc
	errorc chan error

	mutex     sync.Mutex
	recording bool
	prepares  []map[string]*loadEntry
	records   []map[string]*loadEntry
}

// browserProcData searches browser-related process IDs and fill their data
// to |procNames|.
func browserProcData(procNames map[int32]string) error {
	browserPID, err := chrome.GetRootPID()
	if err != nil {
		return errors.Wrap(err, "failed to find the browser process")
	}
	gpuProcs, err := chrome.GetGPUProcesses()
	if err != nil {
		return errors.Wrap(err, "failed to find the GPU process")
	}
	if len(gpuProcs) != 1 {
		return errors.Errorf("found %d GPU processes, expected to have one", len(gpuProcs))
	}
	procNames[int32(browserPID)] = "browser"
	procNames[gpuProcs[0].Pid] = "gpu"
	return nil
}

// arcProcData searches ARC-related process IDs and fill their data to
// |procNames|.
func arcProcData(procNames map[int32]string) error {
	procs, err := process.Processes()
	if err != nil {
		return errors.Wrap(err, "failed to get the process list")
	}
	for _, proc := range procs {
		cmdline, err := proc.Cmdline()
		if err != nil {
			continue
		}
		if cmdline == "system_server" {
			procNames[proc.Pid] = "system_server"
			return nil
		}
	}
	// It's fine the system_server does not exist; ARC might not be activated,
	// or it's ARCVM and we can't check its load.
	// TODO(mukai): consider the case of ARCVM.
	return nil
}

// newLoadRecorder creates a new loadRecorder.
func newLoadRecorder(ctx context.Context, procNames map[int32]string) (*loadRecorder, error) {
	var procs []*process.Process
	for pid, name := range procNames {
		p, err := process.NewProcess(pid)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find the process %d (%s)", pid, name)
		}
		procs = append(procs, p)
	}

	// before dispatching all CPU state; invokes cpu.Percent to reset the 'last'
	// state.
	if _, err := cpu.Percent(0, false); err != nil {
		return nil, errors.Wrap(err, "failed to compute the entire CPU percent")
	}
	ctx, cancel := context.WithCancel(ctx)
	lr := &loadRecorder{
		procNames: procNames,
		procs:     procs,
		cancel:    cancel,
		errorc:    make(chan error),
		recording: true,
		records:   make([]map[string]*loadEntry, 0, waitUntilReady/checkInterval),
	}
	go func() {
		defer close(lr.errorc)
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			err := lr.check()
			if err != nil {
				lr.errorc <- err
				return
			}
		}
	}()
	if err := func() error {
		if err := testing.Sleep(ctx, waitUntilReady); err != nil {
			return err
		}
		lr.mutex.Lock()
		defer lr.mutex.Unlock()
		// No data is recorded for prepares. Causing an error.
		if len(lr.records) == 0 {
			return errors.New("no records found for preparations")
		}
		lr.recording = false
		// Copying the records so far to the prepare.
		lr.prepares = lr.records
		lr.records = make([]map[string]*loadEntry, 0, cap(lr.prepares))
		return nil
	}(); err != nil {
		lr.Stop()
		return nil, err
	}
	return lr, nil
}

// Stop finishes the recording goroutine and returns the error when the
// goroutine has met an error. If it's been stopped already, do nothing and
// return nil.
func (lr *loadRecorder) Stop() error {
	if lr.cancel == nil {
		return nil
	}
	lr.cancel()
	lr.cancel = nil
	return <-lr.errorc
}

func (lr *loadRecorder) StartRecording() {
	lr.mutex.Lock()
	lr.recording = true
	lr.mutex.Unlock()
}

func (lr *loadRecorder) StopRecording() {
	lr.mutex.Lock()
	lr.recording = false
	lr.mutex.Unlock()
}

func (lr *loadRecorder) check() error {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()
	if !lr.recording {
		return nil
	}

	record := make(map[string]*loadEntry, len(lr.procs)+1)
	for _, p := range lr.procs {
		pid := p.Pid
		name := lr.procNames[pid]
		mstat, err := p.MemoryInfoEx()
		if err != nil {
			return errors.Wrapf(err, "failed to get memory info for %d", pid)
		}
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			return errors.Wrapf(err, "failed to get CPU percent for %d", pid)
		}
		record[name] = &loadEntry{
			privateMemory: int64(mstat.RSS) - int64(mstat.Shared),
			cpuPercent:    cpuPercent,
		}
	}
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return errors.Wrap(err, "failed to obtain the entire CPU percent")
	}
	record[totalEntryName] = &loadEntry{cpuPercent: cpuPercents[0]}

	lr.records = append(lr.records, record)
	return nil
}

func (lr *loadRecorder) Save(pv *perf.Values) error {
	if len(lr.records) == 0 {
		return errors.New("no load data are recorded")
	}

	prepareMaxes := make(map[string]*loadEntry, len(lr.procNames)+1)
	for _, record := range lr.prepares {
		for name, data := range record {
			pm, ok := prepareMaxes[name]
			if ok {
				if pm.privateMemory < data.privateMemory {
					pm.privateMemory = data.privateMemory
				}
				pm.cpuPercent = math.Max(pm.cpuPercent, data.cpuPercent)
			} else {
				prepareMaxes[name] = &loadEntry{privateMemory: data.privateMemory, cpuPercent: data.cpuPercent}
			}
		}
	}
	sum := make(map[string]*loadEntry, len(lr.procNames)+1)
	maxIncreases := make(map[string]*loadEntry, len(lr.procNames)+1)
	for _, record := range lr.records {
		for name, data := range record {
			if sumStat, ok := sum[name]; ok {
				sumStat.privateMemory += data.privateMemory
				sumStat.cpuPercent += data.cpuPercent
			} else {
				sum[name] = &loadEntry{privateMemory: data.privateMemory, cpuPercent: data.cpuPercent}
			}
			prepareMax := prepareMaxes[name]
			increase := &loadEntry{
				privateMemory: data.privateMemory - prepareMax.privateMemory,
				cpuPercent:    data.cpuPercent - prepareMax.cpuPercent,
			}
			if maxIncrease, ok := maxIncreases[name]; ok {
				if maxIncrease.privateMemory < increase.privateMemory {
					maxIncrease.privateMemory = increase.privateMemory
				}
				maxIncrease.cpuPercent = math.Max(maxIncrease.cpuPercent, increase.cpuPercent)
			} else {
				maxIncreases[name] = increase
			}
		}
	}
	for name, sumStat := range sum {
		if name != totalEntryName {
			pv.Set(perf.Metric{
				Name:      fmt.Sprintf("%s.privateMemory", name),
				Variant:   "average",
				Direction: perf.SmallerIsBetter,
				Unit:      "bytes",
			}, float64(sumStat.privateMemory)/float64(len(lr.records)))
		}
		pv.Set(perf.Metric{
			Name:      fmt.Sprintf("%s.cpuPercent", name),
			Variant:   "average",
			Direction: perf.SmallerIsBetter,
			Unit:      "percent",
		}, sumStat.cpuPercent/float64(len(lr.records)))
	}
	for name, maxIncrease := range maxIncreases {
		if name != totalEntryName {
			pv.Set(perf.Metric{
				Name:      fmt.Sprintf("%s.privateMemory", name),
				Variant:   "maxIncrease",
				Direction: perf.SmallerIsBetter,
				Unit:      "bytes",
			}, float64(maxIncrease.privateMemory))
		}
		pv.Set(perf.Metric{
			Name:      fmt.Sprintf("%s.cpuPercent", name),
			Variant:   "maxIncrease",
			Direction: perf.SmallerIsBetter,
			Unit:      "percent",
		}, maxIncrease.cpuPercent)
	}
	return nil
}
