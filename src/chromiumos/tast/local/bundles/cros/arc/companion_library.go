// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package arc

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/arc"
	"chromiumos/tast/local/arc/ui"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ash"
	"chromiumos/tast/local/chrome/display"
	"chromiumos/tast/local/input"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         CompanionLibrary,
		Desc:         "Test all ARC++ companion library",
		Contacts:     []string{"sstan@google.com", "arc-framework+tast@google.com"},
		Attr:         []string{"group:mainline", "informational"},
		SoftwareDeps: []string{"android_p", "chrome"},
		Data:         []string{"ArcCompanionLibDemo.apk"},
		Pre:          arc.Booted(),
		Timeout:      5 * time.Minute,
	})
}

const pkg = "org.chromium.arc.companionlibdemo"

type companionLibMessage struct {
	MessageID int    `json:"mid"`
	Type      string `json:"type"`
	API       string `json:"api"`
	LogMsg    *struct {
		Msg string `json:"msg"`
	} `json:"LogMsg"`
	CaptionHeightMsg *struct {
		CaptionHeight int `json:"caption_height"`
	} `json:"CaptionHeightMsg"`
	DeviceModeMsg *struct {
		DeviceMode string `json:"device_mode"`
	} `json:"DeviceModeMsg"`
	WorkspaceInsetMsg *struct {
		InsetBound string `json:"inset_bound"`
	} `json:"WorkspaceInsetMsg"`
}

func CompanionLibrary(ctx context.Context, s *testing.State) {
	const (
		apk = "ArcCompanionLibDemo.apk"

		mainActivity     = ".MainActivity"
		resizeActivityID = ".MoveResizeActivity"
	)

	cr := s.PreValue().(arc.PreData).Chrome

	tconn, err := cr.TestAPIConn(ctx)
	if err != nil {
		s.Fatal("Failed to create Test API connection: ", err)
	}

	tabletModeEnabled, err := ash.TabletModeEnabled(ctx, tconn)
	if err != nil {
		s.Fatal("Failed to get tablet mode: ", err)
	}
	// Restore tablet mode to its original state on exit.
	defer ash.SetTabletModeEnabled(ctx, tconn, tabletModeEnabled)

	// Force Chrome to be in clamshell mode, where windows are resizable.
	if err := ash.SetTabletModeEnabled(ctx, tconn, false); err != nil {
		s.Fatal("Failed to disable tablet mode: ", err)
	}

	a := s.PreValue().(arc.PreData).ARC
	if err := a.Install(ctx, s.DataPath(apk)); err != nil {
		s.Fatal("Failed installing app: ", err)
	}

	act, err := arc.NewActivity(a, pkg, mainActivity)
	if err != nil {
		s.Fatal("Failed to create new activity: ", err)
	}
	defer act.Close()

	if err := act.Start(ctx); err != nil {
		s.Fatal("Failed start Settings activity: ", err)
	}

	d, err := ui.NewDevice(ctx, a)
	if err != nil {
		s.Fatal("Failed to get device: ", err)
	}
	defer d.Close()

	if err := act.WaitForResumed(ctx, time.Second); err != nil {
		s.Fatal("Failed to wait for activity to resume: ", err)
	}

	// All of tests in this block running on MainActivity.
	type testFunc func(context.Context, *chrome.Conn, *arc.Activity, *ui.Device) error
	for _, test := range []struct {
		name string
		fn   testFunc
	}{
		{"Window State", testWindowState},
		{"Get Workspace Insets", testWorkspaceInsets},
		{"Caption Button", testCaptionButton},
		{"Get Device Mode", testDeviceMode},
		{"Get Caption Height", testCaptionHeight},
	} {
		s.Logf("Running %q", test.name)
		if err := act.Start(ctx); err != nil {
			s.Fatal("Failed to start context: ", err)
		}
		if err := act.WaitForResumed(ctx, time.Second); err != nil {
			s.Fatal("Failed to wait for activity to resuyme: ", err)
		}
		if err := test.fn(ctx, tconn, act, d); err != nil {
			s.Errorf("%s test failed: %v", test.name, err)
		}
		if err := act.Stop(ctx); err != nil {
			s.Fatal("Failed to stop context: ", err)
		}
	}

	resizeAct, err := arc.NewActivity(a, pkg, resizeActivityID)
	if err != nil {
		s.Fatal("Could not create ResizeActivity: ", err)
	}
	if err := resizeAct.Start(ctx); err != nil {
		s.Fatal("Could not start ResizeActivity: ", err)
	}
	defer func() {
		if err := resizeAct.Stop(ctx); err != nil {
			s.Fatal("Could not stop resize activity: ", err)
		}
	}()
	if err := setWindowStateSync(ctx, resizeAct, arc.WindowStateNormal); err != nil {
		s.Fatal("Could not set window normal state: ", err)
	}
	if err := testResizeWindow(ctx, tconn, resizeAct, d); err != nil {
		s.Error("Move & Resize Window test failed: ", err)
	}
	// End of function, i.e. no resizeAct.Stop() call here, because it is called in defer.
}

