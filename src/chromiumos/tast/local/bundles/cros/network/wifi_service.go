// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package network

import (
	"context"
	"encoding/hex"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"chromiumos/tast/common/network/protoutil"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/shill"
	"chromiumos/tast/local/upstart"
	"chromiumos/tast/services/cros/network"
	"chromiumos/tast/testing"
	"chromiumos/tast/timing"
)

func init() {
	testing.AddService(&testing.Service{
		Register: func(srv *grpc.Server, s *testing.ServiceState) {
			network.RegisterWifiServiceServer(srv, &WifiService{s: s})
		},
	})
}

// wifiTestProfileName is the profile we create and use for WiFi tests.
const wifiTestProfileName = "test"

// WifiService implements tast.cros.network.Wifi gRPC service.
type WifiService struct {
	s *testing.ServiceState
}

// InitDUT properly initializes the DUT for WiFi tests.
func (s *WifiService) InitDUT(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	// Stop UI to avoid interference from UI (e.g. request scan).
	if err := upstart.StopJob(ctx, "ui"); err != nil {
		return nil, errors.Wrap(err, "failed to stop ui")
	}

	m, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Manager object")
	}
	// Turn on the WiFi device.
	iface, err := shill.WifiInterface(ctx, m, 5*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a WiFi device")
	}
	dev, err := m.DeviceByName(ctx, iface)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find the device for interface %s", iface)
	}
	if err := dev.Enable(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to enable WiFi device")
	}
	if err := s.reinitTestState(ctx, m); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// reinitTestState prepare the environment for WiFi testcase.
func (s *WifiService) reinitTestState(ctx context.Context, m *shill.Manager) error {
	// Clean old profiles.
	if err := s.cleanProfiles(ctx, m); err != nil {
		return errors.Wrap(err, "cleanProfiles failed")
	}
	if err := s.removeWifiEntries(ctx, m); err != nil {
		return errors.Wrap(err, "removeWifiEntries failed")
	}
	// Try to create the test profile.
	if _, err := m.CreateProfile(ctx, wifiTestProfileName); err != nil {
		return errors.Wrap(err, "failed to create the test profile")
	}
	// Push the test profile.
	if _, err := m.PushProfile(ctx, wifiTestProfileName); err != nil {
		return errors.Wrap(err, "failed to push the test profile")
	}
	return nil
}

// ReinitTestState cleans and sets up the environment for a single WiFi testcase.
func (s *WifiService) ReinitTestState(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	m, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Manager object")
	}
	if err := s.reinitTestState(ctx, m); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// TearDown reverts the settings made by InitDUT and InitTestState.
func (s *WifiService) TearDown(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	m, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Manager object")
	}

	var retErr error
	if err := s.cleanProfiles(ctx, m); err != nil {
		retErr = errors.Wrapf(retErr, "cleanProfiles failed: %s", err)
	}
	if err := s.removeWifiEntries(ctx, m); err != nil {
		retErr = errors.Wrapf(retErr, "removeWifiEntries failed: %s", err)
	}
	if err := upstart.EnsureJobRunning(ctx, "ui"); err != nil {
		testing.ContextLog(ctx, "Failed to start ui: ", err)
	}
	if retErr != nil {
		return nil, retErr
	}
	return &empty.Empty{}, nil
}

func (s *WifiService) discoverService(ctx context.Context, m *shill.Manager, props map[string]interface{}) (*shill.Service, error) {
	ctx, st := timing.Start(ctx, "discoverService")
	defer st.End()
	testing.ContextLog(ctx, "Discovering a WiFi service with properties: ", props)

	var service *shill.Service
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		var err error
		service, err = m.FindMatchingService(ctx, props)
		if err == nil {
			return nil
		}
		// Scan WiFi AP again if the expected AP is not found.
		if err2 := m.RequestScan(ctx, shill.TechnologyWifi); err2 != nil {
			return testing.PollBreak(errors.Wrap(err2, "failed to request active scan"))
		}
		return err
	}, &testing.PollOptions{
		Timeout:  15 * time.Second,
		Interval: 200 * time.Millisecond, // RequestScan is spammy, but shill handles that for us.
	}); err != nil {
		return nil, err
	}
	return service, nil
}

