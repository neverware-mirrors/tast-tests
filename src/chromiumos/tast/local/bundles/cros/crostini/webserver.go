// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"chromiumos/tast/errors"
	"chromiumos/tast/local/crostini"
	"chromiumos/tast/local/testexec"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         Webserver,
		Desc:         "Runs a webserver in the container, and confirms that the host can connect to it",
		Contacts:     []string{"smbarber@chromium.org", "cros-containers-dev@google.com"},
		Attr:         []string{"group:mainline"},
		Vars:         []string{"keepState"},
		Timeout:      7 * time.Minute,
		Data:         []string{crostini.ImageArtifact},
		Pre:          crostini.StartedByArtifact(),
		SoftwareDeps: []string{"chrome", "vm_host"},
		Params: []testing.Param{
			{
				Name:              "artifact",
				ExtraHardwareDeps: crostini.CrostiniStable,
			},
			{
				Name:              "artifact_unstable",
				ExtraHardwareDeps: crostini.CrostiniUnstable,
				ExtraAttr:         []string{"informational"},
			},
		},
	})
}

func Webserver(ctx context.Context, s *testing.State) {
	pre := s.PreValue().(crostini.PreData)
	cr := pre.Chrome
	cont := pre.Container
	defer crostini.RunCrostiniPostTest(ctx, cont)

	const expectedWebContent = "nothing but the web"

	cmd := cont.Command(ctx, "sh", "-c",
		fmt.Sprintf("echo '%s' > ~/index.html", expectedWebContent))
	if err := cmd.Run(testexec.DumpLogOnError); err != nil {
		s.Fatal("Failed to add test index.html: ", err)
	}

	conn, err := cr.NewConn(ctx, "")
	if err != nil {
		s.Fatal("Failed to create renderer: ", err)
	}
	defer conn.Close()
	defer conn.CloseTarget(ctx)

	sockaddrs := []net.TCPAddr{
		{IP: net.IPv4(127, 0, 0, 1), Port: 6789}, // An unprivileged port listening on localhost should be tunneled.
		{IP: net.IPv4zero, Port: 999},            // Privileged ports will be accessible via penguin.linux.test but not localhost.
		{IP: net.IPv4zero, Port: 8000},           // Common dev webserver port.
		{IP: net.IPv4zero, Port: 12345},          // Uncommon dev webserver port.
		{IP: net.IPv6unspecified, Port: 8001},    // IPv6 all interfaces.
		{IP: net.IPv6loopback, Port: 8002},       // IPv6 loopback.
	}

	// Assemble the list of URLs to check. For localhost tunneling, check both
	// IPv4 and IPv6 connectivity. If listening on the unspecified address, also
	// check penguin.linux.test.
	checkURLs := make(map[string]bool)
	for _, sockaddr := range sockaddrs {
		if sockaddr.IP.IsLoopback() || sockaddr.IP.IsUnspecified() {
			// Localhost tunneling only works on unprivileged ports.
			if sockaddr.Port > 1023 {
				checkURLs[fmt.Sprintf("http://127.0.0.1:%d", sockaddr.Port)] = true
				checkURLs[fmt.Sprintf("http://[::1]:%d", sockaddr.Port)] = true
			}
		}

		// penguin.linux.test only works for IPv4 sockets listening on all interfaces.
		if sockaddr.IP.IsUnspecified() && sockaddr.IP.To4() != nil {
			checkURLs[fmt.Sprintf("http://penguin.linux.test:%d", sockaddr.Port)] = true
		}
	}

	testing.ContextLog(ctx, "Waiting for webservers to start up")
	for _, sockaddr := range sockaddrs {
		cmd = cont.Command(ctx, "sudo", "busybox", "httpd", "-f", "-p", sockaddr.String())
		if err := cmd.Start(); err != nil {
			s.Fatalf("Failed to start webserver on %v:%d: %v", sockaddr.IP, sockaddr.Port, err)
		}
		defer cmd.Wait()
		defer cmd.Kill()
	}

	checkNavigation := func(url string) error {
		if err = conn.Navigate(ctx, url); err != nil {
			return errors.Wrapf(err, "navigating to %q failed", url)
		}
		var actual string
		if err = conn.Eval(ctx, "document.documentElement.innerText", &actual); err != nil {
			return errors.Wrapf(err, "failed to get page content for %s", url)
		}
		if !strings.HasPrefix(actual, expectedWebContent) {
			return errors.Errorf("unexpected page content for %s; got %q, want %q", url, actual, expectedWebContent)
		}

		return nil
	}

	testing.ContextLog(ctx, "Checking URLs for webservers")
	for url := range checkURLs {
		if err := testing.Poll(ctx, func(ctx context.Context) error {
			return checkNavigation(url)
		}, &testing.PollOptions{Timeout: 10 * time.Second}); err != nil {
			s.Fatal("error polling for webserver: ", err)
		}
	}
}
