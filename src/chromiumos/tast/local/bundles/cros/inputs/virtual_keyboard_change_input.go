// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inputs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ime"
	"chromiumos/tast/local/chrome/ui"
	"chromiumos/tast/local/chrome/ui/faillog"
	"chromiumos/tast/local/chrome/vkb"
	"chromiumos/tast/local/input"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         VirtualKeyboardChangeInput,
		Desc:         "Checks that changing input method in different ways",
		Contacts:     []string{"shend@chromium.org", "essential-inputs-team@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"chrome", "google_virtual_keyboard"},
		Timeout:      3 * time.Minute,
	})
}

func VirtualKeyboardChangeInput(ctx context.Context, s *testing.State) {
	cr, err := chrome.New(ctx, chrome.ExtraArgs("--enable-virtual-keyboard"), chrome.ExtraArgs("--force-tablet-mode=touch_view"))
	if err != nil {
		s.Fatal("Failed to start Chrome: ", err)
	}
	defer cr.Close(ctx)

	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create test API connection: ", err)
	}

	defer faillog.DumpUITreeOnError(ctx, s.OutDir(), s.HasError, tconn)

	const (
		defaultInputMethod      = "xkb:us::eng"
		defaultInputMethodLabel = "US keyboard"
		language                = "fr-FR"
		inputMethod             = "xkb:fr::fra"
	)

	if err := ime.EnableLanguage(ctx, tconn, language); err != nil {
		s.Fatalf("Failed to enable the language %q: %v", language, err)
	}

	if err := ime.AddInputMethod(ctx, tconn, vkb.ImePrefix+inputMethod); err != nil {
		s.Fatalf("Failed to enable the IME %q: %v", inputMethod, err)
	}

	s.Log("Start a local server to test chrome")
	const identifier = "e14s-inputbox"
	html := fmt.Sprintf(`<input type="text" id="text" autocorrect="off" aria-label=%q/>`, identifier)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		io.WriteString(w, html)
	}))
	defer server.Close()

	conn, err := cr.NewConn(ctx, server.URL)
	if err != nil {
		s.Fatal("Failed to render test page: ", err)
	}
	defer conn.Close()

	inputFieldElement, err := ui.FindWithTimeout(ctx, tconn, ui.FindParams{Name: identifier}, 5*time.Second)
	if err != nil {
		s.Fatalf("Failed to find input element %s: %v", identifier, err)
	}
	defer inputFieldElement.Release(ctx)

	s.Log("Click input field to trigger virtual keyboard")
	if err := inputFieldElement.LeftClick(ctx); err != nil {
		s.Fatal("Failed to click the input element: ", err)
	}

	s.Log("Waiting for the virtual keyboard to show")
	if err := vkb.WaitUntilShown(ctx, tconn); err != nil {
		s.Fatal("Failed to wait for the virtual keyboard to show: ", err)
	}

	assertInputMethod := func(ctx context.Context, inputMethod string) {
		s.Logf("Wait for current input method to be %q", inputMethod)
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			currentInputMethod, err := vkb.GetCurrentInputMethod(ctx, tconn)
			if err != nil {
				return errors.Wrap(err, "failed to get current input method")
			} else if currentInputMethod != vkb.ImePrefix+inputMethod {
				return errors.Errorf("failed to verify current input method. got %q; want %q", currentInputMethod, vkb.ImePrefix+inputMethod)
			}
			return nil
		}, &testing.PollOptions{Timeout: 30 * time.Second}); err != nil {
			s.Fatal("Failed to assert input method: ", err)
		}
	}

	// Assert default input method.
	assertInputMethod(ctx, defaultInputMethod)

	s.Log("Switch input method with keybaord shortcut Ctrl+Shift+Space")
	keyboard, err := input.Keyboard(ctx)
	if err != nil {
		s.Fatal("Failed to get keyboard: ", err)
	}
	defer keyboard.Close()

	if err := keyboard.Accel(ctx, "Ctrl+Shift+Space"); err != nil {
		s.Fatal("Accel(Ctrl+Shift+Space) failed: ", err)
	}

	// Assert new input method after switching with keyboard shortcut.
	assertInputMethod(ctx, inputMethod)

	s.Log("Switch input method on virtual keyboard")
	if err := vkb.TapKey(ctx, tconn, "open keyboard menu"); err != nil {
		s.Fatal("Failed to click language menu on virtual keyboard: ", err)
	}

	// Wait for language option menu full positioned.
	if err := ui.WaitForLocationChangeCompleted(ctx, tconn); err != nil {
		s.Fatal("Failed to wait for language option menu positioned: ", err)
	}

	languageOptionParams := ui.FindParams{
		Name: defaultInputMethodLabel,
	}

	languageOption, err := ui.FindWithTimeout(ctx, tconn, languageOptionParams, 10*time.Second)
	if err != nil {
		s.Fatalf("Failed to find language option with %v: %v", languageOptionParams, err)
	}
	defer languageOption.Release(ctx)

	if err := languageOption.LeftClick(ctx); err != nil {
		s.Fatal("Failed to click default language: ", err)
	}

	assertInputMethod(ctx, defaultInputMethod)
}