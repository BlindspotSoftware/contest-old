// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package sleep

import (
	"encoding/json"
	"fmt"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
)

// Name is the name used to look this plugin up.
var Name = "Sleep"

const (
	parametersKeyword = "parameters"
)

// Events defines the events that a TestStep is allow to emit
var Events = []event.Name{}

type parameters struct {
	Duration xjson.Duration `json:"duration"`
}

// TestStep implementation for this teststep plugin
type TestStep struct {
	parameters
}

// Name returns the name of the Step
func (ts *TestStep) Name() string {
	return Name
}

// Run executes the step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) populateParams(stepParams test.TestStepParameters) error {
	var parameters *test.Param

	if parameters = stepParams.GetOne(parametersKeyword); parameters.IsEmpty() {
		return fmt.Errorf("parameters cannot be empty")
	}

	if err := json.Unmarshal(parameters.JSON(), &ts.parameters); err != nil {
		return fmt.Errorf("failed to deserialize parameters: %v", err)
	}

	if ts.Duration == 0 {
		return fmt.Errorf("sleep time cannot be zero")
	}

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *TestStep) ValidateParameters(ctx xcontext.Context, params test.TestStepParameters) error {
	return ts.populateParams(params)
}

// New initializes and returns a new EchoStep. It implements the TestStepFactory
// interface.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, events.Events
}
