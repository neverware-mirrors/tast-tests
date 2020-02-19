// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package accessibility provides functions to assist with interacting with accessibility settings
// in ARC accessibility tests.
package accessibility

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"chromiumos/tast/ctxutil"
	"chromiumos/tast/errors"
	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/audio"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/input"
	"chromiumos/tast/local/screenshot"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/testing"
)

const (
	// ArcAccessibilityHelperService is the full class name of ArcAccessibilityHelperService.
	ArcAccessibilityHelperService = "org.chromium.arc.accessibilityhelper/org.chromium.arc.accessibilityhelper.ArcAccessibilityHelperService"

	// ApkName is the name of apk which is used in ARC++ accessibility tests.
	ApkName      = "ArcAccessibilityTest.apk"
	packageName  = "org.chromium.arc.testapp.accessibilitytest"
	activityName = ".AccessibilityActivity"

	// TODO(b/149791978): Remove extOldURL after crrev.com/c/2051037 merged into ChromeOS.
	extURL    = "chrome-extension://mndnfokpggljbaajbnioimlmbfngpief/chromevox/background/background.html"
	extOldURL = "chrome-extension://mndnfokpggljbaajbnioimlmbfngpief/background/background.html"

	// CheckBox class name.
	CheckBox = "android.widget.CheckBox"
	// EditText class name.
	EditText = "android.widget.EditText"
	// SeekBar class name.
	SeekBar = "android.widget.SeekBar"
	// TextView class name.
	TextView = "android.widget.TextView"
	// ToggleButton class name.
	ToggleButton = "android.widget.ToggleButton"
)

// Feature represents an accessibility feature in Chrome OS.
type Feature string

// List of accessibility features that interacts with ARC.
const (
	SpokenFeedback Feature = "spokenFeedback"
	SwitchAccess           = "switchAccess"
	SelectToSpeak          = "selectToSpeak"
	FocusHighlight         = "focusHighlight"
)

// AutomationNode represents an accessibility struct, which contains properties from chrome.automation.Automation.
// This is defined at:
// https://developer.chrome.com/extensions/automation#type-AutomationNode
// Only the properties which are used in tast tests are defined here.
type AutomationNode struct {
	ClassName     string
	Checked       string
	Tooltip       string
	ValueForRange int
}

// FocusedNode returns the currently focused node of ChromeVox.
// chrome.automation.AutomationNode properties are defined using getters
// see: https://cs.chromium.org/chromium/src/extensions/renderer/resources/automation/automation_node.js?q=automationNode&sq=package:chromium&dr=CSs&l=1218
// meaning that resolve(node) cannot be applied here.
func FocusedNode(ctx context.Context, chromeVoxConn *chrome.Conn) (*AutomationNode, error) {
	var automationNode AutomationNode
	const script = `new Promise((resolve, reject) => {
				chrome.automation.getFocus((node) => {
					resolve({
						Checked: node.checked,
						ClassName: node.className,
						Tooltip: node.tooltip,
						ValueForRange: node.valueForRange
					});
				});
			})`
	if err := chromeVoxConn.EvalPromise(ctx, script, &automationNode); err != nil {
		return nil, err
	}
	return &automationNode, nil
}

// IsEnabledAndroid checks if accessibility is enabled in Android.
func IsEnabledAndroid(ctx context.Context, a *arc.ARC) (bool, error) {
	res, err := a.Command(ctx, "settings", "--user", "0", "get", "secure", "accessibility_enabled").Output(testexec.DumpLogOnError)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(res)) == "1", nil
}

// EnabledAndroidAccessibilityServices returns enabled AccessibilityService in Android.
func EnabledAndroidAccessibilityServices(ctx context.Context, a *arc.ARC) ([]string, error) {
	res, err := a.Command(ctx, "settings", "--user", "0", "get", "secure", "enabled_accessibility_services").Output(testexec.DumpLogOnError)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(res)), ":"), nil
}

// chromeVoxExtConn returns a connection to the ChromeVox extension's background page.
// If the extension is not ready, the connection will be closed before returning.
// Otherwise the calling function will close the connection.
func chromeVoxExtConn(ctx context.Context, c *chrome.Chrome) (*chrome.Conn, error) {
	testing.ContextLogf(ctx, "Waiting for ChromeVox background page at %s or %s", extURL, extOldURL)
	extConn, err := c.NewConnForTarget(ctx, func(t *chrome.Target) bool {
		return t.URL == extURL || t.URL == extOldURL
	})
	if err != nil {
		return nil, err
	}

	// Ensure that we don't attempt to use the extension before its APIs are
	// available: https://crbug.com/789313.
	if err := extConn.WaitForExpr(ctx, "ChromeVoxState.instance"); err != nil {
		extConn.Close()
		return nil, errors.Wrap(err, "ChromeVox unavailable")
	}

	testing.ContextLog(ctx, "Extension is ready")
	return extConn, nil
}

// SetFeatureEnabled sets the specified accessibility feature enabled/disabled using the provided connection to the extension.
func SetFeatureEnabled(ctx context.Context, conn *chrome.Conn, feature Feature, enable bool) error {
	script := fmt.Sprintf(
		`new Promise((resolve, reject) => {
			chrome.accessibilityFeatures.%s.set({value: %t}, resolve);
		})`, feature, enable)
	if err := conn.EvalPromise(ctx, script, nil); err != nil {
		return errors.Wrapf(err, "failed to toggle %v to %t", feature, enable)
	}
	return nil
}