// testCaptionHeight verifies that the caption height length getting from ChromeOS companion library is correct.
func testCaptionHeight(ctx context.Context, tconn *chrome.Conn, act *arc.Activity, d *ui.Device) error {
	const getCaptionHeightButtonID = pkg + ":id/get_caption_height"

	dispMode, err := ash.InternalDisplayMode(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to get display mode")
	}

	// Read JSON format window caption height infomation.
	baseMessage, err := getLastJSONMessage(ctx, d)
	if err != nil {
		return errors.Wrap(err, "failed to get base json message")
	}
	if err := d.Object(ui.ID(getCaptionHeightButtonID)).Click(ctx); err != nil {
		return errors.Wrap(err, "failed to click Get Caption Height button")
	}
	var msg *companionLibMessage
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		var err error
		msg, err = getLastJSONMessage(ctx, d)
		if err != nil {
			return testing.PollBreak(err)
		}
		// Waiting for new message coming
		if baseMessage.MessageID == msg.MessageID {
			return errors.New("still waiting the new json message")
		}
		return nil
	}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to get new message of caption height")
	}
	if msg.CaptionHeightMsg == nil {
		return errors.Errorf("unexpected JSON message format: no CaptionHeightMsg; got %v", msg)
	}

	appWindow, err := ash.GetARCAppWindowInfo(ctx, tconn, pkg)
	if err != nil {
		return errors.Wrap(err, "failed to get arc app window")
	}

	actualHeight := int(math.Round(float64(appWindow.CaptionHeight) * dispMode.DeviceScaleFactor))
	if actualHeight != msg.CaptionHeightMsg.CaptionHeight {
		return errors.Errorf("wrong caption height: got %v, want %v", msg.CaptionHeightMsg.CaptionHeight, actualHeight)
	}
	return nil

}