// connectService connects to a WiFi service and wait until conntected state.
// The time used for association and configuration is returned when success.
func (s *WifiService) connectService(ctx context.Context, service *shill.Service) (assocTime, configTime time.Duration, retErr error) {
	ctx, st := timing.Start(ctx, "connectService")
	defer st.End()
	testing.ContextLog(ctx, "Connecting to the service: ", service)

	start := time.Now()

	// Spawn watcher before connect.
	pw, err := service.CreateWatcher(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to create watcher")
	}
	defer pw.Close(ctx)

	if err := service.Connect(ctx); err != nil {
		return 0, 0, errors.Wrap(err, "failed to connect to service")
	}

	// Wait until connection established.
	// For debug and profile purpose, it is separated into association
	// and configuration stages.

	// Prepare the state list for ExpectIn.
	var connectedStates []interface{}
	for _, s := range shill.ServiceConnectedStates {
		connectedStates = append(connectedStates, s)
	}
	associatedStates := append(connectedStates, shill.ServiceStateConfiguration)

	testing.ContextLog(ctx, "Associating with ", service)
	assocCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	state, err := pw.ExpectIn(assocCtx, shill.ServicePropertyState, associatedStates)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to associate")
	}
	assocTime = time.Since(start)
	start = time.Now()

	testing.ContextLog(ctx, "Configuring ", service)
	if state == shill.ServiceStateConfiguration {
		// We're not yet in connectedStates, wait until connected.
		configCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if _, err := pw.ExpectIn(configCtx, shill.ServicePropertyState, connectedStates); err != nil {
			return 0, 0, errors.Wrap(err, "failed to configure")
		}
	}
	configTime = time.Since(start)

	return assocTime, configTime, nil
}

// Connect connects to a WiFi service with specific config.
// This is the implementation of network.Wifi/Connect gRPC.
func (s *WifiService) Connect(ctx context.Context, request *network.ConnectRequest) (*network.ConnectResponse, error) {
	ctx, st := timing.Start(ctx, "wifi_service.Connect")
	defer st.End()
	testing.ContextLog(ctx, "Attempting to connect with config: ", request)

	m, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a manager object")
	}

	hexSSID := s.hexSSID(request.Ssid)
	start := time.Now()

	// Configure a service for the hidden SSID as a result of manual input SSID.
	if request.Hidden {
		props := map[string]interface{}{
			shill.ServicePropertyType:           shill.TypeWifi,
			shill.ServicePropertyWiFiHexSSID:    hexSSID,
			shill.ServicePropertyWiFiHiddenSSID: request.Hidden,
			shill.ServicePropertySecurityClass:  request.Security,
		}
		if _, err := m.ConfigureService(ctx, props); err != nil {
			return nil, errors.Wrap(err, "failed to configure a hidden SSID")
		}
	}
	props := map[string]interface{}{
		shill.ServicePropertyType:          shill.TypeWifi,
		shill.ServicePropertyWiFiHexSSID:   hexSSID,
		shill.ServicePropertySecurityClass: request.Security,
	}

	service, err := s.discoverService(ctx, m, props)
	if err != nil {
		return nil, err
	}
	discoveryTime := time.Since(start)

	shillProps, err := protoutil.DecodeFromShillValMap(request.Shillprops)
	if err != nil {
		return nil, err
	}
	for k, v := range shillProps {
		if err = service.SetProperty(ctx, k, v); err != nil {
			return nil, errors.Wrapf(err, "failed to set properties %s to %v", k, v)
		}
	}

	assocTime, configTime, err := s.connectService(ctx, service)
	if err != nil {
		return nil, err
	}

	return &network.ConnectResponse{
		ServicePath:       string(service.ObjectPath()),
		DiscoveryTime:     discoveryTime.Nanoseconds(),
		AssociationTime:   assocTime.Nanoseconds(),
		ConfigurationTime: configTime.Nanoseconds(),
	}, nil
}

// Disconnect disconnects from a WiFi service.
// This is the implementation of network.Wifi/Disconnect gRPC.
func (s *WifiService) Disconnect(ctx context.Context, request *network.DisconnectRequest) (*empty.Empty, error) {
	ctx, st := timing.Start(ctx, "wifi_service.Disconnect")
	defer st.End()

	service, err := shill.NewService(ctx, dbus.ObjectPath(request.ServicePath))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create service object")
	}
	// Spawn watcher before disconnect.
	pw, err := service.CreateWatcher(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher")
	}
	defer pw.Close(ctx)
	if err := service.Disconnect(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to disconnect")
	}
	testing.ContextLog(ctx, "Wait for the service to be idle")
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pw.Expect(timeoutCtx, shill.ServicePropertyState, shill.ServiceStateIdle); err != nil {
		return nil, err
	}
	testing.ContextLog(ctx, "Disconnected")
	return &empty.Empty{}, nil
}

// uint16sToUint32s converts []uint16 to []uint32.
func uint16sToUint32s(s []uint16) []uint32 {
	ret := make([]uint32, len(s))
	for i, v := range s {
		ret[i] = uint32(v)
	}
	return ret
}

