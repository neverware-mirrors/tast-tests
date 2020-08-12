// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wifi

import (
	"context"
	"time"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/remote/network/iw"
	"chromiumos/tast/remote/wificell"
	"chromiumos/tast/services/cros/network"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:        LinkMonitorFailure,
		Desc:        "Verifies how fast the DUT detects the link failure and reconnects to the AP when an AP changes its DHCP configuration",
		Contacts:    []string{"chharry@google.com", "chromeos-platform-connectivity@google.com"},
		Attr:        []string{"group:wificell", "wificell_unstable", "wificell_func"},
		ServiceDeps: []string{"tast.cros.network.WifiService"},
		Vars:        []string{"router"},
	})
}

func LinkMonitorFailure(ctx context.Context, s *testing.State) {
	const (
		// Passive link monitor takes up to 25 seconds to fail; active link monitor takes up to 50 seconds to fail.
		linkFailureDetectedTimeout = 80 * time.Second
		reassociateTimeout         = 10 * time.Second
	)

	s.Log("Setting up the test fixture and AP")
	router, _ := s.Var("router")
	tf, err := wificell.NewTestFixture(ctx, ctx, s.DUT(), s.RPCHint(), wificell.TFRouter(router))
	if err != nil {
		s.Fatal("Failed to set up the test fixture: ", err)
	}
	defer func(ctx context.Context) {
		if err := tf.Close(ctx); err != nil {
			s.Log("Failed to tear down test fixture: ", err)
		}
	}(ctx)
	ctx, cancel := tf.ReserveForClose(ctx)
	defer cancel()

	ap, err := tf.DefaultOpenNetworkAP(ctx)
	if err != nil {
		s.Fatal("Failed to configure the AP: ", err)
	}
	defer func(ctx context.Context) {
		if err := tf.DeconfigAP(ctx, ap); err != nil {
			s.Error("Failed to deconfig the AP: ", err)
		}
	}(ctx)
	ctx, cancel = tf.ReserveForDeconfigAP(ctx, ap)
	defer cancel()
	s.Log("Test fixture setup done; connecting the DUT to the AP")

	if _, err := tf.ConnectWifiAP(ctx, ap); err != nil {
		s.Fatal("Failed to connect to WiFi: ", err)
	}
	defer func(ctx context.Context) {
		if err := tf.DisconnectWifi(ctx); err != nil {
			s.Error("Failed to disconnect WiFi: ", err)
		}
		req := &network.DeleteEntriesForSSIDRequest{Ssid: []byte(ap.Config().SSID)}
		if _, err := tf.WifiClient().DeleteEntriesForSSID(ctx, req); err != nil {
			s.Errorf("Failed to remove entries for ssid=%s: %v", ap.Config().SSID, err)
		}
	}(ctx)
	ctx, cancel = ctxutil.Shorten(ctx, 5*time.Second)
	defer cancel()
	s.Log("WiFi connected; starting the test")

	if err := tf.PingFromDUT(ctx, ap.ServerIP().String()); err != nil {
		s.Fatal("Failed to ping from the DUT: ", err)
	}

	el, err := iw.NewEventLogger(ctx, s.DUT())
	if err != nil {
		s.Fatal("Failed to create iw event logger: ", err)
	}
	defer el.Stop(ctx)

	// Start to change the DHCP config.

	// Obtain current time from the DUT because we use the "disconnect" event timestamp as
	// the end time of the link failure detection duration, which is from the DUT's clock.
	linkFailureTime, err := tf.CurrentClientTime(ctx)
	if err != nil {
		s.Fatal("Failed to get the current DUT time: ", err)
	}
	if err := tf.Router().ChangeAPIfaceSubnetIdx(ctx, ap); err != nil {
		s.Fatal("Failed to change the subnet index of the AP: ", err)
	}

	s.Log("Waiting for link failure detected event")
	if err := testing.Poll(ctx, func(context.Context) error {
		_, err := el.DisconnectTime()
		return err
	}, &testing.PollOptions{
		Timeout:  linkFailureDetectedTimeout,
		Interval: time.Second * 3,
	}); err != nil {
		s.Fatal("Failed to wait for link failure detected event: ", err)
	}

	// Calculate duration for sensing the link failure.
	linkFailureDetectedTime, err := el.DisconnectTime()
	if err != nil {
		s.Fatal("Failed to get link failure detection time: ", err)
	}
	linkFailureDetectedDuration := linkFailureDetectedTime.Sub(linkFailureTime)
	if linkFailureDetectedDuration > linkFailureDetectedTimeout {
		s.Error("Failed to detect link failure within given timeout")
	}
	s.Logf("Link failure detection time: %.2f seconds", linkFailureDetectedDuration.Seconds())

	s.Log("Waiting for reassociation to complete")
	if err := testing.Poll(ctx, func(context.Context) error {
		_, err := el.ConnectedTime()
		return err
	}, &testing.PollOptions{
		Timeout:  reassociateTimeout,
		Interval: time.Second * 1,
	}); err != nil {
		s.Error("Failed to wait for reassociation to complete: ", err)
	}

	// Get the reassociation time.
	connectedTime, err := el.ConnectedTime()
	if err != nil {
		s.Fatal("Failed to get connected time: ", err)
	}
	reassociateDuration := connectedTime.Sub(linkFailureDetectedTime)
	if reassociateDuration < 0 {
		s.Errorf("Unexpected reassociate duration: %d is negative", reassociateDuration)
	}
	if reassociateDuration > reassociateTimeout {
		s.Error("Failed to reassociate within given timeout")
	}
	s.Logf("Reassociate time: %.2f seconds", reassociateDuration.Seconds())
}