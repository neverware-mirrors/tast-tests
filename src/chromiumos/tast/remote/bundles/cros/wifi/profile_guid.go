// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wifi

import (
	"context"
	"encoding/hex"
	"strings"

	"chromiumos/tast/common/network/protoutil"
	"chromiumos/tast/common/shillconst"
	"chromiumos/tast/common/wifi/security/wpa"
	"chromiumos/tast/errors"
	"chromiumos/tast/remote/wificell"
	"chromiumos/tast/remote/wificell/hostapd"
	"chromiumos/tast/services/cros/network"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:        ProfileGUID,
		Desc:        "Verifies that shill correctly handles GUIDs (Globally Unique IDentifier) in the context of WiFi services",
		Contacts:    []string{"chharry@google.com", "chromeos-platform-connectivity@google.com"},
		Attr:        []string{"group:wificell", "wificell_func", "wificell_unstable"},
		ServiceDeps: []string{"tast.cros.network.WifiService"},
		Vars:        []string{"router"},
	})
}

func ProfileGUID(ctx context.Context, s *testing.State) {
	const (
		guid      = "01234"
		password1 = "chromeos1"
		password2 = "chromeos2"
	)

	router, _ := s.Var("router")
	tf, err := wificell.NewTestFixture(ctx, ctx, s.DUT(), s.RPCHint(), wificell.TFRouter(router))
	if err != nil {
		s.Fatal("Failed to set up test fixture: ", err)
	}
	defer func(ctx context.Context) {
		if err := tf.Close(ctx); err != nil {
			s.Error("Failed to tear down test fixture: ", err)
		}
	}(ctx)
	ctx, cancel := tf.ReserveForClose(ctx)
	defer cancel()

	ssid := hostapd.RandomSSID("TAST_TEST_")
	defer func(ctx context.Context) {
		req := &network.DeleteEntriesForSSIDRequest{Ssid: []byte(ssid)}
		if _, err := tf.WifiClient().DeleteEntriesForSSID(ctx, req); err != nil {
			s.Errorf("Failed to remove entries for ssid=%s: %v", ssid, err)
		}
	}(ctx)

	configureAPWithPassword := func(ctx context.Context, password string) (*wificell.APIface, error) {
		apOps := []hostapd.Option{
			hostapd.SSID(ssid),
			hostapd.Mode(hostapd.Mode80211b),
			hostapd.Channel(1),
		}
		secConfFac := wpa.NewConfigFactory(
			password,
			wpa.Mode(wpa.ModePureWPA),
			wpa.Ciphers(wpa.CipherCCMP),
		)
		return tf.ConfigureAP(ctx, apOps, secConfFac)
	}

	shillPropsWithGUID := func(conf *hostapd.Config) (map[string]interface{}, error) {
		props := map[string]interface{}{
			shillconst.ServicePropertyGUID:           guid,
			shillconst.ServicePropertyType:           shillconst.TypeWifi,
			shillconst.ServicePropertyWiFiHexSSID:    strings.ToUpper(hex.EncodeToString([]byte(conf.SSID))),
			shillconst.ServicePropertyWiFiHiddenSSID: conf.Hidden,
			shillconst.ServicePropertySecurityClass:  conf.SecurityConfig.Class(),
			shillconst.ServicePropertyAutoConnect:    true,
		}
		secProps, err := conf.SecurityConfig.ShillServiceProperties()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get shill security properties")
		}
		for k, v := range secProps {
			props[k] = v
		}
		return props, nil
	}

	func(ctx context.Context) {
		ap, err := configureAPWithPassword(ctx, password1)
		if err != nil {
			s.Fatal("Failed to configure ap: ", err)
		}
		defer func(ctx context.Context) {
			if err := tf.DeconfigAP(ctx, ap); err != nil {
				s.Fatal("Failed to deconfig ap: ", err)
			}
		}(ctx)
		ctx, cancel := tf.ReserveForDeconfigAP(ctx, ap)
		defer cancel()

		// Configure service with complete properties, including GUID.
		props, err := shillPropsWithGUID(ap.Config())
		if err != nil {
			s.Fatal("Failed to generate shill properties: ", err)
		}
		servicePath, err := configureAndAssertAutoConnect(ctx, tf, props)
		if err != nil {
			s.Fatal("Failed to configure service and wait for connection: ", err)
		}

		if err := assertGUID(ctx, tf, servicePath, guid); err != nil {
			s.Fatal("Failed on GUID assert: ", err)
		}
	}(ctx)

	func(ctx context.Context) {
		ap, err := configureAPWithPassword(ctx, password2)
		if err != nil {
			s.Fatal("Failed to configure ap: ", err)
		}
		defer func(ctx context.Context) {
			if err := tf.DeconfigAP(ctx, ap); err != nil {
				s.Fatal("Failed to deconfig ap: ", err)
			}
		}(ctx)
		ctx, cancel := tf.ReserveForDeconfigAP(ctx, ap)
		defer cancel()

		// Change the password of the AP and modify only the password of the configuration with GUID.
		props := map[string]interface{}{
			shillconst.ServicePropertyGUID:       guid,
			shillconst.ServicePropertyPassphrase: password2,
		}
		servicePath, err := configureAndAssertAutoConnect(ctx, tf, props)
		if err != nil {
			s.Fatal("Failed to configure service and wait for connection: ", err)
		}

		if err := assertGUID(ctx, tf, servicePath, guid); err != nil {
			s.Fatal("Failed on GUID assert: ", err)
		}

		// Make sure that the GUID is missing after deleting the entries.
		req := &network.DeleteEntriesForSSIDRequest{Ssid: []byte(ssid)}
		if _, err := tf.WifiClient().DeleteEntriesForSSID(ctx, req); err != nil {
			s.Fatalf("Failed to remove entries for ssid=%s: %v", ssid, err)
		}
		if err := assertGUID(ctx, tf, servicePath, ""); err != nil {
			s.Fatal("Failed on GUID assert: ", err)
		}
	}(ctx)
}

func assertGUID(ctx context.Context, tf *wificell.TestFixture, servicePath, expectedGUID string) error {
	res, err := tf.WifiClient().QueryService(ctx, &network.QueryServiceRequest{Path: servicePath})
	if err != nil {
		return errors.Wrap(err, "failed to query service info")
	}
	if res.Guid != expectedGUID {
		return errors.Errorf("GUID not match: got %q want %q", res.Guid, expectedGUID)
	}
	return nil
}

func configureAndAssertAutoConnect(ctx context.Context, tf *wificell.TestFixture, props map[string]interface{}) (string, error) {
	propsEnc, err := protoutil.EncodeToShillValMap(props)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode shill properties")
	}
	service, err := tf.WifiClient().ConfigureAndAssertAutoConnect(ctx,
		&network.ConfigureAndAssertAutoConnectRequest{Props: propsEnc},
	)
	if err != nil {
		return "", err
	}
	return service.Path, nil
}