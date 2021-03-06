// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package uig is an action graph library for UI automation on Chrome OS.
// Using uig you create a graph of UI actions to be executed,
// then call Do to execute the graph.
//
// Because the graph construction functions do not return multiple values(eg errors)
// you can freely compose them and handle errors/cleanup/return values only when the
// graph is actually executed. This results in much denser and easier to read UI automation code.
//
// Example:
//  statusArea := uig.FindWithTimeout(ui.FindParams{ClassName: "ash/StatusAreaWidgetDelegate"}, 10*time.Second)
//  steps := uig.Steps(
//     statusArea.LeftClick(),
//     uig.FindWithTimeout(ui.FindParams{Name: name, ClassName: "FeaturePodIconButton"}, 10*time.Second).LeftClick()
//     statusArea.LeftClick())
//  if err := uig.Do(ctx, tconn, steps); err != nil {
//     return errors.Wrap(err, "failed to toggle tray setting")
//  }
package uig

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/chrome"
	"chromiumos/tast/local/chrome/ui"
)

// Action is a node in a UI automation graph.  Call Do to execute the entire graph starting with the given action.
type Action struct {
	// Execute the graph starting from this node.
	//
	// Root is the root of the context in which the graph should be evaluated.  It should be passed through the graph
	// until it reaches reach uig.Root() where it will become the returned node.  Ownership of the root reference is
	// passed to the do function, however, in the common case the reference is simply passed on to the parent action.
	//
	// Execute returns the underlying ui automation node from the last performed node of the graph, which must be released by the caller.
	// If there is an error, the *nodeRef return will always be nil.
	// If there is no error, the *nodeRef will never be nil.  If you are implementing an action and have nothing sensible to return,
	// return the node from uig.Root().
	do func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error)

	// name is a string representation of the graph starting at this node, typically used in error messages.
	name string
}

// nodeRef is a reference counted wrapper around a ui.Node.
// This allows us to pass a parent nodeRef to multiple graphs in Steps and have it only
// be Release()ed when all of the graphs complete.
type nodeRef struct {
	refs int
	node *ui.Node
}

// newNodeRef creates a nodeRef from a *ui.Node.  The reference count is set to 1
// so a single release() call will release the underlying ui.Node.
func newNodeRef(node *ui.Node) *nodeRef {
	return &nodeRef{refs: 1, node: node}
}

// acquire creates a "copy" of the ui.Node pointer. Internally it simply increments the reference counter.
func (r *nodeRef) acquire() {
	r.refs++
}

// release releases a "copy" of the ui.Node pointer.  Once all existing "copies" have been released the underlying
// ui.Node will be released.  Internally it simply decrements the reference counter, and calls Release() when the
// counter reaches 0.
func (r *nodeRef) release(ctx context.Context) {
	r.refs--
	if r.refs == 0 {
		r.node.Release(ctx)
		r.node = nil
	}
}

// String returns a string representation of the graph starting at this node.
func (a *Action) String() string {
	return a.name
}

// WithNamef gives an action graph an alternative name.
//
// This works best if name resembles a function call, eg:
// func FindColorButton(color string) *Action {
//     return uig.Find(...).
//         Find(...).
//         WithNamef("FindColorButton(%q)", color)
// }
//
// Any errors that occur inside this action graph will be wrapped with the
// name, eg:
// file.go:27: got an error: FindColorButton("blue"): uig.Find(...).Find(...): couldn't find element.
//
// Any errors that occur in child actions of this action will have the parent
// actions replaced with the name, eg:
// file.go:27: got an error: FindColorButton("blue").LeftClick(): couldn't click.
func (a *Action) WithNamef(format string, params ...interface{}) *Action {
	name := fmt.Sprintf(format, params...)
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}

}