// QueryService queries shill service information.
// This is the implementation of network.Wifi/QueryService gRPC.
func (s *WifiService) QueryService(ctx context.Context, req *network.QueryServiceRequest) (*network.QueryServiceResponse, error) {
	ctx, st := timing.Start(ctx, "wifi_service.QueryService")
	defer st.End()

	service, err := shill.NewService(ctx, dbus.ObjectPath(req.Path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create service object")
	}
	props, err := service.GetProperties(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service properties")
	}

	name, err := props.GetString(shill.ServicePropertyName)
	if err != nil {
		return nil, err
	}
	device, err := props.GetObjectPath(shill.ServicePropertyDevice)
	if err != nil {
		return nil, err
	}
	serviceType, err := props.GetString(shill.ServicePropertyType)
	if err != nil {
		return nil, err
	}
	mode, err := props.GetString(shill.ServicePropertyMode)
	if err != nil {
		return nil, err
	}
	state, err := props.GetString(shill.ServicePropertyState)
	if err != nil {
		return nil, err
	}
	visible, err := props.GetBool(shill.ServicePropertyVisible)
	if err != nil {
		return nil, err
	}
	isConnected, err := props.GetBool(shill.ServicePropertyIsConnected)
	if err != nil {
		return nil, err
	}

	bssid, err := props.GetString(shill.ServicePropertyWiFiBSSID)
	if err != nil {
		return nil, err
	}

	ftEnabled, err := props.GetBool(shill.ServicePropertyFTEnabled)
	if err != nil {
		return nil, err
	}
	frequency, err := props.GetUint16(shill.ServicePropertyWiFiFrequency)
	if err != nil {
		return nil, err
	}
	frequencyList, err := props.GetUint16s(shill.ServicePropertyWiFiFrequencyList)
	if err != nil {
		return nil, err
	}
	hexSSID, err := props.GetString(shill.ServicePropertyWiFiHexSSID)
	if err != nil {
		return nil, err
	}
	hiddenSSID, err := props.GetBool(shill.ServicePropertyWiFiHiddenSSID)
	if err != nil {
		return nil, err
	}
	phyMode, err := props.GetUint16(shill.ServicePropertyWiFiPhyMode)
	if err != nil {
		return nil, err
	}

	return &network.QueryServiceResponse{
		Name:        name,
		Device:      string(device),
		Type:        serviceType,
		Mode:        mode,
		State:       state,
		Visible:     visible,
		IsConnected: isConnected,
		Wifi: &network.QueryServiceResponse_Wifi{
			Bssid:         bssid,
			FtEnabled:     ftEnabled,
			Frequency:     uint32(frequency),
			FrequencyList: uint16sToUint32s(frequencyList),
			HexSsid:       hexSSID,
			HiddenSsid:    hiddenSSID,
			PhyMode:       uint32(phyMode),
		},
	}, nil
}

// DeleteEntriesForSSID deletes all WiFi profile entries for a given SSID.
func (s *WifiService) DeleteEntriesForSSID(ctx context.Context, request *network.DeleteEntriesForSSIDRequest) (*empty.Empty, error) {
	ctx, st := timing.Start(ctx, "wifi_service.DeleteEntriesForSSID")
	defer st.End()

	m, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Manager object")
	}
	filter := map[string]interface{}{
		shill.ServicePropertyWiFiHexSSID: s.hexSSID(request.Ssid),
		shill.ProfileEntryPropertyType:   shill.TypeWifi,
	}
	if err := s.removeMatchedEntries(ctx, m, filter); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// cleanProfiles pops and removes all active profiles until default profile and
// then removes the WiFi test profile if still exists.
func (s *WifiService) cleanProfiles(ctx context.Context, m *shill.Manager) error {
	for {
		profile, err := m.ActiveProfile(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get active profile")
		}
		props, err := profile.GetProperties(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get properties from profile object")
		}
		name, err := props.GetString(shill.ProfilePropertyName)
		if name == shill.DefaultProfileName {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to get profile name")
		}
		if err := m.PopProfile(ctx, name); err != nil {
			return errors.Wrap(err, "failed to pop profile")
		}
		if err := m.RemoveProfile(ctx, name); err != nil {
			return errors.Wrap(err, "failed to delete profile")
		}
	}
	// Try to remove the test profile.
	m.RemoveProfile(ctx, wifiTestProfileName)
	return nil
}

// removeWifiEntries removes all the entries with type=wifi in all profiles.
func (s *WifiService) removeWifiEntries(ctx context.Context, m *shill.Manager) error {
	filter := map[string]interface{}{
		shill.ProfileEntryPropertyType: shill.TypeWifi,
	}
	return s.removeMatchedEntries(ctx, m, filter)
}

// removeMatchedEntries traverses all profiles and removes all entries matching the properties in propFilter.
func (s *WifiService) removeMatchedEntries(ctx context.Context, m *shill.Manager, propFilter map[string]interface{}) error {
	profiles, err := m.Profiles(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get profiles")
	}
	for _, p := range profiles {
		props, err := p.GetProperties(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get properties from profile object")
		}
		entryIDs, err := props.GetStrings(shill.ProfilePropertyEntries)
		if err != nil {
			return errors.Wrapf(err, "failed to get entryIDs from profile %s", p.String())
		}
	entryLoop:
		for _, entryID := range entryIDs {
			entry, err := p.GetEntry(ctx, entryID)
			if err != nil {
				return errors.Wrapf(err, "failed to get entry %s", entryID)
			}

			for k, expect := range propFilter {
				v, ok := entry[k]
				if !ok || !reflect.DeepEqual(expect, v) {
					// not matched, try new entry.
					continue entryLoop
				}
			}
			if err := p.DeleteEntry(ctx, entryID); err != nil {
				return errors.Wrapf(err, "failed to delete entry %s", entryID)
			}
		}
	}
	return nil
}

// GetInterface returns the WiFi device interface name (e.g., wlan0).
func (s *WifiService) GetInterface(ctx context.Context, e *empty.Empty) (*network.GetInterfaceResponse, error) {
	manager, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shill manager proxy")
	}
	netIf, err := shill.WifiInterface(ctx, manager, 5*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the WiFi interface")
	}
	return &network.GetInterfaceResponse{
		Name: netIf,
	}, nil
}