// testResizeWindow verifies that the resize function in ChromeOS companion library works as expected.
// ARC companion library demo provide a activity for resize test, there are four draggable hit-boxes in four sides.
// The test maximizing the window by drag from four side inner hit-boxes. The events will be handled by Companion Library, not Chrome.
func testResizeWindow(ctx context.Context, tconn *chrome.Conn, act *arc.Activity, d *ui.Device) error {
	dispMode, err := ash.InternalDisplayMode(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to get display mode")
	}
	dispInfo, err := display.GetInternalInfo(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to get internal display info")
	}
	appWindow, err := ash.GetARCAppWindowInfo(ctx, tconn, pkg)
	if err != nil {
		return errors.Wrap(err, "failed to get arc window info")
	}

	tsw, err := input.Touchscreen(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to open touchscreen device")
	}
	defer tsw.Close()

	stw, err := tsw.NewSingleTouchWriter()
	if err != nil {
		return errors.Wrap(err, "could not create TouchEventWriter")
	}
	defer stw.Close()

	// Calculate Pixel (screen display) / Tuxel (touch device) ratio.
	dispW := dispMode.WidthInNativePixels
	dispH := dispMode.HeightInNativePixels
	pixelToTuxelX := float64(tsw.Width()) / float64(dispW)
	pixelToTuxelY := float64(tsw.Height()) / float64(dispH)

	captionHeight := int(math.Round(float64(appWindow.CaptionHeight) * dispMode.DeviceScaleFactor))
	bounds := ash.ConvertBoundsFromDpToPx(appWindow.BoundsInRoot, dispMode.DeviceScaleFactor)
	testing.ContextLogf(ctx, "The original window bound is %v, try to maximize it by drag inner hit-boxes", bounds)

	// Waiting for hit-boxes UI ready.
	if err := d.WaitForIdle(ctx, 10*time.Second); err != nil {
		return errors.Wrap(err, "failed to wait for idle")
	}

	innerMargin := 5
	middleX := bounds.Left + bounds.Width/2
	middleY := bounds.Top + bounds.Height/2
	for _, test := range []struct {
		startX, startY, endX, endY int
	}{
		{startX: bounds.Left + innerMargin, startY: middleY, endX: 0, endY: middleY},                        //left
		{startX: bounds.Left + bounds.Width - innerMargin, startY: middleY, endX: dispW - 1, endY: middleY}, //right
		{startX: middleX, startY: bounds.Top + innerMargin + captionHeight, endX: middleX, endY: 0},         //top
		{startX: middleX, startY: bounds.Top + bounds.Height - innerMargin, endX: middleX, endY: dispH - 1}, //bottom
	} {
		// Wait for application's UI ready.
		x0 := input.TouchCoord(float64(test.startX) * pixelToTuxelX)
		y0 := input.TouchCoord(float64(test.startY) * pixelToTuxelY)

		x1 := input.TouchCoord(float64(test.endX) * pixelToTuxelX)
		y1 := input.TouchCoord(float64(test.endY) * pixelToTuxelY)

		testing.ContextLogf(ctx, "Running the swipe gesture from {%d,%d} to {%d,%d} to ensure to start drag move", x0, y0, x1, y1)
		if err := stw.Swipe(ctx, x0, y0, x1, y1, 2*time.Second); err != nil {
			return errors.Wrap(err, "failed to execute a swipe gesture")
		}
		if err := stw.End(); err != nil {
			return errors.Wrap(err, "failed to finish the swipe gesture")
		}
		// Resize by companion library will take long time waiting for application's UI ready.
		if _, err := d.WaitForWindowUpdate(ctx, pkg, 10*time.Second); err != nil {
			return errors.Wrap(err, "failed to wait window updated after swipe resize")
		}
	}
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		appWindow, err = ash.GetARCAppWindowInfo(ctx, tconn, pkg)
		if err != nil {
			return testing.PollBreak(errors.Wrap(err, "failed to get arc window info"))
		}
		if appWindow.BoundsInRoot != ash.Rect(*dispInfo.WorkArea) {
			return errors.Errorf("resize window doesn't have the expected bounds yet; got %v, want %v", appWindow.BoundsInRoot, dispInfo.WorkArea)
		}
		return nil
	}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
		return err
	}
	return nil
}