// Steps combines actions into a sequence of steps that are executed one after another.
// The context root of each step is the node resulting from the action on which steps is called.
func (a *Action) Steps(actions ...*Action) *Action {
	var steps []string
	for _, action := range actions {
		steps = append(steps, action.String())
	}
	name := fmt.Sprintf("%v.Steps(%s)", a, strings.Join(steps, ", "))

	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, errors.Wrap(err, name)
			}
			for i, action := range actions {
				node.acquire()
				child, err := action.do(ctx, tconn, node)
				if err != nil {
					if len(actions) > 1 {
						return nil, errors.Wrapf(err, "Step %d", i+1)
					}
					return nil, err
				}
				child.release(ctx)
			}
			return node, nil
		},
	}
}

// Steps is a shortcut for uig.Root().Steps(...)
func Steps(actions ...*Action) *Action {
	return Root().Steps(actions...)
}

// Retry retries a given action graph if it returns error.
//
// The graphs will be executed up to times times, including the first attempt.
//
// The last error will be returned.  Any other errors will be silently ignored.
func (a *Action) Retry(times int, action *Action) *Action {
	name := fmt.Sprintf("%v.Retry(%d, %s)", a, times, action)

	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, errors.Wrap(err, name)
			}
			var actionErr error
			for i := 0; i < times; i++ {
				node.acquire()
				var child *nodeRef
				child, actionErr = action.do(ctx, tconn, node)
				if actionErr == nil {
					node.release(ctx)
					return child, nil
				}
			}
			node.release(ctx)
			return nil, errors.Wrapf(actionErr, "action failed %d times, last error", times)
		},
	}
}

// Retry is a shortcut for uig.Root().Retry(...)
func Retry(times int, action *Action) *Action {
	return Root().Retry(times, action)
}

// LeftClick sends a left mouse click to the screen location of the given node.
//
// Note that if the node is not on the screen it cannot be clicked.  thus you
// may need to call Focus() first.
//
// Also note that if something else is on top of the node (eg a notification)
// that will be clicked instead.
func (a *Action) LeftClick() *Action {
	name := fmt.Sprintf("%s.LeftClick()", a.String())
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			if err := node.node.LeftClick(ctx); err != nil {
				node.release(ctx)
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}
}

// RightClick sends a right mouse click to the screen location of the given node.
//
// Note that if the node is not on the screen it cannot be clicked.  thus you
// may need to call Focus() first.
//
// Also note that if something else is on top of the node (eg a notification)
// that will be clicked instead.
func (a *Action) RightClick() *Action {
	name := fmt.Sprintf("%s.RightClick()", a.String())
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			if err := node.node.RightClick(ctx); err != nil {
				node.release(ctx)
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}
}

// FocusAndWait puts the keyboard focus on to a node.
//
// An important side effect of this is that it scrolls the node into view,
// which may be necessary before you can call LeftClick on it.
func (a *Action) FocusAndWait(timeout time.Duration) *Action {
	name := fmt.Sprintf("%s.Focus()", a.String())
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			if err := node.node.FocusAndWait(ctx, timeout); err != nil {
				node.release(ctx)
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}
}

// Find finds a descendant of the node it is called on.
func (a *Action) Find(params ui.FindParams) *Action {
	name := fmt.Sprintf("%s.Find(%+v)", a.String(), params)
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			child, err := node.node.Descendant(ctx, params)
			node.release(ctx)
			if err != nil {
				return nil, errors.Wrap(err, name)
			}
			return newNodeRef(child), nil
		},
	}
}

// Find is a shortcut for uig.Root().Find(...).
func Find(params ui.FindParams) *Action {
	return Root().Find(params).WithNamef("uig.Find(%+v)", params)
}

// FindWithTimeout finds a descendant of the node it is called on. It returns an error if the timeout expires.
func (a *Action) FindWithTimeout(params ui.FindParams, timeout time.Duration) *Action {
	name := fmt.Sprintf("%s.FindWithTimeout(%+v, %v)", a.String(), params, timeout)
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			child, err := node.node.DescendantWithTimeout(ctx, params, timeout)
			node.release(ctx)
			if err != nil {
				return nil, errors.Wrap(err, name)
			}
			return newNodeRef(child), nil
		},
	}
}

