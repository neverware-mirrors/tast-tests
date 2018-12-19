// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package ownership provides utilities to run ownership API related tests.
package ownership

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/pkcs12"

	"chromiumos/policy/enterprise_management"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/session"
)

// ExtractPrivKey reads a PKCS #12 format file at path, then extracts and
// returns RSA private key.
func ExtractPrivKey(path string) (*rsa.PrivateKey, error) {
	p12, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", path)
	}
	key, _, err := pkcs12.Decode(p12, "" /* password */)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode p12 file")
	}
	privKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("RSA private key is not found")
	}
	return privKey, nil
}

// BuildTestSettings returns the ChromeDeviceSettingsProto instance which
// can be used for testing settings.
func BuildTestSettings(user string) *enterprise_management.ChromeDeviceSettingsProto {
	boolTrue := true
	boolFalse := false
	return &enterprise_management.ChromeDeviceSettingsProto{
		GuestModeEnabled: &enterprise_management.GuestModeEnabledProto{
			GuestModeEnabled: &boolFalse,
		},
		ShowUserNames: &enterprise_management.ShowUserNamesOnSigninProto{
			ShowUserNames: &boolTrue,
		},
		DataRoamingEnabled: &enterprise_management.DataRoamingEnabledProto{
			DataRoamingEnabled: &boolTrue,
		},
		AllowNewUsers: &enterprise_management.AllowNewUsersProto{
			AllowNewUsers: &boolFalse,
		},
		UserWhitelist: &enterprise_management.UserWhitelistProto{
			UserWhitelist: []string{user, "a@b.c"},
		},
	}
}

// StoreSettings requests given SessionManager to store the
// ChromeDeviceSettingsProto data for the user with key.
func StoreSettings(ctx context.Context, sm *session.SessionManager, user string, key *rsa.PrivateKey, s *enterprise_management.ChromeDeviceSettingsProto) error {
	sdata, err := proto.Marshal(s)
	if err != nil {
		return errors.Wrap(err, "failed to serialize settings")
	}
	polType := "google/chromeos/device"
	pol := &enterprise_management.PolicyData{
		PolicyType:  &polType,
		Username:    &user,
		PolicyValue: sdata,
	}
	polData, err := proto.Marshal(pol)
	if err != nil {
		return errors.Wrap(err, "failed to serialize policy")
	}
	polSign, err := sign(key, polData)
	if err != nil {
		return errors.Wrap(err, "failed to sign policy data")
	}

	pubDer, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return errors.Wrap(err, "failed to marshal public key to DER")
	}
	pubSign, err := sign(key, pubDer)
	if err != nil {
		return errors.Wrap(err, "failed to serialize public key")
	}

	response := &enterprise_management.PolicyFetchResponse{
		PolicyData:            polData,
		PolicyDataSignature:   polSign,
		NewPublicKey:          pubDer,
		NewPublicKeySignature: pubSign,
	}

	// Send the data to session_manager.
	w, err := sm.WatchPropertyChangeComplete(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start watching PropertyChangeComplete signal")
	}
	defer w.Close(ctx)
	if err := sm.StorePolicy(ctx, response); err != nil {
		return errors.Wrap(err, "failed to call StorePolicy")
	}
	select {
	case <-w.Signals:
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "timed out waiting for PropertyChangeComplete signal")
	}

	return nil
}

// sign signs the blob with the given key, and returns the signature.
func sign(key *rsa.PrivateKey, blob []byte) ([]byte, error) {
	h := sha1.New()
	h.Write(blob)
	digest := h.Sum(nil)
	return rsa.SignPKCS1v15(nil, key, crypto.SHA1, digest)
}

// RetrieveSettings requests to given SessionManager to return the currently
// stored ChromeDeviceSettingsProto.
func RetrieveSettings(ctx context.Context, sm *session.SessionManager) (*enterprise_management.ChromeDeviceSettingsProto, error) {
	ret, err := sm.RetrievePolicy(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve policy")
	}

	rPol := &enterprise_management.PolicyData{}
	if err = proto.Unmarshal(ret.PolicyData, rPol); err != nil {
		return nil, errors.Wrap(err, "failed to parse PolicyData")
	}

	rsettings := &enterprise_management.ChromeDeviceSettingsProto{}
	if err = proto.Unmarshal(rPol.PolicyValue, rsettings); err != nil {
		return nil, errors.Wrap(err, "failed to parse PolicyValue")
	}
	return rsettings, nil
}
