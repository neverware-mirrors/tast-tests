// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"chromiumos/tast/common/perf"
	"chromiumos/tast/local/colorcmp"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/input"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         MousePerf,
		Desc:         "Performance test for mouse responsiveness",
		Contacts:     []string{"hollingum@google.com", "cros-containers-dev@google.com"},
		Attr:         []string{"group:crosbolt", "crosbolt_perbuild"},
		Vars:         []string{"keepState"},
		Data:         []string{"mouse_perf.py"},
		SoftwareDeps: []string{"chrome", "vm_host"},
		Params: []testing.Param{
			// Parameters generated by params_test.go. DO NOT EDIT.
			{
				Name:              "amd64",
				ExtraData:         []string{"crostini_vm_amd64.zip", "crostini_test_container_metadata_buster_amd64.tar.xz", "crostini_test_container_rootfs_buster_amd64.tar.xz"},
				ExtraSoftwareDeps: []string{"amd64"},
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
			}, {
				Name:              "arm",
				ExtraData:         []string{"crostini_vm_arm.zip", "crostini_test_container_metadata_buster_arm.tar.xz", "crostini_test_container_rootfs_buster_arm.tar.xz"},
				ExtraSoftwareDeps: []string{"arm"},
				Pre:               crostini.StartedByComponentBuster(),
				Timeout:           7 * time.Minute,
			},
		},
	})
}

func MousePerf(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	cr := pre.Chrome
	tconn := pre.TestAPIConn
	cont := pre.Container
	defer crostini.RunCrostiniPostTest(ctx, s.PreValue().(crostini.PreData))

	const appletFile string = "mouse_perf.py"

	// Get access to the mouse and put it nearer the top-left corner.
	mouse, err := input.Mouse(ctx)
	if err != nil {
		s.Fatal("Failed to get the mouse: ", err)
	}
	defer mouse.Close()
	// TODO(hollingum): put some absolute positioning in the API.
	mouse.Move(-1000, -1000)

	// In order to correct for variance in the receive times, we track the send times.
	var sendTimes []float64
	doMouseMove := func(ctx context.Context) error {
		// We want to make sure the app is running and visible (foregrounded),
		// so match a screenshot against the colour the app is known to render in.
		if err := crostini.MatchScreenshotDominantColor(ctx, cr, colorcmp.RGB(127, 0, 255), filepath.Join(s.OutDir(), "screenshot.png")); err != nil {
			return err
		}
		for i := 0; i < 400; i++ {
			startTime := time.Now()
			// Send the event.
			if err := mouse.Move(1, 1); err != nil {
				return err
			}
			// Wait a fixed time before sending the next event.
			if err := testing.Sleep(ctx, 1*time.Millisecond); err != nil {
				return err
			}
			sendTimes = append(sendTimes, float64(time.Now().Sub(startTime))/float64(time.Millisecond))
		}
		return nil
	}

	// Launch the app.
	if err := cont.PushFile(ctx, s.DataPath(appletFile), "/home/testuser/"+appletFile); err != nil {
		s.Fatalf("Failed to push %v to container: %v", appletFile, err)
	}
	output, err := crostini.RunWindowedApp(ctx, tconn, cont, pre.Keyboard, 30*time.Second, doMouseMove, true, "mouse_perf", []string{"python3", appletFile})
	if err != nil {
		s.Fatal("Failed to run app: ", err)
	}

	// Process the output to generate this run's stats.
	//
	// For every motion event the applet receives, it prints x, y coords, and the time, on separate lines (in that order).
	type sample struct {
		X float64
		Y float64
		T float64
	}
	var parsedSamples []sample
	var tempSample sample
	for i, v := range strings.Split(output, "\n") {
		if v == "" {
			continue
		} else if p, err := strconv.ParseFloat(v, 64); err == nil {
			switch i % 3 {
			case 0:
				tempSample.X = p
			case 1:
				tempSample.Y = p
			case 2:
				tempSample.T = p
				parsedSamples = append(parsedSamples, tempSample)
			}
		} else {
			s.Fatalf("Failed to parse %s as a float: %v", v, err)
		}
	}
	if len(parsedSamples) < 2 {
		s.Fatalf("Failed to sample enough mouse movements, received %v samples", len(parsedSamples))
	}
	// Convert the samples to deltas (i.e. with one fewer)
	var distances []float64
	var times []float64
	for i := 1; i < len(parsedSamples); i++ {
		timeDelta := parsedSamples[i].T - parsedSamples[i-1].T
		// This works out to |x| + |y| (a.k.a manhattan distance)
		distanceDelta := math.Abs(parsedSamples[i].X-parsedSamples[i-1].X) + math.Abs(parsedSamples[i].Y-parsedSamples[i-1].Y)
		// TODO(hollingum): For some reason every alternate sample shows no change from the previous. Ignore them until we figure out why.
		if timeDelta == 0 && distanceDelta == 0 {
			continue
		}
		distances = append(distances, distanceDelta)
		times = append(times, timeDelta)
	}
	sendStats := processMousePerfStats(sendTimes)
	timeStats := processMousePerfStats(times)
	distanceStats := processMousePerfStats(distances)

	// Record the stats for Crosbolt.
	value := perf.NewValues()
	recordMousePerfStats(ctx, "send_t", "ms", sendStats, value)
	recordMousePerfStats(ctx, "receive_t", "ms", timeStats, value)
	recordMousePerfStats(ctx, "distance", "px", distanceStats, value)
	value.Save(s.OutDir())
}

type mousePerfStats struct {
	Count             float64 // This needs to be a float for metrics recording.
	UpperBound        float64
	LowerBound        float64
	Average           float64
	StandardDeviation float64
}

func processMousePerfStats(samples []float64) mousePerfStats {
	var lb = samples[0]
	var ub = samples[0]
	var sum, sumVar float64
	for _, sample := range samples {
		ub = math.Max(ub, sample)
		lb = math.Min(lb, sample)
		sum += sample
	}
	n := float64(len(samples))
	avg := sum / n
	for _, sample := range samples {
		sumVar += (sample - avg) * (sample - avg)
	}
	return mousePerfStats{
		Count:             n,
		UpperBound:        ub,
		LowerBound:        lb,
		Average:           avg,
		StandardDeviation: math.Sqrt(sumVar / (n - 1)),
	}
}

func recordMousePerfStats(ctx context.Context, name, unit string, stat mousePerfStats, value *perf.Values) {
	testing.ContextLogf(ctx, "Recording stats for %q: %v", name, stat)
	value.Set(perf.Metric{
		Name:      "crostini_mouse_perf",
		Variant:   name + "_range",
		Unit:      unit,
		Direction: perf.SmallerIsBetter,
		Multiple:  false,
	}, stat.UpperBound-stat.LowerBound)
	value.Set(perf.Metric{
		Name:      "crostini_mouse_perf",
		Variant:   name + "_stdev",
		Unit:      unit,
		Direction: perf.SmallerIsBetter,
		Multiple:  false,
	}, stat.StandardDeviation)
	value.Set(perf.Metric{
		Name:      "crostini_mouse_perf",
		Variant:   name + "_count",
		Unit:      "n",
		Direction: perf.BiggerIsBetter,
		Multiple:  false,
	}, stat.Count)
}