// FindWithTimeout is a shortcut for uig.Root().FindWithTimeout(...).
func FindWithTimeout(params ui.FindParams, timeout time.Duration) *Action {
	return Root().FindWithTimeout(params, timeout).WithNamef("uig.FindWithTimeout(%+v, %v)", params, timeout)
}

// WaitUntilDescendantExists waits until a given node is found as a descendant
// of the node this is called on.
//
// If the timeout expires while the node does not exist it will return error.
func (a *Action) WaitUntilDescendantExists(params ui.FindParams, timeout time.Duration) *Action {
	name := fmt.Sprintf("%s.WaitUntilDescendantExists(%+v, %v)", a.String(), params, timeout)
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			err = node.node.WaitUntilDescendantExists(ctx, params, timeout)
			if err != nil {
				node.release(ctx)
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}
}

// WaitUntilDescendantExists is a shortcut for uig.Root().WaitUntilDescendantExists(...)
func WaitUntilDescendantExists(params ui.FindParams, timeout time.Duration) *Action {
	return Root().WaitUntilDescendantExists(params, timeout).WithNamef("uig.WaitUntilDescendantExists(%+v, %v)", params, timeout)
}

// WaitUntilDescendantGone waits until a given node can no longer be found as a descendant
// of the node this is called on.
//
// If the timeout expires while the node still exists it will return error.
func (a *Action) WaitUntilDescendantGone(params ui.FindParams, timeout time.Duration) *Action {
	name := fmt.Sprintf("%s.WaitUntilDescendantGone(%+v, %v)", a.String(), params, timeout)
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			err = node.node.WaitUntilDescendantGone(ctx, params, timeout)
			if err != nil {
				node.release(ctx)
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}
}

// WaitUntilDescendantGone is a shortcut for uig.Root().WaitUntilDescendantGone(...)
func WaitUntilDescendantGone(params ui.FindParams, timeout time.Duration) *Action {
	return Root().WaitUntilDescendantGone(params, timeout).WithNamef("uig.WaitUntilDescendantGone(%+v, %v)", params, timeout)
}

// WaitForLocationChangeCompleted waits until there are no LocationChanged
// events on the node for at least 2 seconds. This is commonly used to wait
// for the completion of UI animations.
func (a *Action) WaitForLocationChangeCompleted() *Action {
	name := fmt.Sprintf("%s.WaitForLocationChangeCompleted()", a.String())
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			node, err := a.do(ctx, tconn, root)
			if err != nil {
				return nil, err
			}
			if err := ui.WaitForLocationChangeCompletedOnNode(ctx, tconn, node.node); err != nil {
				node.release(ctx)
				return nil, errors.Wrap(err, name)
			}
			return node, nil
		},
	}
}

// WaitForLocationChangeCompleted is a shortcut for uig.Root().WaitForLocationChangeCompleted()
func WaitForLocationChangeCompleted() *Action {
	return Root().WaitForLocationChangeCompleted().WithNamef("uig.WaitForLocationChangeCompleted()")
}

// Root gets the root node of the context this graph is being executed in.  This is typically the
// ChromeOS Desktop, although another context root can be specified by calling Steps on it.
//
// For example, in the following code uig.Root will return the desktop:
//
//     uig.Do(ctx, tconn, uig.Root().Find(...))
//
// However, in the following code uig.Root will return the node from the uig.Find(...):
//
//     uig.Do(ctx, tconn, uig.Find(...).Steps(uig.Root().Find(...)))
func Root() *Action {
	name := "uig.Root()"
	return &Action{
		name: name,
		do: func(ctx context.Context, tconn *chrome.TestConn, root *nodeRef) (*nodeRef, error) {
			return root, nil
		},
	}
}