// testWorkspaceInsets verifies that the workspace insets info from ChromeOS companion library is correct.
func testWorkspaceInsets(ctx context.Context, tconn *chrome.Conn, act *arc.Activity, d *ui.Device) error {
	const getWorkspaceInsetsButtonID = pkg + ":id/get_workspace_insets"

	parseRectString := func(rectShortString string, mode *display.DisplayMode) (ash.Rect, error) {
		// The rectangle short string generated by android /frameworks/base/graphics/java/android/graphics/Rect.java
		// Parse it to rectangle format with native pixel size.
		var left, top, right, bottom int
		if n, err := fmt.Sscanf(rectShortString, "[%d,%d][%d,%d]", &left, &top, &right, &bottom); err != nil {
			return ash.Rect{}, errors.Wrap(err, "Error on parse Rect text")
		} else if n != 4 {
			return ash.Rect{}, errors.Errorf("The format of Rect text is not valid: %q", rectShortString)
		}
		return ash.Rect{
			Left:   left,
			Top:    top,
			Width:  mode.WidthInNativePixels - left - right,
			Height: mode.HeightInNativePixels - top - bottom,
		}, nil
	}

	// Workspace insets infomation computed by window shelf info need several numeric conversion, which easy cause floating errors.
	const epsilon = 2
	isSimilarRect := func(lhs ash.Rect, rhs ash.Rect) bool {
		Abs := func(num int) int {
			if num >= 0 {
				return num
			}
			return -num
		}
		return Abs(lhs.Left-rhs.Left) <= epsilon && Abs(lhs.Width-rhs.Width) <= epsilon && Abs(lhs.Top-rhs.Top) <= epsilon && Abs(lhs.Height-rhs.Height) <= epsilon
	}

	dispMode, err := ash.InternalDisplayMode(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to get display mode")
	}
	dispInfo, err := display.GetInternalInfo(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "failed to get internal display info")
	}

	for _, test := range []struct {
		shelfAlignment ash.ShelfAlignment
		shelfBehavior  ash.ShelfBehavior
	}{
		{ash.ShelfAlignmentLeft, ash.ShelfBehaviorAlwaysAutoHide},
		{ash.ShelfAlignmentLeft, ash.ShelfBehaviorNeverAutoHide},
		{ash.ShelfAlignmentRight, ash.ShelfBehaviorAlwaysAutoHide},
		{ash.ShelfAlignmentRight, ash.ShelfBehaviorNeverAutoHide},
		{ash.ShelfAlignmentBottom, ash.ShelfBehaviorAlwaysAutoHide},
		{ash.ShelfAlignmentBottom, ash.ShelfBehaviorNeverAutoHide},
	} {
		if err := ash.SetShelfBehavior(ctx, tconn, dispInfo.ID, test.shelfBehavior); err != nil {
			return errors.Wrapf(err, "failed to set shelf behavior to %v", test.shelfBehavior)
		}
		if err := ash.SetShelfAlignment(ctx, tconn, dispInfo.ID, test.shelfAlignment); err != nil {
			return errors.Wrapf(err, "failed to set shelf alignment to %v", test.shelfAlignment)
		}
		var expectedShelfRect arc.Rect
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			// Confirm the shelf attribute has changed.
			if actualShelfAlignment, err := ash.GetShelfAlignment(ctx, tconn, dispInfo.ID); err != nil {
				return errors.Wrap(err, "failed to get shelf alignment")
			} else if actualShelfAlignment != test.shelfAlignment {
				return errors.Errorf("shelf alignment has not changed yet: got %v, want %v", actualShelfAlignment, test.shelfAlignment)
			}
			dispInfo, err := display.GetInternalInfo(ctx, tconn)
			if err != nil {
				return errors.Wrap(err, "failed to get internal display info")
			}
			// The unit of WorkArea is DP.
			expectedShelfRect = arc.Rect{
				Left:   dispInfo.WorkArea.Left,
				Top:    dispInfo.WorkArea.Top,
				Width:  dispInfo.WorkArea.Width,
				Height: dispInfo.WorkArea.Height,
			}
			return nil
		}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
			return errors.Wrap(err, "could not change the system shelf alignment")
		}

		// Read JSON format window insets size from CompanionLib Demo.
		baseMessage, err := getLastJSONMessage(ctx, d)
		if err != nil {
			return errors.Wrap(err, "failed to get basement json message")
		}
		if err := d.Object(ui.ID(getWorkspaceInsetsButtonID)).Click(ctx); err != nil {
			return errors.Wrap(err, "failed to click Get Workspace Insets button")
		}
		var msg *companionLibMessage
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			var err error
			msg, err = getLastJSONMessage(ctx, d)
			if err != nil {
				return testing.PollBreak(err)
			}
			// Waiting for new message coming
			if baseMessage.MessageID == msg.MessageID {
				return errors.New("still waiting the new json message")
			}
			return nil
		}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
			return errors.Wrap(err, "failed to get new message of device mode")
		}
		if msg.WorkspaceInsetMsg == nil {
			return errors.Errorf("unexpected JSON message format: no WorkspaceInsetMsg; got %v", msg)
		}
		parsedShelfRect, err := parseRectString(msg.WorkspaceInsetMsg.InsetBound, dispMode)
		if err != nil {
			return errors.Wrap(err, "failed to parse message")
		}
		// Convert two rectangle to same unit.
		expectedShelfRectPX := ash.ConvertBoundsFromDpToPx(ash.Rect(expectedShelfRect), dispMode.DeviceScaleFactor)

		if !isSimilarRect(expectedShelfRectPX, parsedShelfRect) {
			return errors.Errorf("Workspace Inset is not expected: got %v, want %v", parsedShelfRect, expectedShelfRectPX)
		}
	}
	return nil
}

