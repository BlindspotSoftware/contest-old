package dutctl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
)

const (
	defaultTimeout    = time.Minute
	parametersKeyword = "parameters"
)

// Name is the name used to look this plugin up.
var Name = "DUTCtl"

type parameters struct {
	Host    string   `json:"host"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	UART    int      `json:"uart,omitempty"`
	Input   string   `json:"input,omitempty"`

	Expect []struct {
		Regex string `json:"regex,omitempty"`
	} `json:"expect,omitempty"`
}

// TestStep implementation for this teststep plugin
type TestStep struct {
	parameters
	options options.Parameters
}

// Name returns the plugin name.
func (ts TestStep) Name() string {
	return Name
}

// Run executes the Dutctl action.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters,
	ev testevent.Emitter, resumeState json.RawMessage,
) (json.RawMessage, error) {
	if err := ts.validateAndPopulate(params); err != nil {
		return nil, err
	}

	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

// Retrieve all the parameters defines through the jobDesc
func (ts *TestStep) validateAndPopulate(stepParams test.TestStepParameters) error {
	var parameters, optionsParams *test.Param

	if parameters = stepParams.GetOne(parametersKeyword); parameters.IsEmpty() {
		return fmt.Errorf("parameters cannot be empty")
	}

	if err := json.Unmarshal(parameters.JSON(), &ts.parameters); err != nil {
		return fmt.Errorf("failed to deserialize parameters: %v", err)
	}

	optionsParams = stepParams.GetOne(options.Keyword)

	if err := json.Unmarshal(optionsParams.JSON(), &ts.options); err != nil {
		return fmt.Errorf("failed to deserialize options: %v", err)
	}

	if ts.Host == "" {
		return fmt.Errorf("host must not be empty")
	}

	if ts.Command == "" {
		return fmt.Errorf("command must not be empty")
	}

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *TestStep) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return ts.validateAndPopulate(params)
}

// New initializes and returns a new awsDutctl test step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and evend which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, nil
}