// GetNode executes an action graph.  It returns the ui node from the last action.
// The caller is responsible for calling Release() on the returned *ui.Node.
// The graph is executed with the context root of the ChromeOS Desktop.
func GetNode(ctx context.Context, tconn *chrome.TestConn, graph *Action) (*ui.Node, error) {
	root, err := ui.Root(ctx, tconn)
	if err != nil {
		return nil, errors.Wrap(err, "could not get desktop ui.Node in uig.GetNode")
	}
	node, err := graph.do(ctx, tconn, newNodeRef(root))
	if err != nil {
		return nil, err
	}
	return node.node, nil
}

// Do executes one or more action graphs in sequence.  It automatically releases any resources as required.
// The graph is executed with the context root of the ChromeOS Desktop.
func Do(ctx context.Context, tconn *chrome.TestConn, graphs ...*Action) error {
	root, err := ui.Root(ctx, tconn)
	if err != nil {
		return errors.Wrap(err, "could not get desktop ui.Node in uig.Do")
	}
	node, err := Steps(graphs...).do(ctx, tconn, newNodeRef(root))
	if err != nil {
		return err
	}
	node.release(ctx)
	return nil
}

// PageObject creates an Action for each field of Action type from the tags in the structure.
// For example, a struct of a dialog page with msg, cancel and ok button could be defined as:
//
// 		type Dialog struct {
// 			Self   *uig.Action `name:"name" role:"dialog"`
// 			Msg    *uig.Action `name:"msg" role:"staticText"`
// 			OK     *uig.Action `name:"OK" role:"button"`
// 			Cancel *uig.Action `name:"Cancel" role:"button"`
// 		}
//
// Where Self is the dialog itself and the rest are the content on the page.
// To create a variable of the dialog page, do this:
//
// 		dialog := &Dialog{}
// 		uig.PageObject(dialog)
//
// After this, the variable dialog is like this:
//
// 		{
// 			Self   = uig.FindWithTimeout(ui.FindParams{Name: "name", Role: ui.RoleTypeDialog},15*time.Second)
// 			Msg    = uig.FindWithTimeout(ui.FindParams{Name: "msg", Role: ui.RoleTypeStaticText},15*time.Second)
// 			OK     = uig.FindWithTimeout(ui.FindParams{Name: "OK", Role: ui.RoleTypeButton},15*time.Second)
// 			Cancel = uig.FindWithTimeout(ui.FindParams{Name: "Cancel", Role: ui.RoleTypeButton},15*time.Second)
// 		}
//
// Then the following code will find/click the item on the dialog:
//		uig.Do(ctx, tconn, dialog.Self)
//		uig.Do(ctx, tconn, dialog.OK.LeftClick())
func PageObject(pg interface{}) {
	v := reflect.ValueOf(pg).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldStruct := v.Type().Field(i)
		fieldValue := v.FieldByName(fieldStruct.Name)
		if fieldValue.IsValid() && fieldValue.CanSet() && fieldStruct.Type == reflect.TypeOf(&Action{}) {
			params := ui.FindParams{Attributes: map[string]interface{}{}}
			hasParams := false

			// Find tag name.
			if name, ok := fieldStruct.Tag.Lookup("name"); ok && name != "" {
				params.Name = name
				hasParams = true
			} else if nameRegex, ok := fieldStruct.Tag.Lookup("nameRegex"); ok && name != "" {
				params.Attributes["name"] = regexp.MustCompile(nameRegex)
				hasParams = true
			}

			// Find tag role.
			if role, ok := fieldStruct.Tag.Lookup("role"); ok && role != "" {
				params.Role = ui.RoleType(role)
				hasParams = true
			}

			// Find tag className.
			if cName, ok := fieldStruct.Tag.Lookup("className"); ok && cName != "" {
				params.ClassName = cName
				hasParams = true
			}

			// TODO(jinrongwu): handle Attributes and State when necessary

			// Set the field value.
			if hasParams {
				fieldValue.Set(reflect.ValueOf(FindWithTimeout(params, 15*time.Second)))
			}
		}
	}
}