// testCaptionButton verifies that hidden caption button API works as expected.
func testCaptionButton(ctx context.Context, tconn *chrome.Conn, act *arc.Activity, d *ui.Device) error {
	const (
		setCaptionButtonID                      = pkg + ":id/set_caption_buttons_visibility"
		checkCaptionButtonMinimizeBox           = pkg + ":id/caption_button_minimize"
		checkCaptionButtonMaximizeAndRestoreBox = pkg + ":id/caption_button_maximize_and_restore"
		checkCaptionButtonLegacyMenuBox         = pkg + ":id/caption_button_legacy_menu"
		checkCaptionButtonGoBackBox             = pkg + ":id/caption_button_go_back"
		checkCaptionButtonCloseBox              = pkg + ":id/caption_button_close"
	)

	resetCaptionCheckboxes := func() error {
		for _, checkboxID := range []string{
			checkCaptionButtonMinimizeBox,
			checkCaptionButtonMaximizeAndRestoreBox,
			checkCaptionButtonLegacyMenuBox,
			checkCaptionButtonGoBackBox,
			checkCaptionButtonCloseBox,
		} {
			checked, err := d.Object(ui.ID(checkboxID)).IsChecked(ctx)
			if err != nil {
				return errors.Wrap(err, "could not get the checkbox statement")
			}
			if checked != false {
				testing.ContextLogf(ctx, "Clean %s checkbox statements", checkboxID)
				if err := d.Object(ui.ID(checkboxID)).Click(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	}

	for _, test := range []struct {
		buttonCheckboxID        string
		buttonVisibleStatusMask ash.CaptionButtonStatus
	}{
		{checkCaptionButtonMinimizeBox, ash.CaptionButtonMinimize},
		{checkCaptionButtonMaximizeAndRestoreBox, ash.CaptionButtonMaximizeAndRestore},
		{checkCaptionButtonLegacyMenuBox, ash.CaptionButtonMenu},
		{checkCaptionButtonGoBackBox, ash.CaptionButtonBack},
		{checkCaptionButtonCloseBox, ash.CaptionButtonClose},
	} {
		testing.ContextLogf(ctx, "Test hiding %v caption button", test.buttonCheckboxID)
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			if err := d.Object(ui.ID(setCaptionButtonID)).Click(ctx); err != nil {
				return errors.Wrap(err, "could not click the setCaptionButton")
			}
			if err := resetCaptionCheckboxes(); err != nil {
				return errors.Wrap(err, "could not clean the button checkboxes setting")
			}
			if err := d.Object(ui.ID(test.buttonCheckboxID)).Click(ctx); err != nil {
				return errors.Wrap(err, "could not check the checkbox")
			}
			if err := d.Object(ui.Text("OK")).Click(ctx); err != nil {
				return errors.Wrap(err, "could not click the OK button")
			}
			return nil
		}, &testing.PollOptions{Timeout: 10 * time.Second}); err != nil {
			return errors.New("Error while changing hidden caption button")
		}

		window, err := ash.GetARCAppWindowInfo(ctx, tconn, pkg)
		if err != nil {
			return errors.Wrap(err, "error while get ARC window")
		}
		if window.CaptionButtonVisibleStatus&int(test.buttonVisibleStatusMask) != 0 {
			return errors.Errorf("Caption Button %v still visible", test.buttonCheckboxID)
		}
	}
	return nil
}

// testDeviceMode verifies that the device mode info from ChromeOS companion library is correct.
func testDeviceMode(ctx context.Context, tconn *chrome.Conn, act *arc.Activity, d *ui.Device) error {
	const getDeviceModeButtonID = pkg + ":id/get_device_mode_button"

	for _, test := range []struct {
		// isTabletMode represents current mode of system which is Tablet mode or clamshell mode.
		isTabletMode bool
		// modeStatus represents the expection of device mode string getting from companion library.
		modeStatus string
	}{
		{isTabletMode: true, modeStatus: "TABLET"},
		{isTabletMode: false, modeStatus: "CLAMSHELL"},
	} {
		// Force Chrome to be in specific system mode.
		if err := ash.SetTabletModeEnabled(ctx, tconn, test.isTabletMode); err != nil {
			return errors.Wrap(err, "failed to set the system mode")
		}

		// Read JSON format window caption height infomation.
		baseMessage, err := getLastJSONMessage(ctx, d)
		if err != nil {
			return errors.Wrap(err, "failed to get basement json message")
		}
		if err := d.Object(ui.ID(getDeviceModeButtonID)).Click(ctx); err != nil {
			return errors.Wrap(err, "could not click the getDeviceMode button")
		}
		var msg *companionLibMessage
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			var err error
			msg, err = getLastJSONMessage(ctx, d)
			if err != nil {
				return testing.PollBreak(err)
			}
			// Waiting for new message coming
			if baseMessage.MessageID == msg.MessageID {
				return errors.New("still waiting the new json message")
			}
			return nil
		}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
			return errors.Wrap(err, "failed to get new message of device mode")
		}
		if msg.DeviceModeMsg == nil {
			return errors.Errorf("unexpected JSON message format: no DeviceModeMsg; got %v", msg)
		}
		if msg.DeviceModeMsg.DeviceMode != test.modeStatus {
			return errors.Errorf("unexpected getDeviceMode result: got %s; want %s", msg.DeviceModeMsg.DeviceMode, test.modeStatus)
		}
	}
	return nil
}