// waitForSpokenFeedbackReady enables spoken feedback.
// A connection to the ChromeVox extension background page is returned, and this will be
// closed by the calling function.
func waitForSpokenFeedbackReady(ctx context.Context, cr *chrome.Chrome, a *arc.ARC) (*chrome.Conn, error) {
	// Wait until spoken feedback is enabled in Android side.
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		if res, err := IsEnabledAndroid(ctx, a); err != nil {
			return testing.PollBreak(errors.Wrap(err, "failed to check whether accessibility is enabled in Android"))
		} else if !res {
			return errors.New("accessibility not enabled")
		}
		return nil
	}, &testing.PollOptions{Timeout: 30 * time.Second}); err != nil {
		return nil, errors.Wrap(err, "failed to ensure accessibility is enabled: ")
	}

	chromeVoxConn, err := chromeVoxExtConn(ctx, cr)
	if err != nil {
		return nil, errors.Wrap(err, "creating connection to ChromeVox extension failed: ")
	}

	// Poll until ChromeVox connection finishes loading.
	if err := chromeVoxConn.WaitForExpr(ctx, `document.readyState === "complete"`); err != nil {
		return nil, errors.Wrap(err, "timed out waiting for ChromeVox connection to be ready")
	}

	testing.ContextLog(ctx, "ChromeVox is ready")
	return chromeVoxConn, nil
}

// WaitForFocusedNode polls until the properties of the focused node matches node.
// Returns an error after 30 seconds.
func WaitForFocusedNode(ctx context.Context, chromeVoxConn *chrome.Conn, node *AutomationNode) error {
	// Wait for focusClassName to receive focus.
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		focusedNode, err := FocusedNode(ctx, chromeVoxConn)
		if err != nil {
			return testing.PollBreak(err)
		} else if !cmp.Equal(focusedNode, node, cmpopts.EquateEmpty()) {
			return errors.Errorf("focused node is incorrect: got %q, want %q", focusedNode, node)
		}
		return nil
	}, &testing.PollOptions{Timeout: 30 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to get current focus")
	}
	return nil
}

// WaitForChromeVoxStopSpeaking polls until ChromeVox TTS has stoped speaking.
func WaitForChromeVoxStopSpeaking(ctx context.Context, chromeVoxConn *chrome.Conn) error {
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		var isSpeaking bool
		if err := chromeVoxConn.Eval(ctx, "ChromeVox.tts.isSpeaking()", &isSpeaking); err != nil {
			return testing.PollBreak(err)
		}
		if isSpeaking {
			return errors.New("ChromeVox is speaking")
		}
		return nil
	}, &testing.PollOptions{Timeout: 30 * time.Second}); err != nil {
		return errors.Wrap(err, "timed out waiting for ChromeVox to finish speaking")
	}
	return nil
}

// RunTest installs the ArcAccessibilityTestApplication, launches it, and waits
// for ChromeVox to be ready.
func RunTest(ctx context.Context, s *testing.State, f func(context.Context, *arc.ARC, *chrome.Conn, *input.KeyboardEventWriter) error) {
	fullCtx := ctx
	ctx, cancel := ctxutil.Shorten(fullCtx, 10*time.Second)
	defer cancel()

	if err := audio.Mute(ctx); err != nil {
		s.Fatal("Failed to mute device: ", err)
	}
	defer audio.Unmute(fullCtx)

	d := s.PreValue().(arc.PreData)
	a := d.ARC
	cr := d.Chrome

	conn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Creating test API connection failed: ", err)
	}

	// It takes some time for ArcServiceManager to be ready.
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		if err := conn.EvalPromise(ctx, "tast.promisify(chrome.autotestPrivate.setArcTouchMode)(true)", nil); err != nil {
			return err
		}
		return nil
	}, &testing.PollOptions{Timeout: 10 * time.Second}); err != nil {
		s.Fatal("Timed out waiting for touch mode: ", err)
	}

	if err := SetFeatureEnabled(ctx, conn, SpokenFeedback, true); err != nil {
		s.Fatal("Failed to enable spoken feedback: ", err)
	}
	defer func() {
		if err := SetFeatureEnabled(fullCtx, conn, SpokenFeedback, false); err != nil {
			s.Fatal("Failed to disable spoken feedback: ", err)
		}
	}()

	chromeVoxConn, err := waitForSpokenFeedbackReady(ctx, cr, a)
	if err != nil {
		s.Fatal(err) // NOLINT: arc/ui returns loggable errors
	}
	defer chromeVoxConn.Close()

	if err := a.Install(ctx, s.DataPath(ApkName)); err != nil {
		s.Fatal("Failed installing app: ", err)
	}

	act, err := arc.NewActivity(a, packageName, activityName)
	if err != nil {
		s.Fatal("Failed to create new activity: ", err)
	}
	defer act.Close()

	if err := act.Start(ctx); err != nil {
		s.Fatal("Failed to start activity: ", err)
	}

	if err := act.WaitForResumed(ctx, 10*time.Second); err != nil {
		s.Fatal("Failed to wait for activity to resume: ", err)
	}

	if err := WaitForChromeVoxStopSpeaking(ctx, chromeVoxConn); err != nil {
		s.Fatal("Failed to wait for ChromeVox finish speaking: ", err)
	}

	ew, err := input.Keyboard(ctx)
	if err != nil {
		s.Fatal("Error with creating EventWriter from keyboard: ", err)
	}
	defer ew.Close()

	if err := f(ctx, a, chromeVoxConn, ew); err != nil {
		// TODO(crbug.com/1044446): Take faillog on testing.State.Fatal() invocation.
		path := filepath.Join(s.OutDir(), "screenshot-with-chromevox.png")
		if err := screenshot.CaptureChrome(ctx, cr, path); err != nil {
			s.Log("Failed to capture screenshot: ", err)
		}
		s.Fatal("Failed to run the test: ", err)
	}
}
