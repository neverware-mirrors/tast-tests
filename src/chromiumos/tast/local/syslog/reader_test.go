// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package syslog

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"chromiumos/tast/testutil"
)

const (
	fakeLine1 = "2019-12-10T11:17:28.123456+09:00 INFO foo[1234]: hello\n"
	fakeLine2 = "2019-12-10T11:17:29.123456+09:00 WARN bar[2345]: crashy\n"
	fakeLine3 = "2019-12-10T11:17:30.123456+09:00 INFO foo[1234]: world\n"
)

var (
	jst = time.FixedZone("JST", 9*60*60)

	fakeEntry1 = &Entry{
		Timestamp: time.Date(2019, 12, 10, 11, 17, 28, 123456000, jst),
		Severity:  "INFO",
		Tag:       "foo[1234]",
		Program:   "foo",
		PID:       1234,
		Content:   "hello",
	}
	fakeEntry2 = &Entry{
		Timestamp: time.Date(2019, 12, 10, 11, 17, 29, 123456000, jst),
		Severity:  "WARN",
		Tag:       "bar[2345]",
		Program:   "bar",
		PID:       2345,
		Content:   "crashy",
	}
	fakeEntry3 = &Entry{
		Timestamp: time.Date(2019, 12, 10, 11, 17, 30, 123456000, jst),
		Severity:  "INFO",
		Tag:       "foo[1234]",
		Program:   "foo",
		PID:       1234,
		Content:   "world",
	}
)

func TestReaderRead(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opts   []Option
		init   string
		writes []string
		want   []*Entry
	}{
		// Initial content handling tests:
		{
			name:   "InitEmpty",
			writes: []string{fakeLine1 + fakeLine2 + fakeLine3},
			want:   []*Entry{fakeEntry1, fakeEntry2, fakeEntry3},
		},
		{
			name:   "InitFullLine",
			init:   "this is the last log message\n",
			writes: []string{fakeLine1 + fakeLine2 + fakeLine3},
			want:   []*Entry{fakeEntry1, fakeEntry2, fakeEntry3},
		},
		{
			name:   "InitMiddleLine",
			init:   "this is ",
			writes: []string{"the last log message\n" + fakeLine1 + fakeLine2 + fakeLine3},
			want:   []*Entry{fakeEntry1, fakeEntry2, fakeEntry3},
		},
		// Write handling tests:
		{
			name:   "WriteAligned",
			writes: []string{fakeLine1, fakeLine2, fakeLine3},
			want:   []*Entry{fakeEntry1, fakeEntry2, fakeEntry3},
		},
		{
			name:   "WriteUnaligned",
			writes: []string{fakeLine1[:10], fakeLine1[10:] + fakeLine2 + fakeLine3[:20], fakeLine3[20:]},
			want:   []*Entry{fakeEntry1, fakeEntry2, fakeEntry3},
		},
		{
			name:   "WriteIncomplete",
			writes: []string{fakeLine1 + fakeLine2 + fakeLine3[:len(fakeLine3)-1]},
			want:   []*Entry{fakeEntry1, fakeEntry2},
		},
		// Parsing tests:
		{
			name:   "ParseNoPID",
			writes: []string{"2019-12-10T11:17:28.123456+09:00 INFO kernel: oops\n"},
			want: []*Entry{{
				Timestamp: time.Date(2019, 12, 10, 11, 17, 28, 123456000, jst),
				Severity:  "INFO",
				Tag:       "kernel",
				Program:   "kernel",
				Content:   "oops",
			}},
		},
		{
			name:   "ParseColonInProgram",
			writes: []string{"2019-12-10T11:17:28.123456+09:00 INFO foo:bar: hi\n"},
			want: []*Entry{{
				Timestamp: time.Date(2019, 12, 10, 11, 17, 28, 123456000, jst),
				Severity:  "INFO",
				Tag:       "foo:bar",
				Program:   "foo:bar",
				Content:   "hi",
			}},
		},
		// Option tests:
		{
			name:   "OptionProgram",
			opts:   []Option{Program("foo")},
			writes: []string{fakeLine1 + fakeLine2 + fakeLine3},
			want:   []*Entry{fakeEntry1, fakeEntry3},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tf, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatal("TempFile failed: ", err)
			}
			defer tf.Close()

			tf.WriteString(tc.init)

			opts := append([]Option{SourcePath(tf.Name())}, tc.opts...)
			r, err := NewReader(opts...)
			if err != nil {
				t.Fatal("NewReader failed: ", err)
			}
			defer r.Close()

			var got []*Entry

			readAll := func() {
				for {
					e, err := r.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal("Read failed: ", err)
					}
					got = append(got, e)
				}
			}

			readAll()
			for _, s := range tc.writes {
				tf.WriteString(s)
				readAll()
			}

			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Result unmatched (-got +want):\n%s", diff)
			}
		})
	}
}

func readAll(t *testing.T, r *Reader) []*Entry {
	t.Helper()
	var es []*Entry
	for {
		e, err := r.Read()
		if err == io.EOF {
			return es
		}
		if err != nil {
			t.Fatal("Read failed: ", err)
		}
		es = append(es, e)
	}
}

func TestReaderReadLogRotation(t *testing.T) {
	td := testutil.TempDir(t)
	defer os.RemoveAll(td)

	path := filepath.Join(td, "syslog")

	if err := ioutil.WriteFile(path, nil, 0644); err != nil {
		t.Fatal("Failed to create an empty syslog: ", err)
	}

	r, err := NewReader(SourcePath(path))
	if err != nil {
		t.Fatal("NewReader failed: ", err)
	}
	defer r.Close()

	if err := ioutil.WriteFile(path, []byte(fakeLine1), 0644); err != nil {
		t.Fatal("Failed to write the first entry: ", err)
	}
	if err := os.Rename(path, path+".rotated"); err != nil {
		t.Fatal("Rename failed: ", err)
	}
	if err := ioutil.WriteFile(path, []byte(fakeLine2), 0644); err != nil {
		t.Fatal("Failed to write the second entry: ", err)
	}

	got := readAll(t, r)
	want := []*Entry{fakeEntry1, fakeEntry2}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Result unmatched (-got +want):\n%s", diff)
	}
}

func TestReaderReadLogRotationRace(t *testing.T) {
	td := testutil.TempDir(t)
	defer os.RemoveAll(td)

	path := filepath.Join(td, "syslog")

	if err := ioutil.WriteFile(path, nil, 0644); err != nil {
		t.Fatal("Failed to create an empty syslog: ", err)
	}

	r, err := NewReader(SourcePath(path))
	if err != nil {
		t.Fatal("NewReader failed: ", err)
	}
	defer r.Close()

	if err := ioutil.WriteFile(path, []byte(fakeLine1), 0644); err != nil {
		t.Fatal("Failed to write the first entry: ", err)
	}
	if err := os.Rename(path, path+".rotated"); err != nil {
		t.Fatal("Rename failed: ", err)
	}

	// Simulate the race condition where a log is rotated but the new
	// file isn't created yet.
	got := readAll(t, r)
	want := []*Entry{fakeEntry1}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Result unmatched (-got +want):\n%s", diff)
	}

	if err := ioutil.WriteFile(path, []byte(fakeLine2), 0644); err != nil {
		t.Fatal("Failed to write the second entry: ", err)
	}

	got = readAll(t, r)
	want = []*Entry{fakeEntry2}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Result unmatched (-got +want):\n%s", diff)
	}
}
