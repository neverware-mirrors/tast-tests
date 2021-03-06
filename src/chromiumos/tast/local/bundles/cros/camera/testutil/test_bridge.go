// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package testutil provides utilities to setup testing environment for camera
// tests.
package testutil

import (
	"context"
	"strings"

	"github.com/mafredri/cdp/protocol/target"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/testing"
)

// TestBridge is used to comminicate with CCA for test specific logic, such as test environment set-up/tear-down flow, performance/error monitoring.
type TestBridge struct {
	cr       *chrome.Chrome
	pageConn *chrome.Conn
	bridge   *chrome.JSObject
	useSWA   bool
}

// NewTestBridge returns a new test bridge instance.
func NewTestBridge(ctx context.Context, cr *chrome.Chrome, useSWA bool) (*TestBridge, error) {
	pageConn, bridge, err := setUpTestBridge(ctx, cr, useSWA)
	if err != nil {
		return nil, err
	}
	return &TestBridge{cr, pageConn, bridge, useSWA}, nil
}

func getPageConnForSWA(ctx context.Context, cr *chrome.Chrome) (*chrome.Conn, error) {
	conn, err := cr.NewConn(ctx, "chrome://camera-app/test/test.html")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build connection")
	}

	shouldCloseConn := true
	defer func() {
		if shouldCloseConn {
			if err := conn.Close(); err != nil {
				testing.ContextLog(ctx, "Failed to close connection: ", conn)
			}
		}
	}()

	// TODO(b/173092399): Remove the fallback for legacy path when Chrome is uprev.
	if pageContent, err := conn.PageContent(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to get page content")
	} else if strings.Contains(pageContent, "This site can’t be reached") {
		// Fallback to use legacy path for test page.
		if err := conn.Navigate(ctx, "chrome://camera-app/views/test.html"); err != nil {
			return nil, errors.Wrap(err, "failed to navigate to legacy test page")
		}
	}

	shouldCloseConn = false
	return conn, nil
}

func setUpTestBridge(ctx context.Context, cr *chrome.Chrome, useSWA bool) (*chrome.Conn, *chrome.JSObject, error) {
	var pageConn *chrome.Conn
	var err error
	if useSWA {
		pageConn, err = getPageConnForSWA(ctx, cr)
	} else {
		tconn, err := cr.TestAPIConn(ctx)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to connect to test extension")
		}
		// Background page will not be available if it is not used for a while.
		// We should wake it up before trying to connect it.
		if err := tconn.Call(ctx, nil, `
		  (id) => {
			chrome.runtime.sendMessage(id, "");
		  }`, ID); err != nil {
			return nil, nil, errors.Wrap(err, "failed to wake background page up")
		}
		pageConn, err = cr.NewConnForTarget(ctx, chrome.MatchTargetURL(BackgroundURL))
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to construct bridge page connection")
	}

	if err := pageConn.Eval(ctx, "location.reload()", nil); err != nil {
		return nil, nil, errors.Wrap(err, "failed to reload the test extension")
	}
	if err := pageConn.WaitForExpr(ctx, "document.readyState === 'complete'"); err != nil {
		return nil, nil, errors.Wrap(err, "failed to wait for the ready state")
	}

	const code = `
	  async function() {
		const workerPath = '/js/test_bridge.js';
		const sharedWorker = new SharedWorker(workerPath, {type: 'module'});
		const Comlink = await import('/js/lib/comlink.js');
		return Comlink.wrap(sharedWorker.port);
	  }`
	var testBridge chrome.JSObject
	if err := pageConn.Call(ctx, &testBridge, code); err != nil {
		if err2 := tearDownBridgePageConnection(ctx, cr, pageConn, useSWA); err2 != nil {
			testing.ContextLog(ctx, "Failed to tear down bridge page connection", err2)
		}
		return nil, nil, errors.Wrap(err, "failed to get test bridge")
	}
	return pageConn, &testBridge, nil
}

func tearDownBridgePageConnection(ctx context.Context, cr *chrome.Chrome, conn *chrome.Conn, useSWA bool) error {
	// For platform app, it does not make sense to close background page.
	if useSWA {
		checkTestPage := func(t *target.Info) bool {
			// TODO(b/173092399): Remove the legacy path when Chrome is uprev.
			return t.URL == "chrome://camera-app/test/test.html" || t.URL == "chrome://camera-app/views/test.html"
		}
		if testPageAlive, err := cr.IsTargetAvailable(ctx, checkTestPage); err == nil {
			if testPageAlive {
				if err := conn.CloseTarget(ctx); err != nil {
					return errors.Wrap(err, "failed to call CloseTarget() on the bridge page connection")
				}
			}
		} else {
			testing.ContextLog(ctx, "Failed to check if test page is alive or not: ", err)
		}
	}
	if err := conn.Close(); err != nil {
		return errors.Wrap(err, "failed to call Close() on the bridge page connection")
	}
	return nil
}

// AppWindow registers and returns the app window which is used to communicate with the foreground window of CCA instance.
func (t *TestBridge) AppWindow(ctx context.Context) (*AppWindow, error) {
	var appWindow chrome.JSObject
	if err := t.bridge.Call(ctx, &appWindow, "function() { return this.registerUnboundWindow(); }"); err != nil {
		return nil, errors.Wrap(err, "failed to register app window")
	}
	return &AppWindow{&appWindow}, nil
}

// Reset reconstructs the connection to test bridge.
func (t *TestBridge) Reset(ctx context.Context) error {
	if err := t.TearDown(ctx); err != nil {
		return err
	}

	pageConn, bridge, err := setUpTestBridge(ctx, t.cr, t.useSWA)
	if err != nil {
		return errors.Wrap(err, "failed to reconstruct test bridge")
	}
	t.pageConn = pageConn
	t.bridge = bridge
	return nil
}

// TearDown tears down the connection of test bridge.
func (t *TestBridge) TearDown(ctx context.Context) error {
	if t.bridge != nil {
		if err := t.bridge.Call(ctx, nil, `function() { this.close(); }`); err != nil {
			testing.ContextLog(ctx, "Failed to close worker: ", err)
		}
		if err := t.bridge.Release(ctx); err != nil {
			testing.ContextLog(ctx, "Failed to release test bridge object: ", err)
		}
		t.bridge = nil
	}
	if t.pageConn != nil {
		if err := tearDownBridgePageConnection(ctx, t.cr, t.pageConn, t.useSWA); err != nil {
			testing.ContextLog(ctx, "Failed to release bridge page connection: ", err)
		}
		t.pageConn = nil
	}
	return nil
}
