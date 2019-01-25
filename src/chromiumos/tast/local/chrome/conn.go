// Copyright 2017 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chrome

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome/jslog"
	"chromiumos/tast/testing"
)

// Conn represents a connection to a web content view, e.g. a tab.
type Conn struct {
	co *rpcc.Conn
	cl *cdp.Client
	lw *jslog.Worker

	chromeErr func(error) error // wraps Chrome.chromeErr
}

func newConn(ctx context.Context, t *devtool.Target, lm *jslog.Master, chromeErr func(error) error) (*Conn, error) {
	testing.ContextLog(ctx, "Connecting to Chrome at ", t.WebSocketDebuggerURL)
	co, err := rpcc.DialContext(ctx, t.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if co != nil {
			co.Close()
		}
	}()

	cl := cdp.NewClient(co)
	if err := cl.Runtime.Enable(ctx); err != nil {
		return nil, err
	}

	if err = cl.Page.Enable(ctx); err != nil {
		return nil, err
	}

	ev, err := cl.Runtime.ConsoleAPICalled(ctx)
	if err != nil {
		return nil, err
	}

	lw := lm.NewWorker(t.ID, t.URL, ev)
	c := &Conn{co, cl, lw, chromeErr}
	co = nil
	return c, nil
}

// Close frees the connection's resources.
func (c *Conn) Close() {
	c.lw.Close()
	c.co.Close()
}

// Exec executes the JavaScript expression expr and discards its result.
// An error is returned if an exception is generated.
func (c *Conn) Exec(ctx context.Context, expr string) error {
	return c.doEval(ctx, expr, false, nil)
}

// Eval evaluates the JavaScript expression expr and stores its result in out.
// An error is returned if the result can't be unmarshaled into out.
//
//	sum := 0
//	err := conn.Eval(ctx, "3 + 4", &sum)
func (c *Conn) Eval(ctx context.Context, expr string, out interface{}) error {
	return c.doEval(ctx, expr, false, out)
}

// EvalPromise evaluates the JavaScript expression expr (which must return a Promise),
// awaits its result, and stores the result in out (if non-nil). An error is returned if
// evaluation fails, an exception is raised, ctx's deadline is reached, or out is non-nil
// and the result can't be unmarshaled into it.
//
//	data := make(map[string]interface{})
//	err := conn.EvalPromise(ctx,
//		`new Promise(function(resolve, reject) {
//			runAsync(function(data) {
//				if (data != null) {
//					resolve(data);
//				} else {
//					reject("it failed");
//				}
//			});
//		})`, &data)
func (c *Conn) EvalPromise(ctx context.Context, expr string, out interface{}) error {
	return c.doEval(ctx, expr, true, out)
}

// doEval is a helper function that evaluates JavaScript code for Exec, Eval, and EvalPromise.
func (c *Conn) doEval(ctx context.Context, expr string, awaitPromise bool, out interface{}) error {
	args := runtime.NewEvaluateArgs(expr)
	if awaitPromise {
		args = args.SetAwaitPromise(true)
	}
	if out != nil {
		args = args.SetReturnByValue(true)
	}

	repl, err := c.cl.Runtime.Evaluate(ctx, args)
	if err != nil {
		return err
	}
	exc := repl.ExceptionDetails
	if exc != nil {
		text := getExceptionText(repl.ExceptionDetails)
		c.lw.Report(time.Now(), "eval-error", text, exc.StackTrace)
		return errors.Errorf("got exception: %s", text)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(repl.Result.Value, out)
}

// getExceptionText extracts an error string from the exception described by d.
//
// The Chrome DevTools Protocol reports exceptions (and failed promises) in different ways depending
// on how they occur. This function tries to return a single-line string that contains the original error.
//
// Exec, Eval: throw new Error("foo"):
//	.Text:                  "Uncaught"
//	.Error:                 "runtime.ExceptionDetails: Uncaught exception at 0:0: Error: foo\n  <stack>"
//	.Exception.Description: "Error: foo\n  <stack>"
//	.Exception.Value:       null
//
// EvalPromise: reject("foo"):
//	.Text:                  "Uncaught (in promise)"
//	.Error:                 "runtime.ExceptionDetails: Uncaught (in promise) exception at 0:0"
//	.Exception.Description: nil
//	.Exception.Value:       "foo"
//
// EvalPromise: reject(new Error("foo")), throw new Error("foo"):
//	.Text:                  "Uncaught (in promise) Error: foo"
//	.Error:                 "runtime.ExceptionDetails: Uncaught (in promise) Error: foo exception at 0:0"
//	.Exception.Description: nil
//	.Exception.Value:       {}
func getExceptionText(d *runtime.ExceptionDetails) string {
	if d.Exception != nil && d.Exception.Description != nil {
		return strings.Split(*d.Exception.Description, "\n")[0]
	}
	var s string
	if err := json.Unmarshal(d.Exception.Value, &s); err == nil {
		return d.Text + ": " + s
	}
	return d.Text
}

// WaitForExpr repeatedly evaluates the JavaScript expression expr until it evaluates to true.
// Errors returned by Eval are treated the same as expr == false.
func (c *Conn) WaitForExpr(ctx context.Context, expr string) error {
	return c.waitForExprImpl(ctx, expr, false)
}

// WaitForExprFailOnErr repeatedly evaluates the JavaScript expression expr until it evaluates to true.
// It returns immediately if Eval returns an error.
func (c *Conn) WaitForExprFailOnErr(ctx context.Context, expr string) error {
	return c.waitForExprImpl(ctx, expr, true)
}

// waitForExprImpl repeatedly evaluates the JavaScript expression expr until it evaluates to true.
// The behavior on evaluation errors depends on the value of exitOnError.
func (c *Conn) waitForExprImpl(ctx context.Context, expr string, exitOnError bool) error {
	boolExpr := "!!(" + expr + ")"
	falseErr := errors.Errorf("%q is false", boolExpr)
	var lastErr error
	err := testing.Poll(ctx, func(ctx context.Context) error {
		v := false
		if err := c.Eval(ctx, boolExpr, &v); err != nil {
			if exitOnError {
				lastErr = err
				return nil
			}
			return err
		} else if !v {
			return falseErr
		}
		return nil
	}, &testing.PollOptions{Interval: 10 * time.Millisecond})
	if lastErr != nil {
		return lastErr
	}
	if err != nil {
		return c.chromeErr(err)
	}
	return nil
}

// PageContent returns the current top-level page content.
func (c *Conn) PageContent(ctx context.Context) (string, error) {
	doc, err := c.cl.DOM.GetDocument(ctx, nil)
	if err != nil {
		return "", err
	}
	result, err := c.cl.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		return "", err
	}
	return result.OuterHTML, nil
}

// Navigate navigates to url.
func (c *Conn) Navigate(ctx context.Context, url string) error {
	testing.ContextLog(ctx, "Navigating to ", url)
	fired, err := c.cl.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer fired.Close()

	if _, err := c.cl.Page.Navigate(ctx, page.NewNavigateArgs(url)); err != nil {
		return err
	}
	if _, err = fired.Recv(); err != nil {
		return err
	}
	return nil
}