// testWindowState verifies that change window state by ChromeOS companion library works as expected.
func testWindowState(ctx context.Context, tconn *chrome.Conn, act *arc.Activity, d *ui.Device) error {
	const (
		setWindowStateButtonID = pkg + ":id/set_task_window_state_button"
		getWindowStateButtonID = pkg + ":id/get_task_window_state_button"
	)
	// TODO(sstan): Add testcase of "Always on top" setting
	for _, test := range []struct {
		windowStateStr string
		windowStateExp arc.WindowState
		isAppManaged   bool
	}{
		{windowStateStr: "Minimize", windowStateExp: arc.WindowStateMinimized, isAppManaged: false},
		{windowStateStr: "Maximize", windowStateExp: arc.WindowStateMaximized, isAppManaged: false},
		{windowStateStr: "Normal", windowStateExp: arc.WindowStateNormal, isAppManaged: false},
	} {
		testing.ContextLogf(ctx, "Testing windowState=%v, appManaged=%t", test.windowStateStr, test.isAppManaged)
		if err := act.Start(ctx); err != nil {
			return errors.Wrap(err, "failed to start context")
		}
		if err := act.WaitForResumed(ctx, time.Second); err != nil {
			return errors.Wrap(err, "failed to wait for Resumed")
		}
		if err := d.Object(ui.ID(setWindowStateButtonID)).Click(ctx); err != nil {
			return errors.Wrap(err, "failed to click Set Task Window State button")
		}
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			if isClickable, err := d.Object(ui.Text(test.windowStateStr)).IsClickable(ctx); err != nil {
				return errors.Wrap(err, "failed check the radio clickable")
			} else if isClickable {
				// If isClickable = false, it will do nothing because the test application logic will automatically check the current window state radio. It can't be clicked if the state radio has been clicked.
				if err := d.Object(ui.Text(test.windowStateStr)).Click(ctx); err != nil {
					return errors.Wrapf(err, "failed to click %v", test.windowStateStr)
				}
			}
			return nil
		}, &testing.PollOptions{Timeout: 10 * time.Second}); err != nil {
			return errors.Wrap(err, "failed to waiting click radio")
		}

		if err := d.Object(ui.Text("OK")).Click(ctx); err != nil {
			return errors.Wrap(err, "failed to click OK button")
		}
		err := testing.Poll(ctx, func(ctx context.Context) error {
			actualWindowState, err := act.GetWindowState(ctx)
			if err != nil {
				return errors.Wrap(err, "could not get window state")
			}
			if actualWindowState != test.windowStateExp {
				return errors.Errorf("unexpected window state: got %v; want %v", actualWindowState, test.windowStateExp)
			}
			return nil
		}, &testing.PollOptions{Timeout: 10 * time.Second})
		if err != nil {
			return errors.Wrap(err, "error while waiting window state setting up")
		}
		if err := act.Stop(ctx); err != nil {
			return errors.Wrap(err, "failed to stop context")
		}
	}
	return nil
}

