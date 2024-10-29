package cpustats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
	"github.com/linuxboot/contest/plugins/teststeps/cpu"
)

// We need a default timeout to avoid endless running tests.
const (
	defaultTimeout    = time.Minute
	parametersKeyword = "parameters"
)

type parameters struct {
	ToolPath string `json:"tool_path"`
	Interval string `json:"interval"`
	Expect   struct {
		General    []cpu.General    `json:"general"`
		Individual []cpu.Individual `json:"individual"`
	} `json:"expect"`
}

type Expect struct {
	General    []cpu.General    `json:"general"`
	Individual []cpu.Individual `json:"individual"`
}

type General struct {
	Option string `json:"option"`
	Value  string `json:"value"`
}

type Individual struct {
	Core   int    `json:"core"`
	Option string `json:"option"`
	Value  string `json:"value"`
}

// Name is the name used to look this plugin up.
var Name = "CPUStats"

// TestStep implementation for this teststep plugin
type TestStep struct {
	parameters
	transport transport.Parameters
	options   options.Parameters
}

// Run executes the cmd step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	// Validate the parameter
	if err := ts.validateAndPopulate(params); err != nil {
		return nil, err
	}

	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) validateAndPopulate(stepParams test.TestStepParameters) error {
	var parameters, transportParams, optionsParams *test.Param

	if parameters = stepParams.GetOne(parametersKeyword); parameters.IsEmpty() {
		return fmt.Errorf("parameters cannot be empty")
	}

	if err := json.Unmarshal(parameters.JSON(), &ts.parameters); err != nil {
		return fmt.Errorf("failed to deserialize parameters: %v", err)
	}

	if transportParams = stepParams.GetOne(transport.Keyword); transportParams.IsEmpty() {
		return fmt.Errorf("transport cannot be empty")
	}

	if err := json.Unmarshal(transportParams.JSON(), &ts.transport); err != nil {
		return fmt.Errorf("failed to deserialize transport: %v", err)
	}

	optionsParams = stepParams.GetOne(options.Keyword)

	if !optionsParams.IsEmpty() {
		if err := json.Unmarshal(optionsParams.JSON(), &ts.options); err != nil {
			return fmt.Errorf("failed to deserialize options: %v", err)
		}
	}

	if ts.ToolPath == "" {
		return fmt.Errorf("missing or empty 'tool_path' parameter")
	}

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *TestStep) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return ts.validateAndPopulate(params)
}

// New initializes and returns a new HWaaS test step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, events.Events
}

// Name returns the plugin name.
func (ts TestStep) Name() string {
	return Name
}
