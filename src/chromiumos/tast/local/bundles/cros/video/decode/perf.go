// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package decode

import (
	"encoding/json"
	"os"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/perf"
)

// This file contains helper functions that can be used to parse the log files
// generated by the video_decode_accelerator_perf_tests.

// parseUncappedPerfMetrics parses the JSON log file generated by the
// MeasureUncappedPerformance test.
func parseUncappedPerfMetrics(metricsPath string, p *perf.Values) error {
	f, err := os.Open(metricsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var metrics struct {
		FrameDeliveryTimeAverage      float64
		FrameDeliveryTimePercentile25 float64
		FrameDeliveryTimePercentile50 float64
		FrameDeliveryTimePercentile75 float64
		FrameDeliveryTimes            []float64
	}

	if err := json.NewDecoder(f).Decode(&metrics); err != nil {
		return errors.Wrapf(err, "failed decoding %s", metricsPath)
	}

	// TODO(dstaessens@): Remove "tast_" prefix after removing video_VDAPerf in autotest.
	p.Set(perf.Metric{
		Name:      "tast_delivery_time.average",
		Unit:      "milliseconds",
		Direction: perf.SmallerIsBetter,
	}, metrics.FrameDeliveryTimeAverage)
	p.Set(perf.Metric{
		Name:      "tast_delivery_time.first",
		Unit:      "milliseconds",
		Direction: perf.SmallerIsBetter,
	}, metrics.FrameDeliveryTimes[0])
	p.Set(perf.Metric{
		Name:      "tast_delivery_time.percentile_0.25",
		Unit:      "milliseconds",
		Direction: perf.SmallerIsBetter,
	}, metrics.FrameDeliveryTimePercentile25)
	p.Set(perf.Metric{
		Name:      "tast_delivery_time.percentile_0.50",
		Unit:      "milliseconds",
		Direction: perf.SmallerIsBetter,
	}, metrics.FrameDeliveryTimePercentile50)
	p.Set(perf.Metric{
		Name:      "tast_delivery_time.percentile_0.75",
		Unit:      "milliseconds",
		Direction: perf.SmallerIsBetter,
	}, metrics.FrameDeliveryTimePercentile75)

	return nil
}

// parseCappedPerfMetrics parses the JSON log file generated by the
// MeasureCappedPerformance test.
func parseCappedPerfMetrics(metricsPath string, p *perf.Values) error {
	f, err := os.Open(metricsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var metrics struct {
		DroppedFramePercentage      float64
		FrameDecodeTimePercentile50 float64
	}

	if err := json.NewDecoder(f).Decode(&metrics); err != nil {
		return errors.Wrapf(err, "failed decoding %s", metricsPath)
	}

	// TODO(dstaessens@): Remove "tast_" prefix after removing video_VDAPerf in autotest.
	p.Set(perf.Metric{
		Name:      "tast_frame_drop_percentage",
		Unit:      "percent",
		Direction: perf.SmallerIsBetter,
	}, metrics.DroppedFramePercentage)
	p.Set(perf.Metric{
		Name:      "tast_decode_time.percentile_0.50",
		Unit:      "milliseconds",
		Direction: perf.SmallerIsBetter,
	}, metrics.FrameDecodeTimePercentile50)

	return nil
}
