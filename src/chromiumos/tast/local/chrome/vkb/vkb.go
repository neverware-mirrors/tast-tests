// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package vkb contains shared code to interact with the virtual keyboard.
package vkb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mafredri/cdp/protocol/target"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/testing"
)

const imePrefix = "_comp_ime_jkghodnilhceideoidjikpgommlajknk"

// ShowVirtualKeyboard forces the virtual keyboard to open.
func ShowVirtualKeyboard(ctx context.Context, tconn *chrome.TestConn) error {
	return tconn.EvalPromise(ctx, `tast.promisify(chrome.inputMethodPrivate.showInputView)()`, nil)
}

// HideVirtualKeyboard forces the virtual keyboard to be hidden.
func HideVirtualKeyboard(ctx context.Context, tconn *chrome.TestConn) error {
	return tconn.EvalPromise(ctx, `tast.promisify(chrome.inputMethodPrivate.hideInputView)()`, nil)
}

// VirtualKeyboard returns a reference to chrome.automation API AutomationNode of virtual keyboard.
func VirtualKeyboard(ctx context.Context, tconn *chrome.TestConn) (*ui.Node, error) {
	params := ui.FindParams{
		Role: ui.RoleTypeKeyboard,
	}
	return ui.FindWithTimeout(ctx, tconn, params, time.Second)
}

// SetCurrentInputMethod sets the current input method used by the virtual
// keyboard.
func SetCurrentInputMethod(ctx context.Context, tconn *chrome.TestConn, inputMethod string) error {
	if err := tconn.EvalPromise(ctx, fmt.Sprintf(`
		new Promise((resolve, reject) => {
			chrome.autotestPrivate.setWhitelistedPref(
				'settings.language.preload_engines', %[1]q, () => {
					chrome.inputMethodPrivate.setCurrentInputMethod(%[1]q, () => {
						if (chrome.runtime.lastError) {
							reject(chrome.runtime.lastError.message);
						} else {
							resolve();
						}
					});
				}
			);
		})
		`, imePrefix+inputMethod), nil); err != nil {
		return errors.Wrapf(err, "failed to set current input method: %q", inputMethod)
	}

	// Change language via tconn requiring a few seconds to install.
	// TODO(b/157686038): Use API to identify completion of changing language
	testing.Sleep(ctx, 3*time.Second)
	return nil
}

// IsShown checks if the virtual keyboard is currently shown. It checks whether
// there is a visible DOM element with an accessibility role of "keyboard".
func IsShown(ctx context.Context, tconn *chrome.TestConn) (shown bool, err error) {
	params := ui.FindParams{
		Role:  ui.RoleTypeKeyboard,
		State: map[ui.StateType]bool{ui.StateTypeInvisible: false},
	}
	return ui.Exists(ctx, tconn, params)
}

// waitUntil checks if the virtual keyboard visibility matches the given expectation.
func waitUntil(ctx context.Context, tconn *chrome.TestConn, expected bool) error {
	expectedState := "shown"
	if !expected {
		expectedState = "hidden"
	}

	if err := testing.Poll(ctx, func(ctx context.Context) error {
		if shown, err := IsShown(ctx, tconn); err != nil {
			return testing.PollBreak(err)
		} else if shown != expected {
			return errors.Errorf("waiting for virtual keyboard to be %q", expectedState)
		}
		return nil
	}, nil); err != nil {
		return errors.Wrapf(err, "failed to wait for virtual keyboard to be %q", expectedState)
	}
	return nil
}

// WaitUntilShown waits for the virtual keyboard to appear. It waits until there
// is a visible DOM element with accessibility role of "keyboard".
func WaitUntilShown(ctx context.Context, tconn *chrome.TestConn) error {
	return waitUntil(ctx, tconn, true)
}

// WaitUntilHidden waits for the virtual keyboard to hide. It waits until there
// is no visible DOM element with accessibility role of "keyboard".
func WaitUntilHidden(ctx context.Context, tconn *chrome.TestConn) error {
	return waitUntil(ctx, tconn, false)
}

// WaitUntilButtonsRender waits for the virtual keyboard to render some buttons.
func WaitUntilButtonsRender(ctx context.Context, tconn *chrome.TestConn) error {
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		keyboard, err := ui.Find(ctx, tconn, ui.FindParams{Role: ui.RoleTypeKeyboard})
		if err != nil {
			return errors.Wrap(err, "virtual keyboard does not exist yet")
		}
		defer keyboard.Release(ctx)
		keys, err := keyboard.Descendants(ctx, ui.FindParams{Role: ui.RoleTypeButton})
		if err != nil {
			return errors.Wrap(err, "keyboard buttons don't exist yet")
		}
		defer keys.Release(ctx)
		// English keyboard should have at least 26 keys.
		if len(keys) <= 26 {
			return errors.New("not all buttons have rendered yet")
		}
		return nil
	}, nil); err != nil {
		return errors.Wrap(err, "failed to wait for virtual keyboad buttons to render")
	}
	return nil
}

