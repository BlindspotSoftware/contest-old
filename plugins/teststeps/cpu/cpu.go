package cpu

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
)

const (
	input  = "input"
	expect = "expect"
)

type inputStepParams struct {
	Command string

	Transport struct {
		Proto   string          `json:"proto"`
		Options json.RawMessage `json:"options,omitempty"`
	} `json:"transport,omitempty"`

	Options struct {
		Timeout xjson.Duration `json:"timeout,omitempty"`
	}
}

type expectStepParams struct{}

// Name is the name used to look this plugin up.
var Name = "CPU"

// TestStep implementation for the exec plugin
type TestStep struct {
	inputStepParams
	expectStepParams
}

// Run executes the step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	if err := ts.populateParams(params); err != nil {
		return nil, err
	}

	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) populateParams(stepParams test.TestStepParameters) error {
	input := stepParams.GetOne(input).JSON()

	if err := json.Unmarshal(input, &ts.inputStepParams); err != nil {
		return fmt.Errorf("failed to deserialize %q parameters", input)
	}

	expect := stepParams.GetOne(expect)
	if cmp.Equal(expect, &test.Param{}) {
		return nil
	}

	if err := json.Unmarshal(expect.JSON(), &ts.expectStepParams); err != nil {
		return fmt.Errorf("failed to deserialize %q parameters", expect)
	}

	return nil
}

// ValidateParameters validates the parameters associated to the step
func (ts *TestStep) ValidateParameters(_ xcontext.Context, stepParams test.TestStepParameters) error {
	return ts.populateParams(stepParams)
}

// New initializes and returns a new exec step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, Events
}

// Name returns the name of the Step
func (ts TestStep) Name() string {
	return Name
}