func setWindowStateSync(ctx context.Context, act *arc.Activity, state arc.WindowState) error {
	if err := act.SetWindowState(ctx, state); err != nil {
		return errors.Wrap(err, "could not set window state to normal")
	}
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		if currentState, err := act.GetWindowState(ctx); err != nil {
			return testing.PollBreak(errors.Wrap(err, "could not get the window state"))
		} else if currentState != state {
			return errors.Errorf("window state has not changed yet: got %s; want %s", currentState, state)
		}
		return nil
	}, &testing.PollOptions{Timeout: 4 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to waiting for change to normal window state")
	}
	return nil
}

// getTextViewContent returns all text in status textview.
func getTextViewContent(ctx context.Context, d *ui.Device) ([]string, error) {
	const statusTextViewID = pkg + ":id/status_text_view"
	text, err := d.Object(ui.ID(statusTextViewID)).GetText(ctx)
	if err != nil {
		// It not always success when get object, poll is necessary.
		return nil, errors.Wrap(err, "StatusTextView not ready yet")
	}
	return strings.Split(text, "\n"), nil
}

// getJSONTextViewContent returns all text in JSON textview.
func getJSONTextViewContent(ctx context.Context, d *ui.Device) ([]string, error) {
	const JSONTextViewID = pkg + ":id/status_jsontext_view"
	text, err := d.Object(ui.ID(JSONTextViewID)).GetText(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "JSONStatusTextView not ready yet")
	}
	return strings.Split(text, "\n"), nil
}

// getLastJSONMessage return last JSON format output message of ChromeOS Companion Library Demo
func getLastJSONMessage(ctx context.Context, d *ui.Device) (*companionLibMessage, error) {
	var lines []string
	if err := testing.Poll(ctx, func(ctx context.Context) error {
		var err error
		lines, err = getJSONTextViewContent(ctx, d)
		// Using poll here to avoid get text failure because UI compontent isn't stable.
		if err != nil {
			return errors.Wrap(err, "failed to get JSON message text")
		}
		return nil
	}, &testing.PollOptions{Timeout: 5 * time.Second}); err != nil {
		return nil, errors.Wrap(err, "failed to get a new line in status text view")
	}
	var msg companionLibMessage
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &msg); err != nil {
		return nil, errors.Wrap(err, "parse JSON format message failure")
	}
	return &msg, nil
}