// GetIPv4Addrs returns the IPv4 addresses for the network interface.
func (s *WifiService) GetIPv4Addrs(ctx context.Context, iface *network.GetIPv4AddrsRequest) (*network.GetIPv4AddrsResponse, error) {
	ifaceObj, err := net.InterfaceByName(iface.InterfaceName)
	if err != nil {
		return nil, err
	}

	addrs, err := ifaceObj.Addrs()
	if err != nil {
		return nil, err
	}

	var ret network.GetIPv4AddrsResponse

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ret.Ipv4 = append(ret.Ipv4, ipnet.IP.String())
		}
	}

	return &ret, nil
}

// hexSSID converts a SSID into the format of WiFi.HexSSID in shill.
// As in our tests, the SSID might contain non-ASCII characters, use WiFi.HexSSID
// field for better compatibility.
// Note: shill has the hex in upper case.
func (s *WifiService) hexSSID(ssid []byte) string {
	return strings.ToUpper(hex.EncodeToString(ssid))
}

// uint16sEqualUint32s returns true if a is equal to b.
func uint16sEqualUint32s(a []uint16, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })

	for i, v := range b {
		if v != uint32(a[i]) {
			return false
		}
	}
	return true
}

// ExpectWifiFrequencies checks if the device discovers the given SSID on the specific frequencies.
func (s *WifiService) ExpectWifiFrequencies(ctx context.Context, req *network.ExpectWifiFrequenciesRequest) (*empty.Empty, error) {
	ctx, st := timing.Start(ctx, "wifi_service.ExpectWifiFrequencies")
	defer st.End()
	testing.ContextLog(ctx, "ExpectWifiFrequencies: ", req)

	m, err := shill.NewManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a manager object")
	}

	hexSSID := s.hexSSID(req.Ssid)
	query := map[string]interface{}{
		shill.ServicePropertyType:        "wifi",
		shill.ServicePropertyWiFiHexSSID: hexSSID,
	}

	service, err := s.discoverService(ctx, m, query)
	if err != nil {
		return nil, err
	}

	// Spawn watcher for checking property change.
	pw, err := service.CreateWatcher(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher")
	}
	defer pw.Close(ctx)

	shortCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	for {
		props, err := service.GetProperties(shortCtx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get service properties")
		}
		freqs, err := props.GetUint16s(shill.ServicePropertyWiFiFrequencyList)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get property %s", shill.ServicePropertyWiFiFrequencyList)
		}
		if uint16sEqualUint32s(freqs, req.Frequencies) {
			testing.ContextLogf(shortCtx, "Got wanted frequencies %v for service with SSID: %s", req.Frequencies, req.Ssid)
			return &empty.Empty{}, nil
		}

		testing.ContextLogf(shortCtx, "Got frequencies %v for service with SSID: %s; want %v; waiting for update", freqs, req.Ssid, req.Frequencies)
		if _, err := pw.WaitAll(shortCtx, shill.ServicePropertyWiFiFrequencyList); err != nil {
			return nil, errors.Wrap(err, "failed to wait for the service property change")
		}
	}
}
