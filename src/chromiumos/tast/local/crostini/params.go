// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package crostini

// This file contains the generators for crostini test
// parameters. Crostini tests generally require multiple sub-tests
// parameterized over how crostini is set up. As these parameters are
// complex and *change over time* you should strongly prefer to use
// these generators rather then copying and pasting from other tests
// or writing your own. This allows for changes to be made in one
// central location (this file) and automatically propagated to all
// tests that rely on them.
//
// Usage:
//
// In general, to generate test parameters for a test you add a unit
// test file to the package containing the test (the name must end in
// "_test.go"). This file should contain a go unit test which creates
// a string containing the test parameters and passes it to
// genparams.Ensure, along with the name of the test file. For
// crostini tests, generate this string by calling one of the
// functions in this file.
//
// By default, this unit test will ensure that the parameters in the
// test match up with the string generated by the unit tests. These
// tests are run on the CQ, so unexpected changes will block CLs from
// landing. When run with TAST_GENERATE_UPDATE=1 set, it will instead
// update the parameters in the test file.
//
// go unit tests can be run using the command
// ~/trunk/src/platform/tast/tools/go.sh test -count=1 chromiumos/tast/local/bundles/...
// parameter regeneration can be done with
// TAST_GENERATE_UPDATE=1 ~/trunk/src/platform/tast/tools/go.sh test -count=1 chromiumos/tast/local/bundles/...
//
// The crostini package has a mega-test in params_test.go which
// generates parameters for most of its tests. If you are adding a
// test to that package, consider if it can be added to that
// file. This should be the case for most tests, which only need the
// "normal" set of parameters. Otherwise, or if you are writing a test
// outside of the crostini package, you will likely need to create a
// new test file. See any of the file in the crostini package ending
// in "_test.go" for examples.
//
// Tests using these generators should not specify timeouts or
// preconditions, as we expect to be able to specify those freely in
// the parameters.
//
// Any test which is controlled by a generator will have a comment
// above its test parameters indicating which file contains the
// relevant unit test. To modify the parameters, update the test and
// run the above command to regenerate the results.

import (
	"time"

	"chromiumos/tast/common/genparams"
	"chromiumos/tast/local/vm"
)

// Param specifies how each set of crostini tests should be generated.
type Param struct {
	// Name of the test case. Generated tests will look like
	// "name_artifact", "name_download_buster" etc.
	Name string

	// ExtraAttr contains additional attributes to add to the
	// generated test's ExtraAttr field beyond what the generator
	// function adds. For example, if you want this test case to
	// be in the "graphics_daily" group without putting the whole
	// test in, you can add that label here.
	ExtraAttr []string

	// ExtraData contains paths of additional data files needed by
	// the test case. Note that data files required for specific
	// crostini preconditions are added automatically to the
	// generated tests and should not be added here.
	ExtraData []string

	// ExtraSoftwareDeps lists software features that are required
	// to run this test case.
	ExtraSoftwareDeps []string

	// Timeout indicates the timeout for this test case. This is
	// used directly for artifact tests, download tests add 3
	// minutes to allow additional time for the VM and container to
	// be downloaded. If unspecified, defaults to 7 * time.Minute.
	Timeout time.Duration

	// Val is a freeform value that can be retrieved from
	// testing.State.Param() method. This string is inserted
	// unmodified and unquoted into the generated test case code
	// as the Val for each test case generated for this object.
	Val string

	// Preconditions is a map from debian version to a string
	// containing a go expression that evaluates to the
	// precondition that should be used to install that
	// version. If not set, defaults to the
	// crostini.StartedByComponent{Stretch,Buster} preconditions.
	Preconditions map[vm.ContainerDebianVersion]string

	// StableHardwareDep contains a go expression that evaluates
	// to a hardware dependency which controls the collection of
	// boards considered stable.
	StableHardwareDep string

	// UnstableHardwareDep contains a go expression that evaluates
	// to a hardware dependency which controls the collection of
	// boards considered unstable. It should be the inverse of
	// StableHardwareDep.
	UnstableHardwareDep string

	// MinimalSet - if true, generate only a minimal set of test
	// parameters such that each device will have at most one test
	// case it can run. This is useful for things like performance
	// tests, which are too expensive to be run in every possible
	// configuration.
	MinimalSet bool

	// IsNotMainline indicates whether the test case is in
	// group:mainline or not. This is important to get right
	// because we can't add the "informational" attribute to
	// non-mainline tests, but leaving it off of a mainline test
	// will cause the test case to become CQ critical. If in
	// doubt, set to false, and if you're wrong you will get an
	// error message when you try to run your tests.
	//
	// This also controls whether separate stable/unstable
	// artifact tests are generated, since this distinction is
	// only relevant to the CQ.
	IsNotMainline bool

	// UseLargeContainer controls whether to use the normal test
	// container, or a larger container with more applications
	// pre-installed.
	UseLargeContainer bool

	// OnlyStable controls whether to only use the stable board
	// variants and exclude all the unstable variants.
	OnlyStableBoards bool
}

type generatedParam struct {
	Name              string
	ExtraAttr         []string
	ExtraData         []string
	ExtraSoftwareDeps []string
	ExtraHardwareDeps string
	Pre               string
	Timeout           time.Duration
	Val               string
}

