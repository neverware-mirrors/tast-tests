// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package assistant supports interaction with Assistant service.
package assistant

import (
	"context"
	"path/filepath"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/screenshot"
	"chromiumos/tast/testing"
)

// QueryResponse contains a subset of the results returned from the Assistant server
// when it received a query. This struct contains the only fields which are used in
// the tests.
type QueryResponse struct {
	// Contains the HTML string of the response.
	HTML string `json:"htmlResponse"`
}

// QueryStatus contains a subset of the status of an interaction with Assistant started
// by sending a query, e.g. query text, mic status, and query response. This struct contains
// the only fields which are used in the tests.
//
// TODO(meilinw): Add a reference for the struct after the API change landed (crrev.com/c/1552293).
type QueryStatus struct {
	QueryResponse `json:"queryResponse"`
}

// Enable brings up Google Assistant service and returns any errors.
func Enable(ctx context.Context, tconn *chrome.TestConn) error {
	return tconn.Call(ctx, nil, `tast.promisify(chrome.autotestPrivate.setAssistantEnabled)`, true, 10*1000 /* timeout_ms */)
}

// Disable stops the Google Assistant service and returns any errors.
func Disable(ctx context.Context, tconn *chrome.TestConn) error {
	return tconn.Call(ctx, nil, `tast.promisify(chrome.autotestPrivate.setAssistantEnabled)`, false, 10*1000 /* timeout_ms */)
}

// Cleanup stops the Google Assistant service so other tests are not impacted.
// If a failure happened, we make a screenshot beforehand so the Assistant UI
// is visible in the screenshot.
func Cleanup(ctx context.Context, hasError func() bool, cr *chrome.Chrome, tconn *chrome.TestConn) error {
	if hasError() {
		outDir, ok := testing.ContextOutDir(ctx)
		if !ok {
			return errors.New("outdir not available")
		}
		screenshot.CaptureChrome(ctx, cr, filepath.Join(outDir, "screenshot.png"))
	}

	return Disable(ctx, tconn)
}

// EnableAndWaitForReady brings up Google Assistant service, waits for
// NEW_READY signal and returns any errors.
func EnableAndWaitForReady(ctx context.Context, tconn *chrome.TestConn) error {
	return tconn.Call(ctx, nil, `tast.promisify(chrome.autotestPrivate.enableAssistantAndWaitForReady)`)
}

// SendTextQuery sends text query to Assistant and returns the query status.
func SendTextQuery(ctx context.Context, tconn *chrome.TestConn, query string) (QueryStatus, error) {
	var status QueryStatus
	err := tconn.Call(ctx, &status, `tast.promisify(chrome.autotestPrivate.sendAssistantTextQuery)`, query, 10*1000 /* timeout_ms */)
	return status, err
}

// SetHotwordEnabled turns on/off "OK Google" hotword detection for Assistant.
func SetHotwordEnabled(ctx context.Context, tconn *chrome.TestConn, enabled bool) error {
	return setPrefValue(ctx, tconn, "settings.voice_interaction.hotword.enabled", enabled)
}

// SetContextEnabled enables/disables the access to the screen context for Assistant.
// This pref corresponds to the "Related Info" setting of Assistant.
func SetContextEnabled(ctx context.Context, tconn *chrome.TestConn, enabled bool) error {
	return setPrefValue(ctx, tconn, "settings.voice_interaction.context.enabled", enabled)
}

// SetVoiceInteractionConsentValue enables/disables the consent value for Assistant voice interaction.
func SetVoiceInteractionConsentValue(ctx context.Context, tconn *chrome.TestConn, value int) error {
	return setPrefValue(ctx, tconn, "settings.voice_interaction.activity_control.consent_status", value)
}

// SetNumSessionsOnboardingShown enables/disables the Assistant onboarding feature by controlling the number
// of sessions where onboarding screen has shown. Note that true pref value will *not* be restored later, so
// tests that need this feature must explicitly enable it during setup.
func SetNumSessionsOnboardingShown(ctx context.Context, tconn *chrome.TestConn, value int) error {
	return setPrefValue(ctx, tconn, "ash.assistant.num_sessions_where_onboarding_shown", value)
}

// setPrefValue is a helper function to set value for Assistant related preferences.
func setPrefValue(ctx context.Context, tconn *chrome.TestConn, prefName string, value interface{}) error {
	return tconn.Call(ctx, nil, `tast.promisify(chrome.autotestPrivate.setWhitelistedPref)`, prefName, value)
}

// ToggleUIWithHotkey mimics the Assistant key press to open/close the Assistant UI.
func ToggleUIWithHotkey(ctx context.Context, tconn *chrome.TestConn) error {
	if err := tconn.Call(ctx, nil, `async () => {
		  let accel = {keyCode: 'assistant', shift: false, control: false, alt: false, search: false, pressed: true};
		  // Triggers the hotkey pressed event.
		  await tast.promisify(chrome.autotestPrivate.activateAccelerator)(accel);
		  // Releases the key for cleanup. This release event will not be handled as it is
		  // not a registered accelerator, so we ignore the result and don't wait for this
		  // async call returned.
		  accel.pressed = false;
		  chrome.autotestPrivate.activateAccelerator(accel, () => {});
		}`); err != nil {
		return errors.Wrap(err, "failed to execute accelerator")
	}

	return nil
}

// VerboseLogging is a helper function passed into chrome.New which will:
//     - Enable VLOG traces in the assistant code.
//     - Use LibAssistant beta backend which will cause LibAssistant to print
//       more detailed logs.
func VerboseLogging() chrome.Option {
	return chrome.ExtraArgs(
		"--vmodule=*assistant*=3,chromeos/services/assistant/service=3",
		"--enable-features=LibAssistantBetaBackend",
	)
}

// VerboseLoggingEnabled creates a new precondition which can be shared by
// tests that require an already-started Chromeobject with verbose logging
// enabled.
func VerboseLoggingEnabled() testing.Precondition { return verboseLoggingPre }

var verboseLoggingPre = chrome.NewPrecondition("verbose-logging", VerboseLogging())
