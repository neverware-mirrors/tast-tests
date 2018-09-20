// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ui

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Available RPC methods are listed at:
// https://github.com/xiaocong/android-uiautomator-server/blob/master/app/src/androidTest/java/com/github/uiautomator/stub/AutomatorService.java

var errTimeout = errors.New("timeout")

// Object is a representation of an Android view.
//
// An instantiated Object does NOT uniquely identify an Android view. Instead,
// it holds a selector to locate a matching view when its methods are called.
// Once you create an instance of Object, it can be reused for different views
// matching the selector.
//
// This object corresponds to UiObject in UI Automator API:
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject
type Object struct {
	d *Device
	s *selector
}

type objectInfo struct {
	Text               string `json:"text"`
	ContentDescription string `json:"contentDescription"`
	PackageName        string `json:"packageName"`
	ClassName          string `json:"className"`
	Checkable          bool   `json:"checkable"`
	Checked            bool   `json:"checked"`
	Clickable          bool   `json:"clickable"`
	Enabled            bool   `json:"enabled"`
	Focusable          bool   `json:"focusable"`
	Focused            bool   `json:"focused"`
	LongClickable      bool   `json:"longClickable"`
	Scrollable         bool   `json:"scrollable"`
	Selected           bool   `json:"selected"`
}

// Object creates an Object from given selectors.
//
// Example:
//  btn := d.Object(ui.ID("foo_button"), ui.Text("bar"))
func (d *Device) Object(opts ...SelectorOption) *Object {
	return &Object{d: d, s: newSelector(opts)}
}

// WaitForExists waits for a view matching the selector to appear.
//
// This method corresponds to UiObject.waitForExists().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#waitForExists(long)
func (o *Object) WaitForExists(ctx context.Context) error {
	return o.callSimple(ctx, "waitForExists", o.s, getTimeoutMs(ctx))
}

// Click clicks a view matching the selector.
//
// This method corresponds to UiObject.click().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#click
func (o *Object) Click(ctx context.Context) error {
	return o.callSimple(ctx, "click", o.s)
}

// SetText sets the text property of a view.
//
// This method corresponds to UiObject.setText().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#settext
func (o *Object) SetText(ctx context.Context, s string) error {
	return o.callSimple(ctx, "setText", o.s, s)
}

// GetText returns the text property of a view.
//
// This method corresponds to UiObject.getText().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#gettext
func (o *Object) GetText(ctx context.Context) (string, error) {
	info, err := o.info(ctx)
	if err != nil {
		return "", fmt.Errorf("GetText failed: %v", err)
	}
	return info.Text, nil
}

// GetContentDescription returns the content description property of a view.
//
// This method corresponds to UiObject.getContentDescription().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#getcontentdescription
func (o *Object) GetContentDescription(ctx context.Context) (string, error) {
	info, err := o.info(ctx)
	if err != nil {
		return "", fmt.Errorf("GetContentDescription failed: %v", err)
	}
	return info.ContentDescription, nil
}

// GetPackageName returns the package name of a view.
//
// This method corresponds to UiObject.getPackageName().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#getpackagename
func (o *Object) GetPackageName(ctx context.Context) (string, error) {
	info, err := o.info(ctx)
	if err != nil {
		return "", fmt.Errorf("GetPackageName failed: %v", err)
	}
	return info.PackageName, nil
}

// GetClassName returns the class name of a view.
//
// This method corresponds to UiObject.getClassName().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#getclassname
func (o *Object) GetClassName(ctx context.Context) (string, error) {
	info, err := o.info(ctx)
	if err != nil {
		return "", fmt.Errorf("GetClassName failed: %v", err)
	}
	return info.ClassName, nil
}

// IsCheckable returns if a view is checkable.
//
// This method corresponds to UiObject.isCheckable().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#ischeckable
func (o *Object) IsCheckable(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsCheckable failed: %v", err)
	}
	return info.Checkable, nil
}

// IsChecked returns if a view is checked.
//
// This method corresponds to UiObject.isChecked().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#ischecked
func (o *Object) IsChecked(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsChecked failed: %v", err)
	}
	return info.Checked, nil
}

// IsClickable returns if a view is clickable.
//
// This method corresponds to UiObject.isClickable().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#isclickable
func (o *Object) IsClickable(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsClickable failed: %v", err)
	}
	return info.Clickable, nil
}

// IsEnabled returns if a view is enabled.
//
// This method corresponds to UiObject.isEnabled().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#isenabled
func (o *Object) IsEnabled(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsEnabled failed: %v", err)
	}
	return info.Enabled, nil
}

// IsFocusable returns if a view is focusable.
//
// This method corresponds to UiObject.isFocusable().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#isfocusable
func (o *Object) IsFocusable(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsFocusable failed: %v", err)
	}
	return info.Focusable, nil
}

// IsFocused returns if a view is focused.
//
// This method corresponds to UiObject.isFocused().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#isfocused
func (o *Object) IsFocused(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsFocused failed: %v", err)
	}
	return info.Focused, nil
}

// IsLongClickable returns if a view is longClickable.
//
// This method corresponds to UiObject.isLongClickable().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#islongclickable
func (o *Object) IsLongClickable(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsLongClickable failed: %v", err)
	}
	return info.LongClickable, nil
}

// IsScrollable returns if a view is scrollable.
//
// This method corresponds to UiObject.isScrollable().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#isscrollable
func (o *Object) IsScrollable(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsScrollable failed: %v", err)
	}
	return info.Scrollable, nil
}

// IsSelected returns if a view is selected.
//
// This method corresponds to UiObject.isSelected().
// https://developer.android.com/reference/android/support/test/uiautomator/UiObject.html#isselected
func (o *Object) IsSelected(ctx context.Context) (bool, error) {
	info, err := o.info(ctx)
	if err != nil {
		return false, fmt.Errorf("IsSelected failed: %v", err)
	}
	return info.Selected, nil
}

// callSimple is a common method to call a RPC method that returns a boolean indicating success.
func (o *Object) callSimple(ctx context.Context, method string, params ...interface{}) error {
	var success bool
	if err := o.d.call(ctx, method, &success, params...); err != nil {
		return wrapMethodError(method, o.s, err)
	}
	if !success {
		return wrapMethodError(method, o.s, errTimeout)
	}
	return nil
}

// info returns an objectInfo of a view matched by the selector.
func (o *Object) info(ctx context.Context) (*objectInfo, error) {
	const method = "objInfo"
	var info objectInfo
	if err := o.d.call(ctx, method, &info, o.s); err != nil {
		return nil, wrapMethodError(method, o.s, err)
	}
	return &info, nil
}

// wrapMethodError wraps an error returned from an RPC method.
func wrapMethodError(method string, s *selector, err error) error {
	return fmt.Errorf("%s (selector=%v) failed: %v", method, s, err)
}

// getTimeoutMs returns the remaining timeout of ctx in milliseconds.
func getTimeoutMs(ctx context.Context) int64 {
	d, ok := ctx.Deadline()
	if !ok {
		return 1000000000000 // long enough timeout
	}
	t := d.Sub(time.Now()).Seconds()
	if t < 0 {
		return 1
	}
	return int64(t * 1000)
}