// UIConn returns a connection to the virtual keyboard HTML page,
// where JavaScript can be executed to simulate interactions with the UI.
// The connection is lazily created, and this function will block until the
// extension is loaded or ctx's deadline is reached. The caller should close
// the returned connection.
func UIConn(ctx context.Context, c *chrome.Chrome) (*chrome.Conn, error) {
	extURLPrefix := "chrome-extension://jkghodnilhceideoidjikpgommlajknk/inputview.html"
	f := func(t *target.Info) bool { return strings.HasPrefix(t.URL, extURLPrefix) }
	return c.NewConnForTarget(ctx, f)
}

// TapKey simulates a tap event on the middle of the specified key via touch event. The key can
// be any letter of the alphabet, "space" or "backspace".
func TapKey(ctx context.Context, tconn *chrome.TestConn, keyName string) error {
	vkNode, err := VirtualKeyboard(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to find virtual keyboad automation node")
	}

	keyParams := ui.FindParams{
		Role: ui.RoleTypeButton,
		Name: keyName,
	}

	keyNode, err := vkNode.Descendant(ctx, keyParams)
	if err != nil {
		return errors.Wrapf(err, "failed to find key with %v", keyParams)
	}

	if err := keyNode.LeftClick(ctx); err != nil {
		return errors.Wrapf(err, "failed to click key %s", keyName)
	}
	return nil
}

// TapKeyJS simulates a tap event on the middle of the specified key via javascript. The key can
// be any letter of the alphabet, "space" or "backspace".
func TapKeyJS(ctx context.Context, kconn *chrome.Conn, key string) error {
	return kconn.Eval(ctx, fmt.Sprintf(`
	(() => {
		// Multiple keys can have the same aria label but only one is visible.
		const keys = document.querySelectorAll('[aria-label=%[1]q]')
		if (!keys) {
			throw new Error('Key %[1]q not found. No element with aria-label %[1]q.');
		}
		for (const key of keys) {
			const rect = key.getBoundingClientRect();
			if (rect.width <= 0 || rect.height <= 0) {
				continue;
			}
			const e = new Event('pointerdown');
			e.clientX = rect.x + rect.width / 2;
			e.clientY = rect.y + rect.height / 2;
			key.dispatchEvent(e);
			key.dispatchEvent(new Event('pointerup'));
			return;
		}
		throw new Error('Key %[1]q not clickable. Found elements with aria-label %[1]q, but they were not visible.');
	})()
	`, key), nil)
}

// SwitchToFloatMode changes virtual keyboard to floating layout.
func SwitchToFloatMode(ctx context.Context, tconn *chrome.TestConn) error {
	return TapKey(ctx, tconn, "make virtual keyboard movable")
}

// TapKeys simulates tap events on the middle of the specified sequence of keys via touch event.
// Each keys can be any letter of the alphabet, "space" or "backspace".
func TapKeys(ctx context.Context, tconn *chrome.TestConn, keys []string) error {
	for _, key := range keys {
		if err := TapKey(ctx, tconn, key); err != nil {
			return err
		}
		testing.Sleep(ctx, 50*time.Millisecond)
	}
	return nil
}

// TapKeysJS simulates tap events on the middle of the specified sequence of keys via javascript.
// Each keys can be any letter of the alphabet, "space" or "backspace".
func TapKeysJS(ctx context.Context, kconn *chrome.Conn, keys []string) error {
	for _, key := range keys {
		if err := TapKeyJS(ctx, kconn, key); err != nil {
			return err
		}
		testing.Sleep(ctx, 50*time.Millisecond)
	}
	return nil
}

// GetSuggestions returns suggestions that are currently displayed by the
// virtual keyboard.
func GetSuggestions(ctx context.Context, kconn *chrome.Conn) ([]string, error) {
	var suggestions []string
	err := kconn.Eval(ctx, `
	(() => {
		const elems = document.querySelectorAll('.candidate-span');
		return Array.prototype.map.call(elems, x => x.textContent);
	})()
`, &suggestions)
	return suggestions, err
}

// InputWithVirtualKeyboard waits for virtual keyboard shown up, types given key series and hide keyboard after.
func InputWithVirtualKeyboard(ctx context.Context, tconn *chrome.TestConn, keys []string) error {
	if err := WaitUntilShown(ctx, tconn); err != nil {
		return errors.Wrap(err, "failed to wait for the virtual keyboard to show")
	}

	if err := WaitUntilButtonsRender(ctx, tconn); err != nil {
		return errors.Wrap(err, "failed to wait for the virtual keyboard to render")
	}

	if err := TapKeys(ctx, tconn, keys); err != nil {
		return errors.Wrapf(err, "failed to tap keys %v: %v", keys, err)
	}

	if err := HideVirtualKeyboard(ctx, tconn); err != nil {
		return errors.Wrap(err, "failed to hide the virtual keyboard")
	}
	return nil
}