const template = `{{range .}} {
	{{if .Name}}              Name:              {{fmt .Name}},                                             {{end}}
	{{if .ExtraAttr}}         ExtraAttr:         []string{ {{range .ExtraAttr}}{{fmt .}},{{end}} },         {{end}}
	{{if .ExtraData}}         ExtraData:         []string{ {{range .ExtraData}}{{fmt .}},{{end}} },         {{end}}
	{{if .ExtraSoftwareDeps}} ExtraSoftwareDeps: []string{ {{range .ExtraSoftwareDeps}}{{fmt .}},{{end}} }, {{end}}
	{{if .ExtraHardwareDeps}} ExtraHardwareDeps: {{.ExtraHardwareDeps}},                                    {{end}}
	{{if .Pre}}               Pre:               {{.Pre}},                                                  {{end}}
	{{if .Timeout}}           Timeout:           {{fmt .Timeout}},                                          {{end}}
	{{if .Val}}               Val:               {{.Val}},                                                  {{end}}
}, {{end}}`

// MakeTestParamsFromList takes a list of test cases (in the form of
// crostini.Param objects) and generates a set of crostini test
// parameters for each. Currently this means all four of artifact,
// artifact_unstable, download_stretch, and download_buster tests. See
// the documentation for crostini.Param for how these values effect
// the results. Each crostini.Param object is treated independently,
// producing its own set of sub-tests.
//
// Normally you should use MakeTestParams instead, but if your test is
// parameterized beyond which crostini preconditions it uses, you will
// need this.
func MakeTestParamsFromList(t genparams.TestingT, baseCases []Param) string {
	var result []generatedParam
	for _, testCase := range baseCases {
		var namePrefix string
		if testCase.Name != "" {
			namePrefix = testCase.Name + "_"
		}

		// Check here if it's possible for any iteration of
		// this test to be critical, i.e. if it doesn't
		// already have the "informational" attribute, and is
		// a mainline test.
		canBeCritical := true
		for _, attr := range testCase.ExtraAttr {
			if attr == "informational" {
				canBeCritical = false
			}
		}

		stableBoards := []bool{true}
		if !testCase.OnlyStableBoards {
			stableBoards = append(stableBoards, false)
		}

		for _, debianVersion := range []vm.ContainerDebianVersion{vm.DebianStretch, vm.DebianBuster} {
			if testCase.MinimalSet && debianVersion != vm.DebianBuster {
				continue
			}

			for _, arch := range []string{"amd64", "arm"} {

				for _, stable := range stableBoards {
					if !stable && testCase.IsNotMainline {
						// The stable/unstable distinction is only important for mainline tests
						continue
					}

					if !stable && testCase.UseLargeContainer {
						// When using the large container we have to restrict ourselves to
						// a smaller range of devices, and the unstable tests are not
						// expected to provide a useful signal.
						continue
					}

					name := ""
					if !testCase.MinimalSet {
						// If we're generating a minimal set then the debian version
						// is always the same and we don't need to include it in the test name.
						name += string(debianVersion) + "_"
					}
					name += arch
					if !testCase.IsNotMainline && !testCase.UseLargeContainer {
						if stable {
							name += "_stable"
						} else {
							name += "_unstable"
						}
					}

					// _unstable tests can never be CQ critical.
					var extraAttr []string
					if !stable && canBeCritical {
						extraAttr = append(extraAttr, "informational")
					}

					extraData := []string{
						getVMArtifact(arch),
						getContainerMetadataArtifact(arch, debianVersion, testCase.UseLargeContainer),
						getContainerRootfsArtifact(arch, debianVersion, testCase.UseLargeContainer),
					}

					extraSoftwareDeps := []string{arch}

					var hardwareDeps string
					if !testCase.IsNotMainline {
						if stable {
							if testCase.StableHardwareDep != "" {
								hardwareDeps = testCase.StableHardwareDep
							} else {
								hardwareDeps = "crostini.CrostiniStable"
							}
						} else {
							if testCase.UnstableHardwareDep != "" {
								hardwareDeps = testCase.UnstableHardwareDep
							} else {
								hardwareDeps = "crostini.CrostiniUnstable"
							}
						}
					}

					var precondition string
					if testCase.Preconditions != nil {
						precondition = testCase.Preconditions[debianVersion]
					} else if debianVersion == vm.DebianStretch {
						precondition = "crostini.StartedByComponentStretch()"
					} else {
						precondition = "crostini.StartedByComponentBuster()"
					}

					var timeout time.Duration
					if testCase.Timeout != time.Duration(0) {
						timeout = testCase.Timeout
					} else {
						timeout = 7 * time.Minute
					}

					testParam := generatedParam{
						Name:              namePrefix + name,
						ExtraAttr:         append(testCase.ExtraAttr, extraAttr...),
						ExtraData:         append(testCase.ExtraData, extraData...),
						ExtraSoftwareDeps: append(testCase.ExtraSoftwareDeps, extraSoftwareDeps...),
						ExtraHardwareDeps: hardwareDeps,
						Pre:               precondition,
						Timeout:           timeout,
						Val:               testCase.Val,
					}
					result = append(result, testParam)
				}
			}
		}
	}
	return genparams.Template(t, template, result)
}

// MakeTestParams generates the default set of crostini test
// parameters using MakeTestParamsFromList. If your test only needs to
// be parameterized over how crostini is acquired and which version is
// installed, use this. Otherwise, you may need to use
// MakeTestParamsFromList.
//
// Sub-tests which are not eligible for being on the CQ (unstable or
// download tests) will be tagged informational. Whether the test as a
// whole is CQ-critical should be controlled by a test-level
// informational attribute.
func MakeTestParams(t genparams.TestingT) string {
	defaultTest := Param{}
	return MakeTestParamsFromList(t, []Param{defaultTest})
}
