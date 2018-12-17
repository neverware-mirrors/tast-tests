// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package security

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"chromiumos/tast/local/upstart"
	"chromiumos/tast/testing"
)

func init() {
	testing.AddTest(&testing.Test{
		Func:         ASLR,
		Desc:         "Verifies that address space is randomized between runs",
		Attr:         []string{"informational"},
		SoftwareDeps: []string{"aslr"},
	})
}

func ASLR(ctx context.Context, s *testing.State) {
	type deviceNumber struct {
		major uint64
		minor uint64
	}

	type addressMapping struct {
		start  uintptr
		end    uintptr
		prot   int
		shared bool
		offset uint64
		device deviceNumber
		inode  uint64
		name   string
	}

	type addressMap []addressMapping

	parseNum := func(str string, base int) uint64 {
		parsed, err := strconv.ParseUint(str, base, 64)
		if err != nil {
			// Fataling here should be fine since we already do validation
			// when we match on the regex below.
			s.Fatalf("Failed to parse %v as base %v: %v", str, base, err)
		}
		return parsed
	}

	parseAddressMap := func(reader io.Reader) addressMap {
		mapping := `([0-9a-f]+)-([0-9a-f]+) +` + // start (1) and end (2)
			`(r|-)(w|-)(x|-)(s|p) +` + // protections (3-5) and sharing (6)
			`([0-9a-f]+) +` + // offset (7)
			`([0-9a-f]+):([0-9a-f]+) +` + // device major (8) and minor (9)
			`(\d+) *` + // inode number (10)
			`(.*)` // name (11)
		mappingMatcher := regexp.MustCompile(mapping)

		var am addressMap
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			group := mappingMatcher.FindStringSubmatch(line)
			if group == nil {
				s.Fatalf("Address map file line failed to parse: %q", line)
			}

			start := uintptr(parseNum(group[1], 16))
			end := uintptr(parseNum(group[2], 16))
			prot := 0
			if group[3][0] == 'r' {
				prot |= syscall.PROT_READ
			}
			if group[4][0] == 'w' {
				prot |= syscall.PROT_WRITE
			}
			if group[5][0] == 'x' {
				prot |= syscall.PROT_EXEC
			}
			shared := (group[6][0] == 's')
			offset := parseNum(group[7], 16)
			major := parseNum(group[8], 16)
			minor := parseNum(group[9], 16)
			inode := parseNum(group[10], 10)
			am = append(am, addressMapping{
				start, end, prot, shared, offset,
				deviceNumber{major, minor}, inode, group[11],
			})
		}

		if err := scanner.Err(); err != nil {
			s.Fatal("Failed to read map file: ", err)
		}

		return am
	}

	getNewJobMap := func(job string) addressMap {
		if err := upstart.RestartJob(ctx, job); err != nil {
			s.Fatalf("Job %v did not restart: %v", job, err)
		}
		_, _, pid, err := upstart.JobStatus(ctx, job)
		if err != nil {
			s.Fatalf("Could not get status for job %v: %v", job, err)
		}
		mapFile, err := os.Open(fmt.Sprintf("/proc/%v/maps", pid))
		if err != nil {
			s.Fatalf("Could not open address map for job %v: %v", job, err)
		}
		defer mapFile.Close()
		return parseAddressMap(mapFile)
	}

	// There will probably be multiple mappings for a lot of the files mapped into
	// memory. To deal with this, we only check the mappings with offset 0.
	getSectionStarts := func(am addressMap) map[string]uintptr {
		starts := make(map[string]uintptr)
		for _, mapping := range am {
			if (mapping.name != "[heap]" && mapping.name != "[stack]" &&
				mapping.inode == 0) || mapping.offset != 0 {
				// This isn't a mapped file or a private mapping we care about. Skip it.
				continue
			}

			starts[mapping.name] = mapping.start
		}
		return starts
	}

	compareStarts := func(m1 map[string]uintptr, m2 map[string]uintptr) {
		for name, start := range m1 {
			otherStart, present := m2[name]
			if present && start == otherStart {
				s.Errorf("Mapping for %v occurred at %#x in two maps", name, start)
			}
		}
	}

	const iterations = 5
	testRandomization := func(job string) {
		s.Log("Testing job ", job)
		originalMap := getNewJobMap(job)
		for i := 0; i < iterations; i++ {
			newMap := getNewJobMap(job)
			compareStarts(getSectionStarts(originalMap), getSectionStarts(newMap))
		}
	}

	for _, job := range []string{"ui", "debugd", "update-engine"} {
		testRandomization(job)
	}
}
